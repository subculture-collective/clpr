import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, within } from '../test/test-utils';
import { WebhookSubscriptionsPage } from './WebhookSubscriptionsPage';

const api = vi.hoisted(() => ({
  createWebhookSubscription: vi.fn(),
  deleteWebhookSubscription: vi.fn(),
  getWebhookDeliveries: vi.fn(),
  getSupportedWebhookEvents: vi.fn(),
  listWebhookSubscriptions: vi.fn(),
  regenerateWebhookSecret: vi.fn(),
  updateWebhookSubscription: vi.fn(),
}));
vi.mock('../lib/webhook-api', () => api);
vi.mock('@dr.pogodin/react-helmet', () => ({ Helmet: () => null }));

const discordSub = {
  id: 's1',
  user_id: 'u1',
  url: 'https://discord.com/api/webhooks/1/token',
  events: ['daily.top_clips'],
  is_active: true,
  created_at: '2026-11-07T00:00:00Z',
  updated_at: '2026-11-07T00:00:00Z',
  target_type: 'discord' as const,
};

describe('WebhookSubscriptionsPage', () => {
  beforeEach(() => {
    Object.values(api).forEach(fn => fn.mockReset());
    api.getSupportedWebhookEvents.mockResolvedValue({
      events: ['clip.submitted', 'clip.approved', 'countdown.published', 'daily.top_clips'],
      discordEvents: ['countdown.published', 'daily.top_clips'],
    });
  });

  it('creates a Discord subscription with only the events Discord can show', async () => {
    api.listWebhookSubscriptions.mockResolvedValue([]);
    api.createWebhookSubscription.mockResolvedValue({ subscription: discordSub, secret: 'unused' });
    const user = userEvent.setup();
    render(<WebhookSubscriptionsPage />);

    await user.click(await screen.findByRole('button', { name: 'Create Webhook' }));
    const dialog = await screen.findByRole('dialog');
    await within(dialog).findByLabelText('clip.submitted');
    await user.click(within(dialog).getByLabelText('clip.approved'));
    await user.click(within(dialog).getByRole('radio', { name: 'Discord channel' }));

    expect(within(dialog).queryByLabelText('clip.approved')).not.toBeInTheDocument();
    await user.type(within(dialog).getByLabelText(/Discord webhook URL/), discordSub.url);
    await user.click(within(dialog).getByLabelText('daily.top_clips'));
    await user.click(within(dialog).getByRole('button', { name: 'Create Webhook' }));

    await vi.waitFor(() => expect(api.createWebhookSubscription).toHaveBeenCalled());
    expect(api.createWebhookSubscription.mock.calls[0][0]).toEqual({
      url: discordSub.url,
      events: ['daily.top_clips'],
      description: '',
      target_type: 'discord',
    });
    await vi.waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(screen.queryByText('unused')).not.toBeInTheDocument();
  });

  it('shows why the server rejected a webhook URL', async () => {
    api.listWebhookSubscriptions.mockResolvedValue([]);
    api.createWebhookSubscription.mockRejectedValue(Object.assign(new Error('Request failed with status code 400'), {
      response: { status: 400, data: { error: 'a Discord webhook URL looks like https://discord.com/api/webhooks/<id>/<token>' } },
    }));
    const user = userEvent.setup();
    render(<WebhookSubscriptionsPage />);
    await user.click(await screen.findByRole('button', { name: 'Create Webhook' }));
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('radio', { name: 'Discord channel' }));
    await user.type(within(dialog).getByLabelText(/Discord webhook URL/), 'https://evil.example/api/webhooks/1/x');
    await user.click(await within(dialog).findByLabelText('daily.top_clips'));
    await user.click(within(dialog).getByRole('button', { name: 'Create Webhook' }));
    expect(await within(dialog).findByText(/a Discord webhook URL looks like/)).toBeInTheDocument();
    expect(within(dialog).queryByText(/status code/)).not.toBeInTheDocument();
  });

  it('marks Discord subscriptions and offers no secret for them', async () => {
    api.listWebhookSubscriptions.mockResolvedValue([discordSub]);
    render(<WebhookSubscriptionsPage />);

    expect(await screen.findByText('Discord')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Regenerate Secret' })).not.toBeInTheDocument();
  });
});
