import { createContext, useContext, useCallback, useRef } from 'react';

interface PlaybackContextType {
    /** Call this when a player starts playing. Returns a unique player ID. */
    requestPlayback: (playerId: string) => void;
    /** Register a pause callback for a player. Returns unregister function. */
    registerPlayer: (playerId: string, pauseFn: () => void) => () => void;
}

const PlaybackContext = createContext<PlaybackContextType | null>(null);

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

/** Hook to integrate a player with the global playback system. */
export function usePlaybackControl(playerId: string) {
    const ctx = useContext(PlaybackContext);
    if (!ctx) {
        // Graceful fallback if provider is missing
        return {
            requestPlayback: () => {},
            registerPlayer: (_pauseFn: () => void) => () => {},
        };
    }
    return {
        requestPlayback: () => ctx.requestPlayback(playerId),
        registerPlayer: (pauseFn: () => void) => ctx.registerPlayer(playerId, pauseFn),
    };
}
