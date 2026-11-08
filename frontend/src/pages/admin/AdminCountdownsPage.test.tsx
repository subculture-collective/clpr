import { http, HttpResponse } from 'msw';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { server } from '../../test/mocks/server';
import { render, screen, within } from '../../test/test-utils';
import { AdminCountdownsPage } from './AdminCountdownsPage';

vi.mock('../../components/SEO', () => ({ SEO: () => null }));

const API = '*/api/v1';
const task = {
  render_id: '11111111-1111-4111-8111-111111111111',
  kind: 'countdown',
  youtube_video_id: 'abcDEF12345',
  edition_date: '2026-11-07',
  series_slug: 'overall',
  clip_id: '22222222-2222-4222-8222-222222222222',
  clip_title: 'Clutch ace',
  broadcaster_name: 'alpha',
  exclusion_reason: 'request',
  exclusion_source: 'self_service',
  excluded_at: '2026-11-07T10:00:00Z',
  due_at: '2099-11-14T10:00:00Z',
};

describe('AdminCountdownsPage', () => {
  it('lists removal tasks and resolves one', async () => {
    let tasks = [task];
    const resolved: unknown[] = [];
    server.use(
      http.get(`${API}/admin/countdowns/removals`, () => HttpResponse.json({ success: true, data: tasks })),
      http.get(`${API}/admin/countdowns/exclusions`, () => HttpResponse.json({ success: true, data: [] })),
      http.post(`${API}/admin/countdowns/removals`, async ({ request }) => {
        resolved.push(await request.json());
        tasks = [];
        return HttpResponse.json({ success: true });
      }),
    );
    render(<AdminCountdownsPage />);
    const row = await screen.findByTestId('removal-task');
    expect(within(row).getByText(/Clutch ace/)).toBeInTheDocument();
    expect(within(row).getByRole('link', { name: 'open in YouTube Studio' })).toHaveAttribute(
      'href', 'https://studio.youtube.com/video/abcDEF12345/edit');

    // Keeping a video needs a note.
    await userEvent.selectOptions(within(row).getByLabelText('Resolution'), 'kept');
    expect(within(row).getByRole('button', { name: 'Mark done' })).toBeDisabled();
    await userEvent.selectOptions(within(row).getByLabelText('Resolution'), 'edited');
    await userEvent.click(within(row).getByRole('button', { name: 'Mark done' }));

    expect(await screen.findByText('No uploaded video shows an excluded clip.')).toBeInTheDocument();
    expect(resolved).toEqual([{ render_id: task.render_id, clip_id: task.clip_id, resolution: 'edited' }]);
  });

  it('excludes a broadcaster who has no account and lifts an exclusion', async () => {
    let exclusions: Record<string, unknown>[] = [];
    const created: unknown[] = [];
    server.use(
      http.get(`${API}/admin/countdowns/removals`, () => HttpResponse.json({ success: true, data: [] })),
      http.get(`${API}/admin/countdowns/exclusions`, () => HttpResponse.json({ success: true, data: exclusions })),
      http.post(`${API}/admin/countdowns/exclusions`, async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        created.push(body);
        exclusions = [{ id: 'x1', ...body, source: 'admin', created_at: '2026-11-07T12:00:00Z' }];
        return HttpResponse.json({ success: true, data: exclusions[0] }, { status: 201 });
      }),
      http.delete(`${API}/admin/countdowns/exclusions/x1`, () => {
        exclusions = [];
        return HttpResponse.json({ success: true, data: { id: 'x1' } });
      }),
    );
    render(<AdminCountdownsPage />);
    await screen.findByText('Nothing is excluded.');
    await userEvent.type(screen.getByLabelText('Twitch broadcaster ID'), '121059319');
    await userEvent.type(screen.getByLabelText('Exclusion note'), 'Asked by email');
    await userEvent.click(screen.getByRole('button', { name: 'Exclude' }));

    const row = await screen.findByTestId('exclusion');
    expect(within(row).getByText('Broadcaster 121059319')).toBeInTheDocument();
    expect(created).toEqual([{ broadcaster_id: '121059319', reason: 'request', note: 'Asked by email' }]);

    await userEvent.click(within(row).getByRole('button', { name: 'Lift' }));
    expect(await screen.findByText('Nothing is excluded.')).toBeInTheDocument();
  });
});
