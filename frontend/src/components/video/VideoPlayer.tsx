import { useEffect, useRef, useState, useId } from 'react';
import { ExternalLink } from 'lucide-react';
import { useVolumePreference } from '@/hooks';
import { usePlaybackControl } from '@/hooks/usePlaybackControl';
import { useTwitchEmbedFits } from '@/hooks/useTwitchEmbedFits';
import { useRegisterTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';
import { MutedIcon } from '@/components/ui';
import { cn } from '@/lib/utils';

import type { TwitchEmbedInstance as TwitchEmbed } from '@/types/twitchEmbed';

// Twitch Embed event constants
const TWITCH_EVENTS = {
    VIDEO_PLAY: 'video.play',
    VIDEO_PAUSE: 'pause',
    ENDED: 'ended',
    READY: 'ready',
};

export interface VideoPlayerProps {
    clipId: string;
    title: string;
    embedUrl: string;
    twitchClipId?: string;
    fit?: 'width' | 'height';
    className?: string;
    onEnded?: () => void;
    onPlay?: () => void;
    onPause?: () => void;
}

// Load the Twitch embed script once
let twitchScriptLoaded = false;
let twitchScriptPromise: Promise<void> | null = null;

function loadTwitchEmbedScript(): Promise<void> {
    if (twitchScriptLoaded && window.Twitch?.Embed) {
        return Promise.resolve();
    }
    if (twitchScriptPromise) return twitchScriptPromise;

    twitchScriptPromise = new Promise((resolve, reject) => {
        if (window.Twitch?.Embed) {
            twitchScriptLoaded = true;
            resolve();
            return;
        }
        const script = document.createElement('script');
        script.src = 'https://embed.twitch.tv/embed/v1.js';
        script.async = true;
        script.onload = () => {
            twitchScriptLoaded = true;
            resolve();
        };
        script.onerror = () => reject(new Error('Failed to load Twitch embed script'));
        document.head.appendChild(script);
    });
    return twitchScriptPromise;
}

export function VideoPlayer({
    clipId,
    title,
    embedUrl,
    twitchClipId,
    fit = 'width',
    className,
    onEnded,
    onPlay,
    onPause,
}: VideoPlayerProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const reactId = useId();
    const embedId = `twitch-embed-${clipId}-${reactId.replaceAll(':', '')}`;
    const embedRef = useRef<TwitchEmbed | null>(null);
    const [useJsEmbed, setUseJsEmbed] = useState(false);
    const { embedMuted, hasSetPreference, setUnmutedPreference } =
        useVolumePreference();
    // Twitch requires at least 400×300 and forbids covering the player, so a
    // smaller box links out and the sound control sits below the player.
    const fits = useTwitchEmbedFits(containerRef);
    useRegisterTwitchPlayer(containerRef, fits === true);

    // Global playback control — only one video plays at a time
    const playerId = useId();
    const { requestPlayback, registerPlayer } = usePlaybackControl(`video-${playerId}-${clipId}`);

    // Register pause function so other players can pause this one
    useEffect(() => {
        const unregister = registerPlayer(() => {
            try {
                embedRef.current?.getPlayer()?.pause();
            } catch {
                // Embed may not be ready yet
            }
        });
        return unregister;
    }, [registerPlayer]);

    // Store callbacks in refs to avoid stale closures
    const onEndedRef = useRef(onEnded);
    const onPlayRef = useRef(onPlay);
    const onPauseRef = useRef(onPause);
    useEffect(() => { onEndedRef.current = onEnded; }, [onEnded]);
    useEffect(() => { onPlayRef.current = onPlay; }, [onPlay]);
    useEffect(() => { onPauseRef.current = onPause; }, [onPause]);

    // Always use Twitch JS embed when twitchClipId is available
    useEffect(() => {
        if (!twitchClipId) return;

        loadTwitchEmbedScript()
            .then(() => setUseJsEmbed(true))
            .catch(() => setUseJsEmbed(false)); // Fall back to iframe
    }, [twitchClipId]);

    // Initialize Twitch JS embed
    useEffect(() => {
        if (!fits || !useJsEmbed || !twitchClipId || !window.Twitch?.Embed) return;

        let active = true;
        // Clear container
        const container = containerRef.current;
        if (!container) return;

        const embedDiv = document.createElement('div');
        embedDiv.id = embedId;
        embedDiv.style.width = '100%';
        embedDiv.style.height = '100%';
        embedDiv.style.position = 'absolute';
        embedDiv.style.inset = '0';

        // Remove old embeds
        const existing = container.querySelector('[id^="twitch-embed-"]');
        if (existing) existing.remove();
        container.appendChild(embedDiv);

        const parentDomain = window.location.hostname;

        try {
            const embed = new window.Twitch.Embed(embedId, {
                clip: twitchClipId,
                width: '100%',
                height: '100%',
                parent: [parentDomain],
                autoplay: true,
                muted: embedMuted,
                layout: 'video',
            });
            embedRef.current = embed;

            embed.addEventListener(TWITCH_EVENTS.VIDEO_PLAY, () => {
                requestPlayback();
                onPlayRef.current?.();
            });
            embed.addEventListener(TWITCH_EVENTS.VIDEO_PAUSE, () => {
                onPauseRef.current?.();
            });
            embed.addEventListener(TWITCH_EVENTS.ENDED, () => {
                onEndedRef.current?.();
            });
        } catch {
            // If JS embed fails, fall back to iframe
            queueMicrotask(() => {
                if (active) setUseJsEmbed(false);
            });
        }

        return () => {
            active = false;
            embedRef.current = null;
            const el = document.getElementById(embedId);
            if (el) el.remove();
        };
    }, [embedId, fits, useJsEmbed, twitchClipId, embedMuted, requestPlayback]);

    const parentDomain =
        typeof window !== 'undefined' ? window.location.hostname : 'localhost';
    const twitchEmbedUrl = `${embedUrl}&parent=${parentDomain}&autoplay=true&muted=${embedMuted}`;

    const twitchUrl = twitchClipId ?
            `https://clips.twitch.tv/${encodeURIComponent(twitchClipId)}`
        :   twitchClipUrlFromEmbed(embedUrl);
    const showSoundHint = fits && embedMuted && !hasSetPreference;

    return (
        <div
            className={cn(
                'flex flex-col',
                fit === 'height' ? 'h-full max-w-full items-center' : 'w-full',
                className,
            )}
        >
            <div
                ref={containerRef}
                className={cn(
                    'relative bg-black rounded-lg overflow-hidden aspect-video',
                    // In height-fit mode the row below is always reserved so the
                    // player never resizes when the sound control comes and goes.
                    fit === 'height' ?
                        'h-[calc(100%-2.75rem)] w-auto max-w-full'
                    :   'w-full',
                )}
            >
                {/* Use JS embed when available (provides play/pause/ended events),
                    fall back to raw iframe otherwise */}
                {fits && !useJsEmbed && (
                    <iframe
                        src={twitchEmbedUrl}
                        className='absolute inset-0 w-full h-full'
                        allowFullScreen
                        title={title}
                        allow='autoplay; fullscreen'
                    />
                )}

                {fits === false && (
                    <div className='absolute inset-0 grid place-items-center p-4 text-center text-white'>
                        <div>
                            <p className='text-sm font-semibold'>
                                Not enough room for the Twitch player.
                            </p>
                            <a
                                href={twitchUrl}
                                target='_blank'
                                rel='noopener noreferrer'
                                className='mt-2 inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-cyan-300 underline'
                            >
                                Watch on Twitch{' '}
                                <ExternalLink size={16} aria-hidden='true' />
                            </a>
                        </div>
                    </div>
                )}
            </div>

            {showSoundHint ?
                <button
                    type='button'
                    onClick={setUnmutedPreference}
                    className='inline-flex min-h-11 shrink-0 items-center gap-1.5 self-start text-xs font-semibold text-text-secondary hover:text-text-primary cursor-pointer'
                >
                    <MutedIcon size='sm' /> Sound off · Turn on for future clips
                </button>
            :   fit === 'height' && <div className='h-11 shrink-0' />}
        </div>
    );
}

function twitchClipUrlFromEmbed(embedUrl: string): string {
    try {
        const slug = new URL(embedUrl).searchParams.get('clip');
        if (slug) return `https://clips.twitch.tv/${encodeURIComponent(slug)}`;
    } catch {
        // Fall through to Twitch's clip index
    }
    return 'https://www.twitch.tv/directory';
}
