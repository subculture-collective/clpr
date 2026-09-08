export interface TwitchPlayerInstance {
    pause: () => void;
    play: () => void;
    getMuted: () => boolean;
    setMuted: (muted: boolean) => void;
}
export interface TwitchEmbedInstance {
    addEventListener: (event: string, callback: () => void) => void;
    getPlayer: () => TwitchPlayerInstance;
    destroy?: () => void;
}
declare global {
    interface Window {
        Twitch?: {
            Embed: new (elementId: string, options: Record<string, unknown>) => TwitchEmbedInstance;
        };
    }
}
