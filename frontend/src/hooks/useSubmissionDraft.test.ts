import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { useSubmissionDraft } from './useSubmissionDraft';
import type { SubmitClipRequest } from '../types/submission';
import type { Tag } from '../types/tag';
import { allowTestConsole } from '../test/setup';

// Mock localStorage
const localStorageMock = (() => {
    let store: Record<string, string> = {};

    return {
        getItem: (key: string) => store[key] || null,
        setItem: (key: string, value: string) => {
            store[key] = value.toString();
        },
        removeItem: (key: string) => {
            delete store[key];
        },
        clear: () => {
            store = {};
        },
    };
})();

Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
});

describe('useSubmissionDraft', () => {
    const initialFormData: SubmitClipRequest = {
        clip_url: '',
        custom_title: '',
        is_nsfw: false,
        submission_reason: '',
        broadcaster_name_override: '',
    };

    beforeEach(() => {
        localStorageMock.clear();
        vi.clearAllTimers();
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('should initialize with no draft', () => {
        const { result } = renderHook(() => useSubmissionDraft());

        expect(result.current.hasDraft).toBe(false);
        expect(result.current.lastSaved).toBeNull();
    });

    it('should save draft to localStorage', () => {
        const { result } = renderHook(() => useSubmissionDraft());

        const formData: SubmitClipRequest = {
            clip_url: 'https://clips.twitch.tv/test',
            custom_title: 'Test Title',
            is_nsfw: false,
            submission_reason: 'Test reason',
            broadcaster_name_override: '',
        };
        const tags: Tag[] = [
            {
                id: '1',
                name: 'Test Tag',
                slug: 'test-tag',
                usage_count: 0,
                created_at: new Date().toISOString(),
            },
        ];

        act(() => {
            result.current.saveDraft(formData, tags);
        });

        expect(result.current.hasDraft).toBe(true);
        expect(result.current.lastSaved).toBeGreaterThan(0);

        const stored = localStorageMock.getItem('submission_draft');
        expect(stored).toBeTruthy();
        const parsed = JSON.parse(stored!);
        expect(parsed.formData).toEqual(formData);
        expect(parsed.selectedTags).toEqual(tags);
    });

    it('should load draft from localStorage', () => {
        const formData: SubmitClipRequest = {
            clip_url: 'https://clips.twitch.tv/test',
            custom_title: 'Test Title',
            is_nsfw: false,
            submission_reason: 'Test reason',
            broadcaster_name_override: '',
        };
        const tags: Tag[] = [
            {
                id: '1',
                name: 'Test Tag',
                slug: 'test-tag',
                usage_count: 0,
                created_at: new Date().toISOString(),
            },
        ];

        // Pre-populate localStorage
        const draft = {
            formData,
            selectedTags: tags,
            lastSaved: Date.now(),
        };
        localStorageMock.setItem('submission_draft', JSON.stringify(draft));

        const { result } = renderHook(() => useSubmissionDraft());

        let loadedDraft;
        act(() => {
            loadedDraft = result.current.loadDraft();
        });

        expect(loadedDraft).toBeTruthy();
        expect(loadedDraft!.formData).toEqual(formData);
        expect(loadedDraft!.selectedTags).toEqual(tags);
        expect(result.current.hasDraft).toBe(true);
    });

    it('should clear draft from localStorage', () => {
        const formData: SubmitClipRequest = {
            clip_url: 'https://clips.twitch.tv/test',
            custom_title: 'Test Title',
            is_nsfw: false,
            submission_reason: '',
            broadcaster_name_override: '',
        };

        const { result } = renderHook(() => useSubmissionDraft());

        // Save a draft first
        act(() => {
            result.current.saveDraft(formData, []);
        });

        expect(result.current.hasDraft).toBe(true);

        // Clear the draft
        act(() => {
            result.current.clearDraft();
        });

        expect(result.current.hasDraft).toBe(false);
        expect(result.current.lastSaved).toBeNull();
        expect(localStorageMock.getItem('submission_draft')).toBeNull();
    });

    it('should detect form with content', () => {
        const { result } = renderHook(() => useSubmissionDraft());

        // Empty form
        expect(result.current.hasContent(initialFormData, [])).toBe(false);

        // Form with URL
        expect(
            result.current.hasContent(
                { ...initialFormData, clip_url: 'https://clips.twitch.tv/test' },
                []
            )
        ).toBe(true);

        // Form with title
        expect(
            result.current.hasContent(
                { ...initialFormData, custom_title: 'Test' },
                []
            )
        ).toBe(true);

        // Form with tags
        const tags: Tag[] = [
            {
                id: '1',
                name: 'Test Tag',
                slug: 'test-tag',
                usage_count: 0,
                created_at: new Date().toISOString(),
            },
        ];
        expect(result.current.hasContent(initialFormData, tags)).toBe(true);
    });

    it('should auto-save after 30 seconds', () => {
        const { result } = renderHook(() => useSubmissionDraft());

        const formData: SubmitClipRequest = {
            clip_url: 'https://clips.twitch.tv/test',
            custom_title: 'Test Title',
            is_nsfw: false,
            submission_reason: '',
            broadcaster_name_override: '',
        };

        act(() => {
            result.current.startAutoSave(formData, []);
        });

        // Initially no draft saved
        expect(result.current.hasDraft).toBe(false);

        // Advance time by 30 seconds
        act(() => {
            vi.advanceTimersByTime(30000);
        });

        // Draft should now be saved
        expect(result.current.hasDraft).toBe(true);
        expect(result.current.lastSaved).toBeGreaterThan(0);
    });

    it('should not auto-save empty form', () => {
        const { result } = renderHook(() => useSubmissionDraft());

        act(() => {
            result.current.startAutoSave(initialFormData, []);
        });

        // Advance time by 30 seconds
        act(() => {
            vi.advanceTimersByTime(30000);
        });

        // Draft should not be saved for empty form
        expect(result.current.hasDraft).toBe(false);
    });

    it('should handle localStorage errors gracefully', () => {
        allowTestConsole('error', /Failed to save draft:.*Storage quota exceeded/);
        const { result } = renderHook(() => useSubmissionDraft());

        // Mock localStorage.setItem to throw an error
        const originalSetItem = localStorageMock.setItem;
        localStorageMock.setItem = vi.fn(() => {
            throw new Error('Storage quota exceeded');
        });

        const formData: SubmitClipRequest = {
            clip_url: 'https://clips.twitch.tv/test',
            custom_title: 'Test Title',
            is_nsfw: false,
            submission_reason: '',
            broadcaster_name_override: '',
        };

        // Should not throw
        expect(() => {
            act(() => {
                result.current.saveDraft(formData, []);
            });
        }).not.toThrow();

        // Restore original function
        localStorageMock.setItem = originalSetItem;
    });

    it('should handle corrupt draft data gracefully', () => {
        allowTestConsole('error', /Failed to load draft:.*not valid json/, 2);
        // Put corrupt data in localStorage
        localStorageMock.setItem('submission_draft', 'not valid json');

        const { result } = renderHook(() => useSubmissionDraft());

        // Should not throw
        expect(() => {
            result.current.loadDraft();
        }).not.toThrow();

        expect(result.current.loadDraft()).toBeNull();
    });
});
