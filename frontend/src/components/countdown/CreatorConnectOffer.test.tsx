import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { CreatorConnectOffer } from './CreatorConnectOffer';

const auth = vi.hoisted(() => ({ user: null as null | { twitch_id?: string } }));
vi.mock('@/context/AuthContext', () => ({ useAuth: () => auth }));

function renderOffer() {
  render(
    <MemoryRouter>
      <CreatorConnectOffer />
    </MemoryRouter>,
  );
}

describe('CreatorConnectOffer', () => {
  it('sends signed-out visitors to sign in', () => {
    auth.user = null;
    renderOffer();
    expect(screen.getByRole('heading', { name: 'Get your clips into the daily countdown' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Sign in with Twitch' })).toHaveAttribute('href', '/login');
  });

  it('sends a Twitch account to its creator dashboard', () => {
    auth.user = { twitch_id: '1234' };
    renderOffer();
    expect(screen.getByRole('link', { name: 'Open my creator dashboard' })).toHaveAttribute('href', '/creator/1234/dashboard');
  });
});
