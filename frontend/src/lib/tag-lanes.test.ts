import { describe, expect, it } from 'vitest';
import { isChipLane, sortByLane, tagHref, tagLabel, tagLane } from './tag-lanes';

describe('tag lanes', () => {
    it('prefers the API lane and display name', () => {
        expect(tagLane({ slug: 'content/english', lane: 'streamer' })).toBe('streamer');
        expect(tagLabel({ name: 'Content: english', display_name: 'English' })).toBe('English');
    });

    it('falls back to the slug namespace and strips stored prefixes', () => {
        expect(tagLane({ slug: 'game/just-chatting' })).toBe('category');
        expect(tagLane({ slug: 'lang/en' })).toBe('language');
        expect(tagLane({ slug: 'goosebumps' })).toBe('community');
        expect(tagLabel({ name: 'Language: English' })).toBe('English');
    });

    it('encodes namespaced slugs as one path segment', () => {
        expect(tagHref('content/boss-fight')).toBe('/tags/content%2Fboss-fight');
    });

    it('orders clip chips by lane and leaves length and language out', () => {
        const tags = [
            { slug: 'streamer/english', lane: 'streamer' as const, usage_count: 9000 },
            { slug: 'duration/short', lane: 'duration' as const, usage_count: 5 },
            { slug: 'goosebumps', lane: 'community' as const, usage_count: 3 },
            { slug: 'game/music', lane: 'category' as const, usage_count: 50 },
            { slug: 'content/singing', lane: 'detected' as const, usage_count: 10 },
        ];
        expect(sortByLane(tags.filter(tag => isChipLane(tagLane(tag)))).map(tag => tag.slug)).toEqual([
            'content/singing',
            'game/music',
            'goosebumps',
            'streamer/english',
        ]);
    });
});
