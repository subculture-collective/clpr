import { describe, it, expect, vi, beforeEach, afterEach, beforeAll } from 'vitest';
import type { Mocked } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { DocsPage } from './DocsPage';
import axios from 'axios';
import type { DocFrontmatter, TOCEntry } from '../lib/markdown-utils';

// Mock the components
vi.mock('../components', () => ({
  Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Card: ({ children, id, className }: { children: React.ReactNode; id?: string; className?: string }) => <div id={id} className={className}>{children}</div>,
  CardBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SEO: ({ title }: { title: string }) => <div data-testid="seo">{title}</div>,
  DocHeader: ({ frontmatter }: { frontmatter: DocFrontmatter }) => <div data-testid="doc-header">{frontmatter?.title}</div>,
  DocTOC: ({ toc }: { toc: TOCEntry[] }) => <div data-testid="doc-toc">{toc.length} items</div>,
}));

vi.mock('axios');

const mockedAxios = axios as Mocked<typeof axios>;

const mockDocsResponse = {
  docs: [
    {
      name: 'compliance',
      path: 'compliance',
      type: 'directory' as const,
      children: [
        {
          name: 'README',
          path: 'compliance/README',
          type: 'file' as const,
          title: 'Twitch Compliance Overview',
        },
        {
          name: 'twitch-embeds',
          path: 'compliance/twitch-embeds',
          type: 'file' as const,
          title: 'Twitch Embed Compliance',
        },
      ],
    },
    {
      name: 'getting-started',
      path: 'getting-started',
      type: 'directory' as const,
      children: [
        {
          name: 'user-guide',
          path: 'getting-started/user-guide',
          type: 'file' as const,
        },
      ],
    },
  ],
};

const documents: Record<string, { path: string; title: string; content: string }> = {
  '/api/v1/docs/content/compliance/README': {
    path: 'compliance/README.md',
    title: 'Twitch Compliance Overview',
    content: [
      '# Twitch Compliance Overview',
      '',
      'See [embeds](twitch-embeds.md#parent-domain), [guardrails](guardrails.md),',
      'and the [privacy policy](/privacy).',
    ].join('\n'),
  },
  '/api/v1/docs/content/compliance/twitch-embeds': {
    path: 'compliance/twitch-embeds.md',
    title: 'Twitch Embed Compliance',
    content: '# Twitch Embed Compliance\nEmbeds use the official player.',
  },
  '/api/v1/docs/content/getting-started/user-guide': {
    path: 'getting-started/user-guide.md',
    title: 'Getting Started',
    content: '---\ntitle: Getting Started\n---\n# Getting Started\nSome content',
  },
};

beforeAll(() => {
  window.scrollTo = vi.fn();
});

beforeEach(() => {
  mockedAxios.get.mockReset();
  mockedAxios.get.mockImplementation((url: string) => {
    if (url === '/api/v1/docs') {
      return Promise.resolve({ data: mockDocsResponse });
    }
    if (url.startsWith('/api/v1/docs/search')) {
      return Promise.resolve({ data: { results: [] } });
    }
    if (documents[url]) {
      return Promise.resolve({ data: documents[url] });
    }
    return Promise.reject(new Error(`unexpected request ${url}`));
  });
});

afterEach(() => {
  mockedAxios.get.mockClear();
});

describe('DocsPage', () => {
  it('renders the documentation hub heading and doc tree once data loads', async () => {
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    expect(await screen.findByRole('heading', { name: /documentation hub/i })).toBeInTheDocument();
    expect(await screen.findByRole('button', { name: /user guide/i })).toBeInTheDocument();
  });

  it('labels each entry with its own title', async () => {
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    expect(await screen.findByRole('button', { name: 'Twitch Embed Compliance' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Twitch Compliance Overview' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'README' })).not.toBeInTheDocument();
  });

  it('opens a nested document through the wildcard content route', async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    const docButton = await screen.findByRole('button', { name: /user guide/i });
    await user.click(docButton);

    await waitFor(() => {
      expect(screen.getByTestId('doc-header')).toBeInTheDocument();
    });
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/v1/docs/content/getting-started/user-guide');
    expect(screen.getByTestId('seo')).toHaveTextContent('Getting Started');
  });

  it('follows relative links to published documents only', async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    await user.click(await screen.findByRole('button', { name: 'Twitch Compliance Overview' }));
    const embedsLink = await screen.findByRole('button', { name: 'embeds' });
    // Unpublished documents are plain text, not links that fail to load.
    expect(screen.getByText('guardrails').closest('a, button')).toBeNull();
    expect(screen.getByRole('link', { name: 'privacy policy' })).toHaveAttribute('href', '/privacy');

    await user.click(embedsLink);
    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalledWith('/api/v1/docs/content/compliance/twitch-embeds');
    });
    expect(await screen.findByText('Embeds use the official player.')).toBeInTheDocument();
  });

  it('links to the API reference and the canonical policy pages', async () => {
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    expect(await screen.findByRole('link', { name: /api reference/i })).toHaveAttribute(
      'href',
      '/openapi/api-reference.md',
    );
    expect(screen.getByRole('link', { name: 'Privacy Policy' })).toHaveAttribute('href', '/privacy');
    expect(screen.getByRole('link', { name: 'Terms of Service' })).toHaveAttribute('href', '/terms');
  });

  it('does not render repository resources', async () => {
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    expect(await screen.findByRole('heading', { name: /documentation hub/i })).toBeInTheDocument();
    expect(screen.queryByText('GitHub Repository')).not.toBeInTheDocument();
  });
});
