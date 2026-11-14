import { getHttpStatus } from '../../lib/error-utils';
import type { BulkModerationResponse } from '../../lib/moderation-api';

/**
 * True when a keydown may trigger a queue shortcut: no Ctrl/Cmd/Alt held, and
 * the focus is not in a field where typing a letter means something else.
 * Checkboxes count as allowed so the shortcut works right after ticking a row.
 */
export function isQueueShortcutEvent(e: KeyboardEvent): boolean {
    if (e.ctrlKey || e.metaKey || e.altKey) {
        return false;
    }
    const target = e.target;
    if (target instanceof HTMLInputElement) {
        return target.type === 'checkbox';
    }
    if (target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement) {
        return false;
    }
    if (target instanceof HTMLElement && target.isContentEditable) {
        return false;
    }
    return true;
}

/** True when the API says the item is gone or was already reviewed. */
export function isStaleQueueError(error: unknown): boolean {
    const status = getHttpStatus(error);
    return status === 409 || status === 404;
}

/** Success text for a bulk action, including how many items were skipped. */
export function bulkResultMessage(
    result: BulkModerationResponse,
    noun: string,
    outcome: string
): string {
    const skipped = Array.isArray(result.skipped) ? result.skipped.length : 0;
    const message = `${result.processed} of ${result.total} ${noun} ${outcome}.`;
    return skipped > 0
        ? `${message} ${skipped} skipped because ${skipped === 1 ? 'it was' : 'they were'} no longer pending.`
        : message;
}

/** Only site paths and http(s) links are safe to render as a link. */
export function safeContentUrl(url: string | undefined): string | undefined {
    if (!url) {
        return undefined;
    }
    return /^(\/(?!\/)|https?:\/\/)/i.test(url) ? url : undefined;
}
