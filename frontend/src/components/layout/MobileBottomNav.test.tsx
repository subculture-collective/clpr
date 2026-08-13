import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MobileBottomNav } from './MobileBottomNav';

const authState = { isAuthenticated: false };

vi.mock('@/context/AuthContext', () => ({
  useAuth: () => authState,
}));

describe('MobileBottomNav', () => {
  beforeEach(() => {
    authState.isAuthenticated = false;
  });

  it('keeps stable destinations and sends protected guest actions to login', () => {
    render(<MemoryRouter><MobileBottomNav /></MemoryRouter>);

    expect(screen.getByRole('link', { name: 'Feed' })).toHaveAttribute('href', '/');
    expect(screen.getByRole('link', { name: 'Discover' })).toHaveAttribute('href', '/discover');
    expect(screen.getByRole('link', { name: 'Saved' })).toHaveAttribute('href', '/login');
    expect(screen.getByRole('link', { name: 'Profile' })).toHaveAttribute('href', '/login');
  });

  it('links signed-in viewers directly to saved clips and their profile', () => {
    authState.isAuthenticated = true;
    render(<MemoryRouter><MobileBottomNav /></MemoryRouter>);

    expect(screen.getByRole('link', { name: 'Saved' })).toHaveAttribute('href', '/favorites');
    expect(screen.getByRole('link', { name: 'Profile' })).toHaveAttribute('href', '/profile');
  });
});
