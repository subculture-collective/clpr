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
});
afterEach(() => { delete window.Twitch; vi.useRealTimers(); });

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

    it('supports keyboard and pointer sound preference and expires the initial hint', () => {
        vi.useFakeTimers();
        render(<VideoPlayer {...props} />);
        const sound = screen.getByRole('button', { name: /enable sound/ });
        fireEvent.keyDown(sound, { key: 'Tab' });
        expect(volume.setUnmutedPreference).not.toHaveBeenCalled();
        fireEvent.keyDown(sound, { key: 'Enter' });
        fireEvent.keyDown(sound, { key: ' ' });
        fireEvent.click(sound);
        expect(volume.setUnmutedPreference).toHaveBeenCalledTimes(3);
        act(() => vi.advanceTimersByTime(3000));
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
