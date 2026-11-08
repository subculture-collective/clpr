import { expect, test, type Page } from '@playwright/test';

const user = (role: 'user' | 'admin') => ({
    id: '00000000-0000-4000-8000-000000000001',
    twitch_id: '121059319',
    username: 'moonmoon',
    display_name: 'MOONMOON',
    role,
    karma_points: 0,
    is_banned: false,
    created_at: '2026-01-01T00:00:00Z',
});

const controls = { broadcaster_id: '121059319', excluded: false, featured: false, clip_download_authorized: false, consent_text_version: '2026-11-07' };

async function mockApi(page: Page, role: 'user' | 'admin', handle: (method: string, path: string, body: unknown) => unknown) {
    // The app probes the session only after a sign-in left this hint.
    await page.addInitScript(() => localStorage.setItem('auth_session_hint', '1'));
    await page.route('**/api/v1/**', async route => {
        const request = route.request();
        const path = new URL(request.url()).pathname.replace(/^.*\/api\/v1/, '');
        if (path === '/auth/me') {
            return route.fulfill({ json: user(role) });
        }
        const json = handle(request.method(), path, request.postDataJSON());
        return route.fulfill({ json: json ?? { success: true, data: [] } });
    });
}

test('a broadcaster leaves their clips out and asks to be featured', async ({ page }) => {
    let state = { ...controls };
    const puts: unknown[] = [];
    await mockApi(page, 'user', (method, path, body) => {
        if (path === '/users/me/countdowns' && method === 'GET') return { success: true, data: state };
        if (path === '/users/me/countdowns/exclusion' && method === 'PUT') {
            puts.push(body);
            state = { ...state, excluded: (body as { excluded: boolean }).excluded, exclusion_source: 'self_service' } as typeof state;
            return { success: true, data: state };
        }
        return undefined;
    });
    await page.goto('/settings');

    const card = page.getByTestId('countdown-settings');
    const toggle = card.getByRole('switch', { name: /Leave my clips out/ });
    await expect(toggle).not.toBeChecked();
    await card.getByText('Leave my clips out', { exact: true }).click();
    await expect(toggle).toBeChecked();
    expect(puts).toEqual([{ excluded: true }]);

    await expect(card.getByRole('link', { name: 'Feature my clips' })).toHaveAttribute(
        'href',
        /\/api\/v1\/twitch\/oauth\/authorize\?purpose=feature&return_to=%2Fsettings$/,
    );
});

test('an admin excludes a broadcaster without an account and clears a removal task', async ({ page }) => {
    let removals = [{
        render_id: '11111111-1111-4111-8111-111111111111', kind: 'short', youtube_video_id: 'abcDEF12345', edition_date: '2026-11-07',
        series_slug: 'overall', clip_id: '22222222-2222-4222-8222-222222222222', clip_title: 'Clutch ace', broadcaster_name: 'alpha',
        exclusion_reason: 'request', exclusion_source: 'admin', excluded_at: '2026-11-07T10:00:00Z', due_at: '2099-11-14T10:00:00Z',
    }];
    let exclusions: Record<string, unknown>[] = [];
    await mockApi(page, 'admin', (method, path, body) => {
        if (path === '/admin/countdowns/removals' && method === 'GET') return { success: true, data: removals };
        if (path === '/admin/countdowns/removals' && method === 'POST') {
            removals = [];
            return { success: true };
        }
        if (path === '/admin/countdowns/exclusions' && method === 'GET') return { success: true, data: exclusions };
        if (path === '/admin/countdowns/exclusions' && method === 'POST') {
            exclusions = [{ id: 'x1', ...(body as object), source: 'admin', created_at: '2026-11-07T12:00:00Z' }];
            return { success: true, data: exclusions[0] };
        }
        return undefined;
    });
    await page.goto('/admin/countdowns');

    await expect(page.getByRole('heading', { level: 1, name: 'Countdowns' })).toBeVisible();
    const task = page.getByTestId('removal-task');
    await expect(task).toContainText('Clutch ace');
    await task.getByRole('button', { name: 'Mark done' }).click();
    await expect(page.getByText('No uploaded video shows an excluded clip.')).toBeVisible();

    await page.getByLabel('Twitch broadcaster ID').fill('987654');
    await page.getByRole('button', { name: 'Exclude' }).click();
    await expect(page.getByTestId('exclusion')).toContainText('Broadcaster 987654');
});
