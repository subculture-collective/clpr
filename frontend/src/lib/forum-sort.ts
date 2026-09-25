import type { ForumBackendSort, ForumSort } from '@/types/forum';

const BACKEND_SORT: Record<ForumSort, ForumBackendSort> = {
  newest: 'recent',
  'most-replied': 'replies',
  popular: 'popular',
};

/**
 * Reads a thread sort from a URL. Older links used `trending` and `hot`,
 * which both meant view-ranked threads; unknown values fall back to newest.
 */
export function parseForumSort(value: string | null | undefined): ForumSort {
  switch (value) {
    case 'most-replied':
    case 'replies':
      return 'most-replied';
    case 'popular':
    case 'trending':
    case 'hot':
      return 'popular';
    default:
      return 'newest';
  }
}

/** The backend's canonical sort name for a UI sort. */
export function toBackendForumSort(sort: ForumSort): ForumBackendSort {
  return BACKEND_SORT[sort] ?? 'recent';
}
