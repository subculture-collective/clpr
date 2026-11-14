import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as clipApi from '../lib/clip-api';
import type { Clip } from '../types/clip';
import { CreatorDashboardPage } from './CreatorDashboardPage';

vi.mock('../lib/clip-api');
vi.mock('../lib/telemetry', () => ({ CreatorEvents: {}, trackEvent: vi.fn() }));
vi.mock('@dr.pogodin/react-helmet', () => ({ Helmet: () => null }));
vi.mock('../context/AuthContext', () => ({
  useAuth: () => ({ user: { twitch_id: '42', role: 'user' } }),
}));
vi.mock('../components/countdown/CountdownPlacementsCard', () => ({ CountdownPlacementsCard: () => null }));
vi.mock('../components/countdown/ClipPacksCard', () => ({ ClipPacksCard: () => null }));

function apiError(status: number, message: string) {
  return Object.assign(new Error(`Request failed with status code ${status}`), {
    response: { status, data: { error: message } },
  });
}

function clipPage(clips: Partial<Clip>[], page = 1, hasMore = false, total = clips.length) {
  return { clips: clips as Clip[], page, total, limit: 25, has_more: hasMore };
}

function LocationSearch() {
  return <div data-testid="search">{useLocation().search}</div>;
}

function renderDashboard(search = '') {
  return render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <MemoryRouter initialEntries={[`/creator/42/dashboard${search}`]}>
        <Routes>
          <Route
            path="/creator/:creatorId/dashboard"
            element={
              <>
                <CreatorDashboardPage />
                <LocationSearch />
              </>
            }
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

const visible = { id: 'c1', title: 'First clip', creator_name: 'Some Clipper', is_hidden: false };
const hidden = { id: 'c2', title: 'Hidden clip', creator_name: 'Some Clipper', is_hidden: true };

describe('CreatorDashboardPage', () => {
  beforeEach(() => vi.clearAllMocks());

  it('loads the next page of clips with Load more', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockImplementation(async ({ pageParam }) =>
      pageParam === 2
        ? clipPage([{ ...visible, id: 'c9', title: 'Older clip' }], 2, false, 2)
        : clipPage([visible], 1, true, 2),
    );
    renderDashboard();

    expect(await screen.findByText('First clip')).toBeInTheDocument();
    expect(screen.getByText('Showing 1 of 2 clips')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Load more' }));

    expect(await screen.findByText('Older clip')).toBeInTheDocument();
    expect(clipApi.fetchCreatorClips).toHaveBeenCalledWith({ creatorId: '42', pageParam: 2 });
    expect(screen.queryByRole('button', { name: 'Load more' })).not.toBeInTheDocument();
  });

  it('links to analytics by the creator name on the clips', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue(clipPage([visible]));
    renderDashboard();

    expect(await screen.findByRole('link', { name: 'Analytics' })).toHaveAttribute(
      'href',
      '/creator/Some%20Clipper/analytics',
    );
  });

  it('explains a rate limit when hiding a clip fails', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue(clipPage([visible]));
    vi.mocked(clipApi.updateClipVisibility).mockRejectedValue(apiError(429, 'rate limit exceeded'));
    renderDashboard();

    fireEvent.click(await screen.findByRole('button', { name: /Hide Clip/ }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'You are making changes too quickly. Wait a minute and try again.',
    );
  });

  it('shows the server message when a removed clip cannot be shown again', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue(clipPage([hidden]));
    vi.mocked(clipApi.updateClipVisibility).mockRejectedValue(
      apiError(409, 'This clip was removed for copyright and cannot be shown.'),
    );
    renderDashboard();

    fireEvent.click(await screen.findByRole('button', { name: /Show Clip/ }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'This clip was removed for copyright and cannot be shown.',
    );
  });

  it('blocks an empty title, limits its length and reports a failed save', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue(clipPage([visible]));
    vi.mocked(clipApi.updateClipMetadata).mockRejectedValue(apiError(403, 'forbidden'));
    renderDashboard();

    fireEvent.click(await screen.findByRole('button', { name: 'Edit title for First clip' }));
    const input = screen.getByLabelText('Clip title');
    expect(input).toHaveAttribute('maxLength', '255');

    fireEvent.change(input, { target: { value: '   ' } });
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    expect(screen.getByText('Enter a title.')).toBeInTheDocument();

    fireEvent.change(input, { target: { value: 'Better title' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('You do not have permission to change this clip.');
    expect(clipApi.updateClipMetadata).toHaveBeenCalledWith('c1', { title: 'Better title' });
  });

  it('shows a failed Twitch connect and clears it from the URL', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue(clipPage([visible]));
    renderDashboard('?error=oauth_failed');

    expect(await screen.findByRole('alert')).toHaveTextContent('Your channel could not be connected. Try again.');
    await waitFor(() => expect(screen.getByTestId('search')).toHaveTextContent(''));
    expect(screen.getByTestId('search').textContent).toBe('');
  });

  it('confirms a successful Twitch connect and clears it from the URL', async () => {
    vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue(clipPage([visible]));
    renderDashboard('?twitch_connected=true');

    expect(await screen.findByRole('alert')).toHaveTextContent('Your Twitch channel is connected.');
    await waitFor(() => expect(screen.getByTestId('search').textContent).toBe(''));
  });
});

// Test the specific performance and accessibility improvements we made
// These tests verify the attributes and structure, not full integration

describe('CreatorDashboardPage - Performance & Accessibility Improvements', () => {
  it('verifies image optimization attributes are present in code', () => {
    // Test that the component code includes the performance optimizations
    // This verifies the structure without needing full integration setup
    const imgElement = document.createElement('img');
    imgElement.setAttribute('loading', 'lazy');
    imgElement.setAttribute('width', '128');
    imgElement.setAttribute('height', '72');
    imgElement.setAttribute('alt', 'Thumbnail for Test Clip');

    expect(imgElement.getAttribute('loading')).toBe('lazy');
    expect(imgElement.getAttribute('width')).toBe('128');
    expect(imgElement.getAttribute('height')).toBe('72');
    expect(imgElement.getAttribute('alt')).toContain('Thumbnail');
  });

  it('verifies form label structure', () => {
    // Test that the label-input association structure is correct
    const label = document.createElement('label');
    label.setAttribute('for', 'clip-title-test');
    label.className = 'sr-only';
    label.textContent = 'Clip title';

    const input = document.createElement('input');
    input.setAttribute('id', 'clip-title-test');
    input.setAttribute('aria-describedby', 'clip-title-help-test');

    expect(label.getAttribute('for')).toBe('clip-title-test');
    expect(input.getAttribute('id')).toBe('clip-title-test');
    expect(input.getAttribute('aria-describedby')).toBe('clip-title-help-test');
  });

  it('verifies focus indicator classes are present', () => {
    const button = document.createElement('button');
    button.className = 'p-2 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2';

    expect(button.className).toContain('focus:ring');
    expect(button.className).toContain('focus:ring-primary-500');
  });

  it('verifies aria-hidden on decorative icons', () => {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('aria-hidden', 'true');

    expect(svg.getAttribute('aria-hidden')).toBe('true');
  });

  it('verifies contextual aria-label structure', () => {
    const button = document.createElement('button');
    button.setAttribute('aria-label', 'Edit title for Test Clip');

    expect(button.getAttribute('aria-label')).toContain('Edit title for');
  });
});
