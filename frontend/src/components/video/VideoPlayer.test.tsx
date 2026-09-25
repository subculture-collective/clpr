import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { VideoPlayer } from './VideoPlayer';

const playback = vi.hoisted(() => ({ requestPlayback: vi.fn(), registerPlayer: vi.fn(), unregister: vi.fn() }));
const volume = vi.hoisted(() => ({ embedMuted: true, hasSetPreference: false, setUnmutedPreference: vi.fn() }));
vi.mock('@/hooks', () => ({ useVolumePreference: () => volume }));
vi.mock('@/hooks/usePlaybackControl', () => ({ usePlaybackControl: () => playback }));
const props = { clipId: 'clip', title: 'Clip video', embedUrl: 'https://clips.twitch.tv/embed?clip=slug' };

beforeEach(() => {
    playback.requestPlayback.mockClear();
    playback.unregister.mockClear();
    playback.registerPlayer.mockReset().mockReturnValue(playback.unregister);
    volume.setUnmutedPreference.mockClear();
    volume.embedMuted = true;
    volume.hasSetPreference = false;
    stubBoxSize(640, 360);
});
afterEach(() => { delete window.Twitch; vi.useRealTimers(); vi.restoreAllMocks(); });

function stubBoxSize(width: number, height: number) {
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({ width, height } as DOMRect);
}

describe('video playback boundary', () => {
    it('keeps the iframe playable with parent and mute preferences when no SDK clip is available', () => {
        const { rerender } = render(<VideoPlayer {...props} />);
        const frame = screen.getByTitle('Clip video') as HTMLIFrameElement;
        const url = new URL(frame.src);
        expect(url.hostname).toBe('clips.twitch.tv');
        expect(url.searchParams.get('parent')).toBe(window.location.hostname);
        expect(url.searchParams.get('muted')).toBe('true');
        volume.embedMuted = false;
        volume.hasSetPreference = true;
        rerender(<VideoPlayer {...props} />);
        expect(new URL(frame.src).searchParams.get('muted')).toBe('false');
        expect(screen.queryByRole('button')).toBeNull();
    });

    it('offers the sound preference below the player, not over it', () => {
        render(<VideoPlayer {...props} />);
        const sound = screen.getByRole('button', { name: /turn on for future clips/i });
        expect(screen.getByTitle('Clip video').parentElement).not.toContainElement(sound);
        fireEvent.click(sound);
        expect(volume.setUnmutedPreference).toHaveBeenCalledOnce();
    });

    it('links to Twitch instead of embedding below the 400×300 minimum', () => {
        stubBoxSize(320, 180);
        render(<VideoPlayer {...props} />);
        expect(screen.queryByTitle('Clip video')).toBeNull();
        expect(screen.getByRole('link', { name: /watch on twitch/i })).toHaveAttribute('href', 'https://clips.twitch.tv/slug');
        expect(screen.queryByRole('button')).toBeNull();
    });

    it('coordinates playback, calls current callbacks, and releases its registration on unmount', async () => {
        const listeners: Record<string, () => void> = {};
        const pause = vi.fn();
        const Embed = vi.fn(function (_elementId: string, _options: Record<string, unknown>) {
            return { addEventListener: (event: string, callback: () => void) => { listeners[event] = callback; },
                getPlayer: () => ({ pause, play: vi.fn(), getMuted: () => true, setMuted: vi.fn() }) };
        });
        window.Twitch = { Embed };
        const oldPlay = vi.fn(), onPlay = vi.fn(), onPause = vi.fn(), onEnded = vi.fn();
        const { rerender, unmount, container } = render(<VideoPlayer {...props} twitchClipId='slug' onPlay={oldPlay} />);
        await waitFor(() => expect(Embed).toHaveBeenCalledTimes(1));
        expect(Embed.mock.calls[0][1]).toMatchObject({ clip: 'slug', parent: [window.location.hostname], muted: true });
        rerender(<VideoPlayer {...props} twitchClipId='slug' onPlay={onPlay} onPause={onPause} onEnded={onEnded} />);
        act(() => { listeners['video.play'](); listeners.pause(); listeners.ended(); });
        expect(playback.requestPlayback).toHaveBeenCalledOnce();
        expect(oldPlay).not.toHaveBeenCalled();
        expect(onPlay).toHaveBeenCalledOnce();
        expect(onPause).toHaveBeenCalledOnce();
        expect(onEnded).toHaveBeenCalledOnce();
        act(() => playback.registerPlayer.mock.calls[0][0]());
        expect(pause).toHaveBeenCalledOnce();
        expect(container.querySelector('[id^="twitch-embed-"]')).not.toBeNull();
        unmount();
        expect(playback.unregister).toHaveBeenCalledOnce();
        expect(container.childElementCount).toBe(0);
    });

    it('falls back to the official iframe when SDK construction fails', async () => {
        window.Twitch = { Embed: vi.fn(function () { throw new Error('SDK unavailable'); }) };
        render(<VideoPlayer {...props} twitchClipId='slug' />);
        await waitFor(() => expect(window.Twitch?.Embed).toHaveBeenCalled());
        await waitFor(() => expect(screen.getByTitle('Clip video')).toHaveAttribute('src', expect.stringContaining('clips.twitch.tv/embed')));
    });
});
