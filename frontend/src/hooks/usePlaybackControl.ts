import { useContext } from 'react';
import { PlaybackContext } from '@/context/playback-context';

export function usePlaybackControl(playerId: string) {
    const context = useContext(PlaybackContext);
    if (!context) {
        return {
            requestPlayback: () => {},
            registerPlayer: (_pauseFn: () => void) => () => {},
        };
    }

    return {
        requestPlayback: () => context.requestPlayback(playerId),
        registerPlayer: (pauseFn: () => void) =>
            context.registerPlayer(playerId, pauseFn),
    };
}
