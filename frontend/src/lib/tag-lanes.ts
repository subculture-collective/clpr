import type { Tag, TagEvidence, TagLane } from '@/types/tag';

/**
 * Tag lanes mirror backend/internal/tagtaxonomy: every tag is shown by the
 * process that produced it rather than by a per-tag colour.
 */
export const TAG_LANES: Record<TagLane, { label: string; description: string; order: number }> = {
    detected: {
        label: 'What clpr saw',
        description: 'Chosen from clpr’s content vocabulary by the vision tagger or title rules.',
        order: 0,
    },
    category: {
        label: 'Twitch category',
        description: 'The category the clip was streamed under, and its genre.',
        order: 1,
    },
    community: {
        label: 'Community',
        description: 'Added by people on clpr, plus older tags from before tags had namespaces.',
        order: 2,
    },
    streamer: {
        label: 'Streamer tags',
        description: 'Tags the broadcaster put on their own Twitch channel. They describe the channel, not this clip.',
        order: 3,
    },
    duration: { label: 'Length', description: 'Clip length bucket.', order: 4 },
    language: { label: 'Language', description: 'Clip language.', order: 5 },
    root: { label: 'Namespace', description: 'A tag namespace.', order: 6 },
};

export const TAG_EVIDENCE: Record<TagEvidence, { short: string; label: string; description: string }> = {
    visible: {
        short: 'Seen',
        label: 'Seen',
        description: 'Something visible in a single frame.',
    },
    contextual: {
        short: 'Ctx',
        label: 'Context',
        description: 'Needs Twitch metadata or a transcript as well as the picture.',
    },
    strong: {
        short: 'Outcome',
        label: 'Outcome',
        description: 'Describes a result or judgement, which needs evidence beyond a title.',
    },
};

const STORED_PREFIX = /^(?:taxonomy|content|game|language|lang|duration|community|streamer):\s*/i;

/** Lane for a tag, falling back to its slug namespace for older responses. */
export function tagLane(tag: Pick<Tag, 'slug' | 'lane'>): TagLane {
    if (tag.lane) return tag.lane;
    const [namespace, rest] = (tag.slug ?? '').split('/', 2);
    if (rest === undefined) return 'community';
    switch (namespace) {
        case 'content':
            return 'detected';
        case 'streamer':
            return 'streamer';
        case 'game':
            return 'category';
        case 'duration':
            return 'duration';
        case 'lang':
            return 'language';
        default:
            return 'community';
    }
}

/** Human label without stored namespace prefixes such as "Content: ". */
export function tagLabel(tag: Pick<Tag, 'name' | 'display_name'>): string {
    if (tag.display_name) return tag.display_name;
    let name = (tag.name ?? '').trim();
    while (STORED_PREFIX.test(name)) name = name.replace(STORED_PREFIX, '');
    return name || tag.name || '';
}

/** Route for a tag. Namespaced slugs travel as one encoded segment. */
export function tagHref(slug: string): string {
    return `/tags/${encodeURIComponent(slug)}`;
}

/** Lanes a clip shows as chips; length and language render as metadata. */
export function isChipLane(lane: TagLane): boolean {
    return lane === 'detected' || lane === 'category' || lane === 'community' || lane === 'streamer';
}

/** Orders tags by lane, then by usage, keeping the input order stable. */
export function sortByLane<T extends Pick<Tag, 'slug' | 'lane' | 'usage_count'>>(tags: T[]): T[] {
    return tags
        .map((tag, index) => ({ tag, index }))
        .sort((a, b) =>
            TAG_LANES[tagLane(a.tag)].order - TAG_LANES[tagLane(b.tag)].order ||
            (b.tag.usage_count ?? 0) - (a.tag.usage_count ?? 0) ||
            a.index - b.index,
        )
        .map(({ tag }) => tag);
}
