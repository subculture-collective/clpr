import { randomUUID } from 'node:crypto';
import { expect, test, type BrowserContext } from '@playwright/test';

const api = process.env.PLAYWRIGHT_API_BASE_URL || 'http://127.0.0.1:18088';

test.beforeEach(async ({ page }) => {
    page.on('pageerror', error => console.error('Authenticated journey browser error:', error.message));
});

async function login(context: BrowserContext, browser: string, role: string) {
    await context.clearCookies();
    const response = await context.request.post(`${api}/api/v1/auth/test-login`, {
        data: { username: `clpr-e2e-${browser}-${role}` },
    });
    expect(response.status()).toBe(200);
    // A browser sign-in records this hint; the app skips the session probe without it.
    await context.addInitScript(() => localStorage.setItem('auth_session_hint', '1'));
    return (await response.json()).user as { id: string };
}

async function mutate(context: BrowserContext, path: string, method: string, data?: unknown) {
    await context.request.get(`${api}/api/v1/auth/me`);
    const csrf = (await context.cookies(api)).find(cookie => cookie.name === 'csrf_token')?.value;
    expect(csrf).toBeTruthy();
    return context.request.fetch(`${api}/api/v1${path}`, {
        method, data, headers: { 'X-CSRF-Token': decodeURIComponent(csrf!) },
    });
}

test('profile and privacy survive a fresh session; logout and invalid sessions lose access', async ({ page, context, browserName }) => {
    await login(context, browserName, 'member');
    await page.goto('/settings');
    const display = `Candidate ${browserName}`;
    await page.getByLabel('Display Name', { exact: true }).fill(display);
    await page.getByRole('button', { name: 'Save Profile', exact: true }).click();
    await expect(page.getByText('Profile updated successfully!')).toBeVisible();
    await page.getByLabel('Profile Visibility', { exact: true }).selectOption('private');
    await page.getByRole('button', { name: 'Save Settings', exact: true }).click();
    await expect(page.getByText('Settings updated successfully!')).toBeVisible();
    const functional = page.getByLabel('Functional Cookies', { exact: true });
    const consentValue = !(await functional.isChecked());
    const consentSaved = page.waitForResponse(response => response.url().endsWith('/api/v1/users/me/consent') && response.request().method() === 'POST');
    await functional.focus();
    await functional.press('Space');
    expect((await consentSaved).status()).toBe(200);
    const persistedConsent = await context.request.get(`${api}/api/v1/users/me/consent`);
    expect(persistedConsent.status()).toBe(200);
    expect((await persistedConsent.json()).data.functional).toBe(consentValue);
    expect((await mutate(context, '/auth/logout', 'POST')).status()).toBe(200);
    expect((await context.request.get(`${api}/api/v1/auth/me`)).status()).toBe(401);
    await login(context, browserName, 'member');
    await page.reload();
    await expect(page.getByLabel('Display Name', { exact: true })).toHaveValue(display);
    await expect(page.getByLabel('Profile Visibility', { exact: true })).toHaveValue('private');
    await expect(page.getByLabel('Functional Cookies', { exact: true })).toBeChecked({ checked: consentValue });
    await context.clearCookies();
    await context.addCookies([{ name: 'access_token', value: 'expired-invalid-session', url: api }]);
    await page.reload();
    await expect(page.getByLabel('Display Name', { exact: true })).toHaveCount(0);
    expect((await context.request.get(`${api}/api/v1/users/me/settings`)).status()).toBe(401);
});

test('private playlist persists edits and excludes another member; privileged operations enforce scope', async ({ page, context, browser, browserName }) => {
    await login(context, browserName, 'member');
    await page.goto('/playlists');
    const title = `Private ${randomUUID()}`;
    await page.getByRole('button', { name: 'Create Playlist', exact: true }).click();
    await page.getByLabel('Title', { exact: false }).fill(title);
    await page.getByLabel('Visibility', { exact: true }).selectOption('private');
    const created = page.waitForResponse(response => response.url().endsWith('/api/v1/playlists') && response.request().method() === 'POST');
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    const response = await created;
    expect(response.status(), await response.text()).toBe(201);
    const playlist = (await response.json()).data as { id: string };
    await page.getByRole('button', { name: `Edit ${title}`, exact: true }).click();
    await page.getByLabel('Title', { exact: false }).fill(`${title} edited`);
    await page.getByRole('button', { name: 'Update', exact: true }).click();
    await expect(page.getByText(`${title} edited`, { exact: true })).toBeVisible();
    await page.reload();
    await expect(page.getByText(`${title} edited`, { exact: true })).toBeVisible();
    // Back navigation must not recover private data from the preceding session.
    await page.goto(`/playlists/${playlist.id}`);
    await expect(page.getByRole('heading', { name: `${title} edited`, exact: true })).toBeVisible();
    const rejectCookies = page.getByRole('button', { name: 'Reject All', exact: true });
    if (await rejectCookies.isVisible()) await rejectCookies.click();
    await page.getByRole('button', { name: 'User menu', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Logout', exact: true }).click();
    await expect(page).toHaveURL(/\/$/);
    expect((await context.request.get(`${api}/api/v1/playlists/${playlist.id}`)).status()).toBe(404);
    await page.goBack({ waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: `${title} edited`, exact: true })).toHaveCount(0);
    await expect(page.getByText('Playlist not found', { exact: true })).toBeVisible();
    await login(context, browserName, 'member');
    await page.reload({ waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: `${title} edited`, exact: true })).toBeVisible();
    const other = await browser.newContext();
    try {
        const second = await login(other, browserName, 'second');
        expect((await other.request.get(`${api}/api/v1/playlists/${playlist.id}`)).status()).toBe(404);
        expect((await mutate(other, `/playlists/${playlist.id}`, 'PATCH', { title: 'Unauthorized edit' })).status()).toBe(403);
        expect((await other.request.get(`${api}/api/v1/admin/users`)).status()).toBe(403);
        expect((await mutate(other, '/admin/submissions/00000000-0000-4000-8000-000000000001/approve', 'POST')).status()).toBe(403);
        await login(other, browserName, 'scoped');
        const ban = await mutate(other, '/moderation/ban', 'POST', {
            channelId: '00000000-0000-4000-8000-000000009902', userId: second.id, reason: 'scope contract',
        });
        expect(ban.status()).toBe(403);
        await login(other, browserName, 'admin');
        expect((await other.request.get(`${api}/api/v1/admin/users`)).status()).toBe(200);
    } finally {
        await other.close();
        expect((await mutate(context, `/playlists/${playlist.id}`, 'DELETE')).status()).toBe(200);
    }
});

test('member submits clips and sees persisted moderator approval and rejection', async ({ page, context, browser, browserName }) => {
    await login(context, browserName, 'member');
    const moderator = await browser.newContext();
    try {
        await login(moderator, browserName, 'moderator');
        const reviewPage = await moderator.newPage();
        for (const outcome of ['approve', 'reject']) {
            await login(context, browserName, outcome === 'approve' ? 'member' : 'second');
            const slug = `clpr-e2e-${randomUUID()}`;
            await page.goto('/submit');
            await page.getByLabel('Twitch Clip URL').fill(`https://clips.twitch.tv/${slug}`);
            await page.getByLabel('Twitch Clip URL').blur();
            const button = page.getByRole('button', { name: 'Submit Clip', exact: true });
            await expect(button).toBeEnabled();
            const [response] = await Promise.all([
                page.waitForResponse(response => response.url().endsWith('/api/v1/submissions') && response.request().method() === 'POST'),
                button.click(),
            ]);
            expect(response.status()).toBe(201);
            await expect(page.getByRole('heading', { name: 'Submission Successful!' })).toBeVisible();
            // Verify persistence independently of the browser's response-body
            // decoder, which differs for service-worker responses in Firefox.
            const saved = await context.request.get(`${api}/api/v1/submissions`);
            expect(saved.status()).toBe(200);
            const submission = (await saved.json()).data.find(
                (record: { twitch_clip_id: string }) => record.twitch_clip_id === slug,
            ) as { id: string; status: string };
            expect(submission).toBeDefined();
            expect(submission.status).toBe('pending');
            const pendingResponse = reviewPage.waitForResponse(response => new URL(response.url()).pathname === '/api/v1/admin/submissions');
            await reviewPage.goto(new URL('/admin/submissions', page.url()).href);
            expect((await pendingResponse).status()).toBe(200);
            await expect(reviewPage.getByRole('button', { name: 'Refresh', exact: true })).toBeEnabled();
            const rejectCookies = reviewPage.getByRole('button', { name: 'Reject All', exact: true });
            if (await rejectCookies.isVisible()) await rejectCookies.click();
            const pending = reviewPage.getByRole('article', { name: `Candidate submission ${slug}`, exact: true });
            while (!(await pending.isVisible())) {
                const next = reviewPage.getByRole('button', { name: 'Next', exact: true });
                await expect(next).toBeEnabled();
                const nextPage = reviewPage.waitForResponse(response => new URL(response.url()).pathname === '/api/v1/admin/submissions');
                await next.click();
                expect((await nextPage).status()).toBe(200);
                await expect(reviewPage.getByRole('button', { name: 'Refresh', exact: true })).toBeEnabled();
            }
            const reviewed = reviewPage.waitForResponse(response => response.url().endsWith(`/admin/submissions/${submission.id}/${outcome}`));
            await pending.getByRole('button', { name: outcome === 'approve' ? 'Approve' : 'Reject', exact: true }).click();
            if (outcome === 'reject') {
                await reviewPage.getByLabel('Rejection Reason', { exact: true }).fill('Candidate rejection contract');
                await reviewPage.getByRole('button', { name: 'Reject Submission', exact: true }).click();
            }
            expect((await reviewed).status()).toBe(200);
            await page.goto('/submissions');
            const persisted = await context.request.get(`${api}/api/v1/submissions`);
            expect(persisted.status()).toBe(200);
            const payload = await persisted.json();
            expect(payload.data.find((record: { id: string }) => record.id === submission.id))
                .toMatchObject({ id: submission.id, status: outcome === 'approve' ? 'approved' : 'rejected' });
            await expect(page.getByText(`Candidate submission ${slug}`, { exact: true })).toBeVisible();
        }
    } finally { await moderator.close(); }
});
