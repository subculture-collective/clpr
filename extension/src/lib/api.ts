/**
 * clpr API client for the browser extension.
 *
 * All requests include `Authorization: Bearer <token>` obtained from the
 * clpr access_token cookie.  The extension must declare host_permissions
 * for the API origin so Chrome relaxes CORS enforcement for extension contexts.
 */

import type {
  ClipMetadata,
  Tag,
  SubmitClipRequest,
  SubmissionResponse,
  ExtensionConfig,
} from '../types';

interface TagListResponse {
  success: boolean;
  data: Tag[];
  tags?: Tag[];
}

interface MetadataApiResponse {
  success: boolean;
  data: ClipMetadata;
}

/**
 * Fetches clip metadata from the clpr API (which proxies Twitch).
 */
export async function fetchClipMetadata(
  clipUrl: string,
  token: string,
  config: ExtensionConfig
): Promise<ClipMetadata> {
  const url = `${config.apiBaseUrl}/submissions/metadata?url=${encodeURIComponent(clipUrl)}`;
  const res = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch metadata: ${res.status}`);
  }
  const body = await res.json() as MetadataApiResponse;
  return body.data;
}

/**
 * Fetches the list of available tags from the clpr API.
 */
export async function fetchTags(
  token: string,
  config: ExtensionConfig
): Promise<Tag[]> {
  const res = await fetch(`${config.apiBaseUrl}/tags?limit=100`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return [];
  const body = await res.json() as TagListResponse;
  return body.data ?? body.tags ?? [];
}

/**
 * Builds an error message from a failed response. Accepts `{ error: string }`,
 * `{ error: { message } }` and `{ message }` JSON bodies; anything else
 * (HTML error pages, empty bodies) falls back to the HTTP status.
 */
async function readErrorMessage(res: Response, prefix: string): Promise<string> {
  const fallback = `${prefix}: ${res.status}${res.statusText ? ` ${res.statusText}` : ''}`;
  try {
    const body = (await res.json()) as {
      error?: string | { message?: string };
      message?: string;
    };
    if (typeof body?.error === 'string' && body.error) return body.error;
    if (typeof body?.error === 'object' && body.error?.message) return body.error.message;
    if (typeof body?.message === 'string' && body.message) return body.message;
  } catch {
    // Non-JSON error body.
  }
  return fallback;
}

/**
 * Submits a clip to clpr.
 */
export async function submitClip(
  request: SubmitClipRequest,
  token: string,
  config: ExtensionConfig
): Promise<SubmissionResponse> {
  const res = await fetch(`${config.apiBaseUrl}/submissions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    throw new Error(await readErrorMessage(res, 'Submission failed'));
  }
  return await res.json() as SubmissionResponse;
}
