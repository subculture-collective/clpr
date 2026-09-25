import { describe, expect, it } from 'vitest';
import { docContentUrl, publishedDocPaths, resolveDocLink } from './docs-links';

const published = new Set([
    'compliance/twitch-embeds',
    'compliance/twitch-api-usage',
    'users/guide',
]);

describe('docContentUrl', () => {
    it('keeps directory separators and encodes each segment', () => {
        expect(docContentUrl('compliance/twitch-embeds')).toBe(
            '/api/v1/docs/content/compliance/twitch-embeds',
        );
        expect(docContentUrl('guides/getting started')).toBe(
            '/api/v1/docs/content/guides/getting%20started',
        );
        expect(docContentUrl('a/?b#c')).toBe('/api/v1/docs/content/a/%3Fb%23c');
    });
});

describe('resolveDocLink', () => {
    it('resolves sibling Markdown links against the current directory', () => {
        expect(
            resolveDocLink('compliance/twitch-embeds', 'twitch-api-usage.md#rate-limits', published),
        ).toEqual({ kind: 'document', path: 'compliance/twitch-api-usage', hash: 'rate-limits' });
        expect(resolveDocLink('compliance/twitch-embeds', './twitch-api-usage.md', published)).toEqual({
            kind: 'document',
            path: 'compliance/twitch-api-usage',
            hash: '',
        });
    });

    it('resolves parent-relative links and root-relative wikilinks', () => {
        expect(resolveDocLink('compliance/twitch-embeds', '../users/guide.md', published)).toEqual({
            kind: 'document',
            path: 'users/guide',
            hash: '',
        });
        expect(resolveDocLink('compliance/twitch-embeds', 'users/guide', published)).toEqual({
            kind: 'document',
            path: 'users/guide',
            hash: '',
        });
    });

    it('does not link to unpublished documents or repository files', () => {
        expect(resolveDocLink('compliance/twitch-embeds', 'guardrails.md', published)).toEqual({
            kind: 'unpublished',
        });
        expect(resolveDocLink('compliance/twitch-embeds', '../operations/runbook.md', published)).toEqual({
            kind: 'unpublished',
        });
        expect(resolveDocLink('compliance/twitch-embeds', '../../../etc/passwd', published)).toEqual({
            kind: 'unpublished',
        });
        expect(resolveDocLink('compliance/twitch-embeds', './openapi.yaml', published)).toEqual({
            kind: 'unpublished',
        });
    });

    it('classifies anchors, external URLs, and site paths', () => {
        expect(resolveDocLink('users/guide', '#voting', published)).toEqual({ kind: 'anchor', hash: 'voting' });
        expect(resolveDocLink('users/guide', 'https://dev.twitch.tv/docs/embed/', published)).toEqual({
            kind: 'external',
            href: 'https://dev.twitch.tv/docs/embed/',
        });
        expect(resolveDocLink('users/guide', 'mailto:support@clpr.tv', published).kind).toBe('external');
        expect(resolveDocLink('users/guide', '/privacy', published)).toEqual({ kind: 'site', href: '/privacy' });
    });
});

describe('publishedDocPaths', () => {
    it('collects file paths from nested directories', () => {
        expect(
            publishedDocPaths([
                { path: 'index', type: 'file' },
                {
                    path: 'compliance',
                    type: 'directory',
                    children: [{ path: 'compliance/twitch-embeds', type: 'file' }],
                },
            ]),
        ).toEqual(new Set(['index', 'compliance/twitch-embeds']));
    });
});
