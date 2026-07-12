import { createContext } from 'react';

export interface PlaybackContextType {
    requestPlayback: (playerId: string) => void;
    registerPlayer: (playerId: string, pauseFn: () => void) => () => void;
}

export const PlaybackContext = createContext<PlaybackContextType | null>(null);
