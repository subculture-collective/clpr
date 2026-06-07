import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { CommunityRulesPage } from './CommunityRulesPage';

// Mock the components
vi.mock('../components', () => ({
  Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Card: ({ children, id }: { children: React.ReactNode; id?: string }) => <div id={id}>{children}</div>,
  CardBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

describe('CommunityRulesPage', () => {
  it('renders the page title', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText('Community Rules')).toBeInTheDocument();
  });

  it('displays the last updated date', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/Last updated:/i)).toBeInTheDocument();
  });

  it('renders all 7 main rules', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/1\. Be Respectful and Kind/i)).toBeInTheDocument();
    expect(screen.getByText(/2\. Share Authentic Content/i)).toBeInTheDocument();
    expect(screen.getByText(/3\. No Spam or Self-Promotion Abuse/i)).toBeInTheDocument();
    expect(screen.getByText(/4\. Keep Content Safe and Appropriate/i)).toBeInTheDocument();
    expect(screen.getByText(/5\. Respect Privacy and Personal Information/i)).toBeInTheDocument();
    expect(screen.getByText(/6\. Follow Platform and Legal Guidelines/i)).toBeInTheDocument();
    expect(screen.getByText(/7\. Report Issues and Help Moderate/i)).toBeInTheDocument();
  });

  it('has anchor IDs for navigation', () => {
    const { container } = render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(container.querySelector('#be-respectful')).toBeInTheDocument();
    expect(container.querySelector('#authentic-content')).toBeInTheDocument();
    expect(container.querySelector('#no-spam')).toBeInTheDocument();
    expect(container.querySelector('#safe-content')).toBeInTheDocument();
    expect(container.querySelector('#respect-privacy')).toBeInTheDocument();
    expect(container.querySelector('#platform-guidelines')).toBeInTheDocument();
    expect(container.querySelector('#report-issues')).toBeInTheDocument();
  });

  it('includes enforcement and consequences section', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText('Enforcement and Consequences')).toBeInTheDocument();
    expect(screen.getByText(/Warning:/i)).toBeInTheDocument();
    expect(screen.getByText(/Content Removal:/i)).toBeInTheDocument();
    expect(screen.getByText(/Temporary Suspension:/i)).toBeInTheDocument();
    expect(screen.getByText(/Permanent Ban:/i)).toBeInTheDocument();
  });

  it('mentions key values and expectations', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/No harassment, hate speech, or discriminatory language/i)).toBeInTheDocument();
    expect(screen.getByText(/Only submit clips from legitimate Twitch streams/i)).toBeInTheDocument();
    expect(screen.getByText(/No NSFW \(Not Safe For Work\) content/i)).toBeInTheDocument();
  });

  it('provides contact information', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/Questions or Concerns\?/i)).toBeInTheDocument();
    const githubLink = screen.getByText('GitHub repository');
    expect(githubLink).toHaveAttribute('href', 'https://git.subcult.tv/subculture-collective/clpr');
  });

  it('mentions the 80/20 self-promotion rule', () => {
    render(
      <MemoryRouter>
        <CommunityRulesPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/80\/20 rule/i)).toBeInTheDocument();
  });
});
