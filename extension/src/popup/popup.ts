/**
 * Popup script for the Clipper browser extension.
 *
 * State machine:
 *  loading → no-clip | login | form
 *  form → success (after successful submission)
 */

import { getAccessToken, verifyAuth, openLoginPage } from '../lib/auth';
import { fetchClipMetadata, fetchTags, submitClip } from '../lib/api';
import type {
  TwitchClipInfo,
  ClipMetadata,
  Tag,
  UserProfile,
  ExtensionConfig,
} from '../types';
import { DEFAULT_CONFIG } from '../types';

// ─── View helpers ─────────────────────────────────────────────────────────────

type ViewId = 'loading' | 'no-clip' | 'login' | 'form' | 'success';

function showView(id: ViewId): void {
  const views = document.querySelectorAll<HTMLElement>('.view');
  views.forEach(v => {
    if (v.id === `view-${id}`) {
      v.classList.remove('hidden');
    } else {
      v.classList.add('hidden');
    }
  });
}

function el<T extends HTMLElement>(id: string): T {
  const e = document.getElementById(id);
  if (!e) throw new Error(`Element #${id} not found`);
  return e as T;
}

// ─── State ────────────────────────────────────────────────────────────────────

let config: ExtensionConfig = { ...DEFAULT_CONFIG };
let currentToken: string | null = null;
let currentClipInfo: TwitchClipInfo | null = null;
let currentMetadata: ClipMetadata | null = null;
let allTags: Tag[] = [];
let selectedTagSlugs: string[] = [];

// Guard: ensure setupTagSearch registers its document-level click listener only once.
let tagSearchInitialized = false;

// ─── Tag UI ───────────────────────────────────────────────────────────────────

function renderSelectedTags(): void {
  const container = el('selected-tags');
  container.innerHTML = '';

  selectedTagSlugs.forEach(slug => {
    const tag = allTags.find(t => t.slug === slug);
    if (!tag) return;

    const chip = document.createElement('span');
    chip.className = 'tag-chip';
    chip.setAttribute('data-slug', slug);

    const label = document.createElement('span');
    label.textContent = tag.name;

    const removeBtn = document.createElement('button');
    removeBtn.className = 'tag-chip-remove';
    removeBtn.textContent = '×';
    removeBtn.setAttribute('aria-label', `Remove tag ${tag.name}`);
    removeBtn.addEventListener('click', () => {
      selectedTagSlugs = selectedTagSlugs.filter(s => s !== slug);
      renderSelectedTags();
    });

    chip.appendChild(label);
    chip.appendChild(removeBtn);
    container.appendChild(chip);
  });
}

function setupTagSearch(): void {
  const input = el<HTMLInputElement>('input-tag-search');
  const dropdown = el<HTMLUListElement>('tag-dropdown');

  function showDropdown(matches: Tag[]): void {
    dropdown.innerHTML = '';
    if (matches.length === 0) {
      dropdown.classList.add('hidden');
      return;
    }

    matches.slice(0, 10).forEach(tag => {
      const li = document.createElement('li');
      li.textContent = tag.name;
      li.setAttribute('role', 'option');
      li.setAttribute('data-slug', tag.slug);

      const isSelected = selectedTagSlugs.includes(tag.slug);
      if (isSelected) {
        li.setAttribute('aria-selected', 'true');
      }

      li.addEventListener('click', () => {
        if (!selectedTagSlugs.includes(tag.slug)) {
          selectedTagSlugs.push(tag.slug);
          renderSelectedTags();
        }
        input.value = '';
        dropdown.classList.add('hidden');
      });

      dropdown.appendChild(li);
    });

    dropdown.classList.remove('hidden');
  }

  input.addEventListener('input', () => {
    const q = input.value.toLowerCase().trim();
    if (!q) {
      dropdown.classList.add('hidden');
      return;
    }
    const matches = allTags.filter(
      t =>
        t.name.toLowerCase().includes(q) ||
        t.slug.toLowerCase().includes(q)
    );
    showDropdown(matches);
  });

  input.addEventListener('keydown', e => {
    if (e.key === 'Escape') {
      dropdown.classList.add('hidden');
    }
  });

  // The document-level click handler must only be registered once to avoid
  // accumulation across re-inits triggered by "Share another".
  if (!tagSearchInitialized) {
    tagSearchInitialized = true;
    document.addEventListener('click', e => {
      if (!input.contains(e.target as Node) && !dropdown.contains(e.target as Node)) {
        dropdown.classList.add('hidden');
      }
    });
  }
}

// ─── Form population ──────────────────────────────────────────────────────────

function populateForm(metadata: ClipMetadata | null, clipInfo: TwitchClipInfo): void {
  const titleInput = el<HTMLInputElement>('input-title');
  const previewWrap = el('clip-preview-wrap');

  if (metadata) {
    // Ensure the preview is visible (it may have been hidden in a previous init cycle).
    previewWrap.style.display = '';
    el<HTMLImageElement>('clip-thumbnail').src = metadata.thumbnail_url;
    el('clip-streamer').textContent = metadata.streamer_name;
    el('clip-game').textContent = metadata.game_name;
    const dur = metadata.duration ?? 0;
    el('clip-duration').textContent = dur > 0 ? `${dur.toFixed(1)}s` : '';
    titleInput.value = metadata.title;
  } else {
    // Fallback: hide preview and use clip ID as title placeholder.
    previewWrap.style.display = 'none';
    titleInput.placeholder = clipInfo.clipId;
  }
}

// ─── Submission ───────────────────────────────────────────────────────────────

async function handleSubmit(): Promise<void> {
  if (!currentClipInfo || !currentToken) return;

  const submitBtn = el<HTMLButtonElement>('btn-submit');
  const errorEl = el('form-error');

  const title = el<HTMLInputElement>('input-title').value.trim();
  const description = el<HTMLTextAreaElement>('input-description').value.trim();
  const isNsfw = el<HTMLInputElement>('input-nsfw').checked;

  if (!title) {
    errorEl.textContent = 'Please enter a title.';
    errorEl.classList.remove('hidden');
    return;
  }

  errorEl.classList.add('hidden');
  submitBtn.disabled = true;
  submitBtn.textContent = 'Submitting…';

  try {
    await submitClip(
      {
        clip_url: currentClipInfo.clipUrl,
        custom_title: title,
        tags: selectedTagSlugs.length > 0 ? selectedTagSlugs : undefined,
        is_nsfw: isNsfw,
        submission_reason: description || undefined,
      },
      currentToken,
      config
    );

    // Show success state.
    showView('success');

    // Also send a desktop notification.
    chrome.notifications.create({
      type: 'basic',
      iconUrl: '../icons/icon-48.png',
      title: 'Clip submitted!',
      message: `"${title}" is now pending review on Clipper.`,
    });
  } catch (err) {
    const message =
      err instanceof Error ? err.message : 'An unknown error occurred.';
    errorEl.textContent = message;
    errorEl.classList.remove('hidden');
    submitBtn.disabled = false;
    submitBtn.textContent = 'Share Clip';
  }
}

// ─── Bootstrap ────────────────────────────────────────────────────────────────

async function init(): Promise<void> {
  showView('loading');

  // 1. Check auth.
  const token = await getAccessToken(config);
  if (!token) {
    showView('login');
    return;
  }

  const user: UserProfile | null = await verifyAuth(token, config);
  if (!user) {
    showView('login');
    return;
  }

  currentToken = token;

  // Show user indicator.
  const badge = el('user-badge');
  badge.title = `Logged in as ${user.display_name ?? user.username}`;

  // 2. Detect clip.
  const response = await chrome.runtime.sendMessage({ type: 'GET_CURRENT_CLIP' }) as { clipInfo: TwitchClipInfo | null };
  currentClipInfo = response?.clipInfo ?? null;

  if (!currentClipInfo) {
    showView('no-clip');
    return;
  }

  // 3. Load metadata and tags in parallel.
  const [metadata, tags] = await Promise.allSettled([
    fetchClipMetadata(currentClipInfo.clipUrl, token, config),
    fetchTags(token, config),
  ]);

  currentMetadata = metadata.status === 'fulfilled' ? metadata.value : null;
  allTags = tags.status === 'fulfilled' ? tags.value : [];

  // 4. Populate and show form.
  populateForm(currentMetadata, currentClipInfo);
  setupTagSearch();
  showView('form');
}

// ─── Event wiring (registered once on DOMContentLoaded) ──────────────────────

document.addEventListener('DOMContentLoaded', async () => {
  // Load config overrides from storage before wiring any handlers so all
  // callbacks (including the login button) use the correct URLs.
  const stored = await chrome.storage.local.get(['apiBaseUrl', 'frontendUrl']) as Partial<ExtensionConfig>;
  if (stored.apiBaseUrl) config.apiBaseUrl = stored.apiBaseUrl;
  if (stored.frontendUrl) config.frontendUrl = stored.frontendUrl;

  // Login button – registered once; always reads current `config` value.
  el<HTMLButtonElement>('btn-login').addEventListener('click', () => {
    openLoginPage(config);
  });

  // Submit button – registered once to prevent duplicate submissions on re-init.
  el<HTMLButtonElement>('btn-submit').addEventListener('click', () => {
    void handleSubmit();
  });

  // Success actions.
  el<HTMLButtonElement>('btn-view-clpr').addEventListener('click', () => {
    chrome.tabs.create({ url: `${config.frontendUrl}/submit` });
  });

  el<HTMLButtonElement>('btn-share-another').addEventListener('click', () => {
    // Reset state and re-init without re-registering event handlers.
    selectedTagSlugs = [];
    currentClipInfo = null;
    currentMetadata = null;
    void init();
  });

  void init();
});
