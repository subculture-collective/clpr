import type { Tag, TagLane } from '@/types/tag';

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

/** Orders tags by usage without prioritizing their source. */
export function sortByPopularity<T extends Pick<Tag, 'usage_count'>>(tags: T[]): T[] {
    return [...tags].sort((a, b) => (b.usage_count ?? 0) - (a.usage_count ?? 0));
}

const TAG_ACCENTS = [
    'text-primary-300',
    'text-success-300',
    'text-warning-300',
    'text-info-300',
    'text-secondary-300',
    'text-error-300',
] as const;

/** A stable decorative accent, independent of the tag's source or evidence. */
export function tagAccent(label: string): string {
    let hash = 0;
    for (const char of label.toLowerCase()) {
        hash = (Math.imul(hash, 31) + char.charCodeAt(0)) >>> 0;
    }
    return TAG_ACCENTS[hash % TAG_ACCENTS.length];
}
