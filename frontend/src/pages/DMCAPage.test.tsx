import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { DMCAPage } from './DMCAPage';

vi.mock('../components', () => ({
  Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Card: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  CardBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SEO: () => null,
}));

describe('DMCAPage', () => {
  const renderPage = () => render(<MemoryRouter><DMCAPage /></MemoryRouter>);

  it('renders current notice and counter-notice guidance', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: 'DMCA Copyright Policy' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /send a takedown notice/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /counter-notices/i })).toBeInTheDocument();
    expect(screen.getByText(/10–14 business days/i)).toBeInTheDocument();
  });

  it('publishes the copyright contact without placeholder agent details', () => {
    const { container } = renderPage();
    expect(screen.getAllByRole('link', { name: /dmca@clpr.tv/i }).length).toBeGreaterThan(0);
    expect(container).not.toHaveTextContent(/\[Agent Name\]|\[Street Address\]|\[Phone Number\]|safe harbor protection/i);
  });

  it('includes all required notice statements', () => {
    renderPage();
    expect(screen.getByText(/good-faith belief the disputed use/i)).toBeInTheDocument();
    expect(screen.getByText(/under penalty of perjury, that the notice is accurate/i)).toBeInTheDocument();
    expect(screen.getAllByText(/physical or electronic signature/i)).toHaveLength(2);
  });
});
