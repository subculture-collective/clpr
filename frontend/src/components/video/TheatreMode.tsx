import { useState, useCallback, useMemo, useEffect, useRef, useId } from 'react';
import { cn } from '@/lib/utils';
import {
    useTheatreMode,
    useQualityPreference,
    useKeyboardControls,
} from '@/hooks';
import { usePlaybackControl } from '@/context/PlaybackContext';
import type { VideoQuality } from '@/lib/adaptive-bitrate';
import { HlsPlayer } from './HlsPlayer';
import { QualitySelector } from './QualitySelector';
import { BitrateIndicator } from './BitrateIndicator';
import { PlaybackControls } from './PlaybackControls';

export interface TheatreModeProps {
    title: string;
    hlsUrl?: string; // Optional HLS URL for clips that support it
    className?: string;
    fit?: 'width' | 'height';
    // Watch history props
    resumePosition?: number;
    hasProgress?: boolean;
    isLoadingProgress?: boolean;
    onProgressUpdate?: (currentTime: number) => void;
    onPause?: (currentTime: number) => void;
    onEnded?: (currentTime: number) => void;
}

const AVAILABLE_QUALITIES: VideoQuality[] = [
    '480p',
    '720p',
    '1080p',
    '2K',
    '4K',
    'auto',
];

/**
 * Theatre mode player component with HLS support
 * Provides immersive viewing experience with quality selection and keyboard shortcuts
 * Falls back gracefully when HLS is not available
 */
export function TheatreMode({
    title,
    hlsUrl,
    className,
    fit = 'width',
    resumePosition = 0,
    hasProgress = false,
    isLoadingProgress = false,
    onProgressUpdate,
    onPause,
    onEnded,
}: TheatreModeProps) {
    const {
        isTheatreMode,
        isFullscreen,
        isPictureInPicture,
        containerRef,
        videoRef,
        toggleTheatreMode,
        toggleFullscreen,
        togglePictureInPicture,
    } = useTheatreMode();

    // Global playback control — only one video plays at a time
    const theatrePlayerId = useId();
    const { requestPlayback, registerPlayer } = usePlaybackControl(`theatre-${theatrePlayerId}`);

    useEffect(() => {
        const unregister = registerPlayer(() => {
            videoRef.current?.pause();
        });
        return unregister;
    }, [registerPlayer, videoRef]);

    const { quality, setQuality } = useQualityPreference();
    const [bandwidth, setBandwidth] = useState<number>();
    const [bufferHealth, setBufferHealth] = useState(100);
    const [showControls, setShowControls] = useState(true);
    const [showResumePrompt, setShowResumePrompt] = useState(false);
    const [hasAppliedResume, setHasAppliedResume] = useState(false);

    // Use refs to store latest callbacks to avoid re-attaching event listeners
    const onProgressUpdateRef = useRef(onProgressUpdate);
    const onPauseRef = useRef(onPause);
    const onEndedRef = useRef(onEnded);

    // Update refs when callbacks change
    useEffect(() => {
        onProgressUpdateRef.current = onProgressUpdate;
    }, [onProgressUpdate]);

    useEffect(() => {
        onPauseRef.current = onPause;
    }, [onPause]);

    useEffect(() => {
        onEndedRef.current = onEnded;
    }, [onEnded]);

    // Reset hasAppliedResume when clip changes
    useEffect(() => {
        queueMicrotask(() => {
            setHasAppliedResume(false);
            setShowResumePrompt(false);
        });
    }, [hlsUrl]);

    // Determine if HLS is available
    const hasHlsSupport = useMemo(() => !!hlsUrl, [hlsUrl]);

    // Current quality for display
    const [currentQuality, setCurrentQuality] = useState<VideoQuality>(quality);

    // Handle quality change
    const handleQualityChange = useCallback(
        (newQuality: VideoQuality) => {
            setQuality(newQuality);
            setCurrentQuality(newQuality);
        },
        [setQuality],
    );

    // Handle play/pause toggle
    const handlePlayPause = useCallback(() => {
        const video = videoRef.current;
        if (!video) return;

        if (video.paused) {
            requestPlayback();
            video.play();
        } else {
            video.pause();
        }
    }, [videoRef, requestPlayback]);

    // Handle mute toggle
    const handleMute = useCallback(() => {
        const video = videoRef.current;
        if (!video) return;

        video.muted = !video.muted;
    }, [videoRef]);

    // Keyboard shortcuts
    useKeyboardControls(
        {
            onPlayPause: handlePlayPause,
            onMute: handleMute,
            onFullscreen: toggleFullscreen,
            onTheatreMode: toggleTheatreMode,
            onPictureInPicture: togglePictureInPicture,
        },
        hasHlsSupport,
    );

    // Show resume prompt when progress is available and not loading
    useEffect(() => {
        if (
            hasProgress &&
            !isLoadingProgress &&
            !hasAppliedResume &&
            resumePosition > 5
        ) {
            queueMicrotask(() => setShowResumePrompt(true));
        }
    }, [hasProgress, isLoadingProgress, hasAppliedResume, resumePosition]);

    // Handle resume position
    const handleResume = useCallback(() => {
        const video = videoRef.current;
        if (video && resumePosition > 0) {
            // Wait for video metadata to load before seeking
            if (video.readyState >= 1) {
                video.currentTime = resumePosition;
            } else {
                // If metadata not loaded yet, wait for it
                const handleLoadedMetadata = () => {
                    video.currentTime = resumePosition;
                    video.removeEventListener(
                        'loadedmetadata',
                        handleLoadedMetadata,
                    );
                };
                video.addEventListener('loadedmetadata', handleLoadedMetadata);
            }
            setShowResumePrompt(false);
            setHasAppliedResume(true);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [resumePosition]);

    const handleDismissResume = useCallback(() => {
        setShowResumePrompt(false);
        setHasAppliedResume(true);
    }, []);

    // Handle Escape key for resume prompt
    useEffect(() => {
        if (!showResumePrompt) return;

        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                handleDismissResume();
            }
        };

        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [showResumePrompt, handleDismissResume]);

    // Progress tracking with timeupdate event
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const handleTimeUpdate = () => {
            onProgressUpdateRef.current?.(video.currentTime);
        };

        video.addEventListener('timeupdate', handleTimeUpdate);
        return () => {
            video.removeEventListener('timeupdate', handleTimeUpdate);
        };
    }, [videoRef]);

    // Play tracking — notify global playback control when this video starts
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const handlePlayEvent = () => {
            requestPlayback();
        };

        video.addEventListener('playing', handlePlayEvent);
        return () => {
            video.removeEventListener('playing', handlePlayEvent);
        };
    }, [videoRef, requestPlayback]);

    // Pause tracking
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const handlePauseEvent = () => {
            onPauseRef.current?.(video.currentTime);
        };

        video.addEventListener('pause', handlePauseEvent);
        return () => {
            video.removeEventListener('pause', handlePauseEvent);
        };
    }, [videoRef]);

    // End tracking
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const handleEndedEvent = () => {
            onEndedRef.current?.(video.currentTime);
        };

        video.addEventListener('ended', handleEndedEvent);
        return () => {
            video.removeEventListener('ended', handleEndedEvent);
        };
    }, [videoRef]);

    // Show/hide controls on mouse movement
    useEffect(() => {
        let timeout: ReturnType<typeof setTimeout>;

        const handleMouseMove = () => {
            setShowControls(true);

            // Hide controls after 3 seconds of inactivity in theatre/fullscreen mode
            if (isTheatreMode || isFullscreen) {
                clearTimeout(timeout);
                timeout = setTimeout(() => {
                    setShowControls(false);
                }, 3000);
            }
        };

        const container = containerRef.current;
        if (container) {
            container.addEventListener('mousemove', handleMouseMove);
        }

        return () => {
            if (container) {
                container.removeEventListener('mousemove', handleMouseMove);
            }
            clearTimeout(timeout);
        };
    }, [isTheatreMode, isFullscreen, containerRef]);

    const containerSizeClass =
        isTheatreMode ? 'fixed inset-0 z-50 w-screen h-screen'
        : fit === 'height' ?
            'h-full w-auto max-w-full aspect-video rounded-lg overflow-hidden'
        :   'w-full aspect-video rounded-lg overflow-hidden';

    // If no HLS support, show message
    if (!hasHlsSupport) {
        return (
            <div
                className={cn(
                    'relative bg-neutral-900',
                    containerSizeClass,
                    'flex items-center justify-center',
                    'min-h-[200px]',
                    className,
                )}
            >
                <div className='text-center p-8'>
                    <svg
                        className='w-16 h-16 mx-auto mb-4 text-neutral-600'
                        fill='none'
                        stroke='currentColor'
                        viewBox='0 0 24 24'
                    >
                        <path
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            strokeWidth={2}
                            d='M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z'
                        />
                    </svg>
                    <p className='text-white text-lg font-semibold mb-2'>
                        Theatre Mode Coming Soon
                    </p>
                    <p className='text-neutral-400 text-sm max-w-md mx-auto'>
                        Theatre mode with adaptive quality streaming will be
                        available once HLS video processing is set up for this
                        clip.
                    </p>
                </div>
            </div>
        );
    }

    return (
        <div
            ref={containerRef}
            className={cn(
                'relative bg-black group',
                containerSizeClass,
                className,
            )}
        >
            {/* HLS Video Player */}
            <HlsPlayer
                src={hlsUrl!}
                quality={quality}
                autoQuality={quality === 'auto'}
                videoRef={videoRef}
                onQualityChange={setCurrentQuality}
                onBandwidthUpdate={setBandwidth}
                onBufferHealthUpdate={setBufferHealth}
                className='absolute inset-0 w-full h-full object-contain'
            />

            {/* Bitrate Indicator */}
            {(isTheatreMode || isFullscreen) && (
                <BitrateIndicator
                    bandwidth={bandwidth}
                    bufferHealth={bufferHealth}
                    currentQuality={currentQuality}
                    className={cn(
                        'transition-opacity duration-300',
                        showControls ? 'opacity-100' : 'opacity-0',
                    )}
                />
            )}

            {/* Top Bar - Title */}
            <div
                className={cn(
                    'absolute top-0 left-0 right-0 p-4',
                    'bg-gradient-to-b from-black/80 to-transparent',
                    'transition-opacity duration-300 pointer-events-none',
                    showControls ? 'opacity-100' : 'opacity-0',
                )}
            >
                <h2 className='text-white text-lg font-semibold line-clamp-1'>
                    {title}
                </h2>
            </div>

            {/* Resume Prompt */}
            {showResumePrompt && (
                <div
                    className='absolute inset-0 flex items-center justify-center z-40 pointer-events-none'
                    role='dialog'
                    aria-modal='true'
                    aria-labelledby='resume-playback-title'
                >
                    <div className='bg-black/90 backdrop-blur-sm rounded-lg p-6 max-w-md mx-4 pointer-events-auto'>
                        <h3
                            id='resume-playback-title'
                            className='text-white text-lg font-semibold mb-2'
                        >
                            Resume Playback?
                        </h3>
                        <p className='text-white/80 text-sm mb-4'>
                            You were at {Math.floor(resumePosition / 60)}:
                            {String(Math.floor(resumePosition % 60)).padStart(
                                2,
                                '0',
                            )}
                        </p>
                        <div className='flex gap-3'>
                            <button
                                onClick={handleDismissResume}
                                className='flex-1 px-4 py-2 bg-white/10 hover:bg-white/20 text-white rounded transition-colors'
                            >
                                Start Over
                            </button>
                            <button
                                onClick={handleResume}
                                className='flex-1 px-4 py-2 bg-primary-500 hover:bg-primary-600 text-white rounded transition-colors'
                            >
                                Resume
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Bottom Controls */}
            <div
                className={cn(
                    'absolute bottom-0 left-0 right-0 p-4',
                    'bg-gradient-to-t from-black/80 to-transparent',
                    'transition-opacity duration-300',
                    showControls ?
                        'opacity-100 pointer-events-auto'
                    :   'opacity-0 pointer-events-none',
                )}
            >
                <div className='space-y-3'>
                    {/* Playback Controls */}
                    <PlaybackControls videoRef={videoRef} />

                    {/* Action Buttons */}
                    <div className='flex items-center justify-between'>
                        <div className='flex items-center gap-2'>
                            {/* Keyboard shortcuts hint */}
                            <span className='text-white/60 text-xs'>
                                Keyboard: Space=play/pause, T=theatre,
                                F=fullscreen, M=mute, P=pip
                            </span>
                        </div>

                        <div className='flex items-center gap-2'>
                            {/* Quality Selector */}
                            <QualitySelector
                                value={quality}
                                onChange={handleQualityChange}
                                availableQualities={AVAILABLE_QUALITIES}
                            />

                            {/* Picture-in-Picture */}
                            <button
                                onClick={togglePictureInPicture}
                                className='p-2 bg-black/60 hover:bg-black/80 rounded transition-colors min-w-[44px] min-h-[44px] flex items-center justify-center'
                                aria-label={
                                    isPictureInPicture ?
                                        'Exit picture-in-picture'
                                    :   'Enter picture-in-picture'
                                }
                                title={
                                    isPictureInPicture ? 'Exit PiP' : (
                                        'Picture-in-Picture (P)'
                                    )
                                }
                            >
                                <svg
                                    className='w-5 h-5 text-white'
                                    fill='none'
                                    stroke='currentColor'
                                    viewBox='0 0 24 24'
                                >
                                    <path
                                        strokeLinecap='round'
                                        strokeLinejoin='round'
                                        strokeWidth={2}
                                        d='M7 4v16M17 4v16M3 8h18M3 16h18'
                                    />
                                </svg>
                            </button>

                            {/* Theatre Mode */}
                            <button
                                onClick={toggleTheatreMode}
                                className={cn(
                                    'px-4 py-2 rounded transition-colors font-medium text-sm',
                                    isTheatreMode ?
                                        'bg-primary-500 hover:bg-primary-600 text-white'
                                    :   'bg-black/60 hover:bg-black/80 text-white',
                                )}
                                aria-label={
                                    isTheatreMode ? 'Exit theatre mode' : (
                                        'Enter theatre mode'
                                    )
                                }
                                title={
                                    isTheatreMode ? 'Exit Theatre Mode' : (
                                        'Theatre Mode (T)'
                                    )
                                }
                            >
                                {isTheatreMode ?
                                    'Exit Theatre'
                                :   'Theatre Mode'}
                            </button>

                            {/* Fullscreen */}
                            <button
                                onClick={toggleFullscreen}
                                className='p-2 bg-black/60 hover:bg-black/80 rounded transition-colors min-w-[44px] min-h-[44px] flex items-center justify-center'
                                aria-label={
                                    isFullscreen ? 'Exit fullscreen' : (
                                        'Enter fullscreen'
                                    )
                                }
                                title={
                                    isFullscreen ? 'Exit Fullscreen' : (
                                        'Fullscreen (F)'
                                    )
                                }
                            >
                                <svg
                                    className='w-5 h-5 text-white'
                                    fill='none'
                                    stroke='currentColor'
                                    viewBox='0 0 24 24'
                                >
                                    {isFullscreen ?
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M6 18L18 6M6 6l12 12'
                                        />
                                    :   <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4'
                                        />
                                    }
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
