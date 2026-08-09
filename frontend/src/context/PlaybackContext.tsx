import { useCallback, useRef } from 'react';
import { PlaybackContext } from './playback-context';

export function PlaybackProvider({ children }: { children: React.ReactNode }) {
    // Map of player IDs to their pause functions
    const playersRef = useRef<Map<string, () => void>>(new Map());
    const activePlayerRef = useRef<string | null>(null);

    const registerPlayer = useCallback((playerId: string, pauseFn: () => void) => {
        playersRef.current.set(playerId, pauseFn);
        return () => {
            playersRef.current.delete(playerId);
            if (activePlayerRef.current === playerId) {
                activePlayerRef.current = null;
            }
        };
    }, []);

    const requestPlayback = useCallback((playerId: string) => {
        // Pause all other players
        playersRef.current.forEach((pauseFn, id) => {
            if (id !== playerId) {
                pauseFn();
            }
        });
        activePlayerRef.current = playerId;
    }, []);

    return (
        <PlaybackContext.Provider value={{ requestPlayback, registerPlayer }}>
            {children}
        </PlaybackContext.Provider>
    );
}
