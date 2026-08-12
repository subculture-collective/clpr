import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { AboutPage } from './AboutPage';

vi.mock('../components', () => ({
  Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Card: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  CardBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SEO: () => null,
}));

describe('AboutPage', () => {
  const renderPage = () => render(<MemoryRouter><AboutPage /></MemoryRouter>);

  it('presents clpr as creator-first live culture discovery', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /find the creators/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /more than gaming/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /made for discovery/i })).toBeInTheDocument();
  });

  it('shows the favorites feature once with one checkmark', () => {
    renderPage();
    const favorite = screen.getByText('Save favorites for later viewing').closest('li');
    expect(favorite).toHaveTextContent('✓Save favorites for later viewing');
    expect(favorite?.querySelectorAll('span')).toHaveLength(2);
  });

  it('does not advertise development or repository links', () => {
    const { container } = renderPage();
    expect(container).not.toHaveTextContent(/open source|technology stack|github/i);
  });

  it('links to community rules, contact, and Patreon', () => {
    renderPage();
    expect(screen.getByRole('link', { name: /community rules/i })).toHaveAttribute('href', '/community-rules');
    expect(screen.getByRole('link', { name: /contact us/i })).toHaveAttribute('href', '/contact');
    expect(screen.getByRole('link', { name: /patreon/i })).toHaveAttribute('href', 'https://patreon.com/subcult');
  });
});
