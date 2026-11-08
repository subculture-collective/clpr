import { useEffect, useState } from 'react';

/** A Twitch clip embed cannot report its end, so playlists advance on the clip's duration. */
export function AutoAdvanceCountdown({
    duration,
    paused,
    onPausedChange,
    onComplete,
}: {
    duration: number;
    paused: boolean;
    onPausedChange: (paused: boolean) => void;
    onComplete: () => void;
}) {
    const [countdown, setCountdown] = useState<number | null>(() =>
        Math.ceil(duration + 3),
    );

    useEffect(() => {
        if (paused || countdown === null) return;

        const timeout = window.setTimeout(() => {
            if (countdown <= 1) {
                setCountdown(null);
                onComplete();
                return;
            }
            setCountdown(countdown - 1);
        }, 1000);

        return () => window.clearTimeout(timeout);
    }, [countdown, onComplete, paused]);

    useEffect(() => {
        const handleVisibility = () => {
            if (document.hidden) onPausedChange(true);
        };
        document.addEventListener('visibilitychange', handleVisibility);
        return () =>
            document.removeEventListener('visibilitychange', handleVisibility);
    }, [onPausedChange]);

    if (countdown === null) return null;

    return (
        <div className='w-full px-4 py-1.5 flex items-center justify-center gap-3 bg-surface/80 text-[12px]'>
            {paused ? (
                <>
                    <span className='text-text-secondary'>
                        Auto-advance paused
                    </span>
                    <button
                        onClick={() => onPausedChange(false)}
                        className='text-cta hover:text-cta-hover transition-colors cursor-pointer font-medium'
                    >
                        Resume
                    </button>
                    <button
                        onClick={() => setCountdown(null)}
                        className='text-text-tertiary hover:text-text-secondary transition-colors cursor-pointer'
                    >
                        Cancel
                    </button>
                </>
            ) : (
                <>
                    <span className='text-text-secondary'>
                        Next clip in {countdown}s
                    </span>
                    <button
                        onClick={() => onPausedChange(true)}
                        className='text-cta hover:text-cta-hover transition-colors cursor-pointer font-medium'
                    >
                        Pause
                    </button>
                    <button
                        onClick={() => {
                            setCountdown(null);
                            onComplete();
                        }}
                        className='text-text-tertiary hover:text-text-secondary transition-colors cursor-pointer'
                    >
                        Skip now
                    </button>
                </>
            )}
        </div>
    );
}
