import { useEffect, useRef, useState, useCallback, useId } from 'react';
import { useVolumePreference } from '@/hooks';
import { usePlaybackControl } from '@/context/PlaybackContext';
import { MutedIcon } from '@/components/ui';
import { cn } from '@/lib/utils';

// Twitch Embed types
declare global {
    interface Window {
        Twitch?: {
            Embed: new (
                elementId: string,
                options: Record<string, unknown>,
            ) => TwitchEmbed;
        };
    }
}

interface TwitchEmbed {
    addEventListener: (event: string, callback: () => void) => void;
    getPlayer: () => TwitchPlayer;
}

interface TwitchPlayer {
    pause: () => void;
    play: () => void;
    getMuted: () => boolean;
    setMuted: (muted: boolean) => void;
}

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
    const embedIdRef = useRef(`twitch-embed-${clipId}-${Date.now()}`);
    const embedRef = useRef<TwitchEmbed | null>(null);
    const [useJsEmbed, setUseJsEmbed] = useState(false);
    const [showMutedIndicator, setShowMutedIndicator] = useState(true);
    const { embedMuted, hasSetPreference, setUnmutedPreference } =
        useVolumePreference();

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
        if (!useJsEmbed || !twitchClipId || !window.Twitch?.Embed) return;

        // Clear container
        const container = containerRef.current;
        if (!container) return;

        const embedDiv = document.createElement('div');
        const embedId = embedIdRef.current;
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
            setUseJsEmbed(false);
        }

        return () => {
            embedRef.current = null;
            const el = document.getElementById(embedId);
            if (el) el.remove();
        };
    }, [useJsEmbed, twitchClipId, embedMuted]);

    // Auto-hide muted indicator after 3 seconds
    useEffect(() => {
        if (!embedMuted || hasSetPreference) return;
        const timer = setTimeout(() => setShowMutedIndicator(false), 3000);
        return () => clearTimeout(timer);
    }, [embedMuted, hasSetPreference]);

    const parentDomain =
        typeof window !== 'undefined' ? window.location.hostname : 'localhost';
    const twitchEmbedUrl = `${embedUrl}&parent=${parentDomain}&autoplay=true&muted=${embedMuted}`;

    const containerClasses = cn(
        'relative bg-black rounded-lg overflow-hidden aspect-video',
        fit === 'height' ? 'h-full w-auto max-w-full' : 'w-full',
        className,
    );

    return (
        <div ref={containerRef} className={containerClasses}>
            {/* Use JS embed when available (provides play/pause/ended events),
                fall back to raw iframe otherwise */}
            {!useJsEmbed && (
                <iframe
                    src={twitchEmbedUrl}
                    className='absolute inset-0 w-full h-full'
                    allowFullScreen
                    title={title}
                    allow='autoplay; fullscreen'
                />
            )}

            {/* Muted indicator */}
            {embedMuted && !hasSetPreference && showMutedIndicator && (
                <div
                    className={cn(
                        'absolute top-3 left-3 bg-black/70 hover:bg-black/90 text-white px-2 py-1 rounded text-xs font-medium flex items-center gap-1 cursor-pointer transition-opacity duration-500 pointer-events-auto z-10',
                        showMutedIndicator ? 'opacity-100' : 'opacity-0',
                    )}
                    onClick={setUnmutedPreference}
                    role='button'
                    tabIndex={0}
                    onKeyDown={e => {
                        if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            setUnmutedPreference();
                        }
                    }}
                    aria-label='Video is muted, click to enable sound on future videos'
                    title='Video starts muted for autoplay compatibility. Click to enable sound on future videos.'
                >
                    <MutedIcon size='sm' />
                    <span>Muted</span>
                </div>
            )}
        </div>
    );
}
