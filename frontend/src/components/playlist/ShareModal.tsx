import { useState, useEffect } from 'react';
import { X, Copy, Check, Share2 } from 'lucide-react';
import apiClient from '@/lib/api';
import type { GetShareLinkResponse, TrackShareRequest } from '@/types/playlist';

interface ShareModalProps {
    playlistId: string;
    onClose: () => void;
}

export function ShareModal({ playlistId, onClose }: ShareModalProps) {
    const [shareData, setShareData] = useState<GetShareLinkResponse | null>(null);
    const [copied, setCopied] = useState<'url' | 'embed' | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchShareLink = async () => {
            try {
                setLoading(true);
                const response = await apiClient.get<{ success: boolean; data: GetShareLinkResponse; error?: { message: string } }>(
                    `/playlists/${playlistId}/share-link`
                );

                if (response.data.success) {
                    setShareData(response.data.data);
                } else {
                    throw new Error(response.data.error?.message || 'Failed to get share link');
                }
            } catch (err) {
                setError(err instanceof Error ? err.message : 'An error occurred');
            } finally {
                setLoading(false);
            }
        };

        fetchShareLink();
    }, [playlistId]);

    const copyToClipboard = async (text: string, type: 'url' | 'embed') => {
        try {
            await navigator.clipboard.writeText(text);
            setCopied(type);
            setTimeout(() => setCopied(null), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    const trackShare = async (platform: TrackShareRequest['platform']) => {
        try {
            await apiClient.post(`/playlists/${playlistId}/track-share`, {
                platform,
                referrer: window.location.href,
            });
        } catch (err) {
            console.error('Failed to track share:', err);
        }
    };

    const shareToSocial = (platform: 'twitter' | 'facebook' | 'discord') => {
        if (!shareData) return;

        trackShare(platform);

        const urls = {
            twitter: `https://twitter.com/intent/tweet?url=${encodeURIComponent(shareData.share_url)}&text=${encodeURIComponent('Check out this playlist!')}`,
            facebook: `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(shareData.share_url)}`,
            discord: shareData.share_url, // Discord handles URL previews automatically
        };

        if (platform === 'discord') {
            copyToClipboard(shareData.share_url, 'url');
        } else {
            window.open(urls[platform], '_blank', 'width=600,height=400');
        }
    };

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-surface rounded-lg max-w-2xl w-full border border-border shadow-xl">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b border-border">
                    <div className="flex items-center gap-2">
                        <Share2 className="h-5 w-5 text-purple-400" />
                        <h2 className="text-xl font-bold text-white">Share Playlist</h2>
                    </div>
                    <button
                        onClick={onClose}
                        className="text-muted-foreground hover:text-white transition-colors"
                    >
                        <X className="h-5 w-5" />
                    </button>
                </div>

                {/* Content */}
                <div className="p-6 space-y-6">
                    {loading && (
                        <div className="text-center py-8">
                            <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-purple-500 border-r-transparent"></div>
                            <p className="mt-4 text-muted-foreground">Generating share link...</p>
                        </div>
                    )}

                    {error && (
                        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
                            <p className="text-red-400">{error}</p>
                        </div>
                    )}

                    {shareData && (
                        <>
                            {/* Share Link */}
                            <div className="space-y-2">
                                <label className="block text-sm font-medium text-muted-foreground">
                                    Share Link
                                </label>
                                <div className="flex gap-2">
                                    <input
                                        type="text"
                                        value={shareData.share_url}
                                        readOnly
                                        className="flex-1 bg-surface-raised text-white px-4 py-2 rounded-lg border border-border focus:outline-none focus:border-purple-500 text-sm"
                                        onClick={(e) => (e.target as HTMLInputElement).select()}
                                    />
                                    <button
                                        onClick={() => {
                                            copyToClipboard(shareData.share_url, 'url');
                                            trackShare('link');
                                        }}
                                        className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg transition-colors flex items-center gap-2 whitespace-nowrap"
                                    >
                                        {copied === 'url' ? (
                                            <>
                                                <Check className="h-4 w-4" />
                                                Copied!
                                            </>
                                        ) : (
                                            <>
                                                <Copy className="h-4 w-4" />
                                                Copy
                                            </>
                                        )}
                                    </button>
                                </div>
                            </div>

                            {/* Social Share Buttons */}
                            <div className="space-y-2">
                                <label className="block text-sm font-medium text-muted-foreground">
                                    Share on Social Media
                                </label>
                                <div className="grid grid-cols-3 gap-2">
                                    <button
                                        onClick={() => shareToSocial('twitter')}
                                        className="px-4 py-3 bg-[#1DA1F2] hover:bg-[#1a8cd8] text-white rounded-lg transition-colors font-medium"
                                    >
                                        Twitter
                                    </button>
                                    <button
                                        onClick={() => shareToSocial('facebook')}
                                        className="px-4 py-3 bg-[#1877F2] hover:bg-[#166fe5] text-white rounded-lg transition-colors font-medium"
                                    >
                                        Facebook
                                    </button>
                                    <button
                                        onClick={() => shareToSocial('discord')}
                                        className="px-4 py-3 bg-[#5865F2] hover:bg-[#4752c4] text-white rounded-lg transition-colors font-medium"
                                    >
                                        Discord
                                    </button>
                                </div>
                                <p className="text-xs text-text-secondary mt-2">
                                    Click Discord to copy the link for pasting into Discord
                                </p>
                            </div>

                            {/* Embed Code */}
                            <div className="space-y-2">
                                <label className="block text-sm font-medium text-muted-foreground">
                                    Embed Code
                                </label>
                                <div className="space-y-2">
                                    <textarea
                                        value={shareData.embed_code}
                                        readOnly
                                        rows={3}
                                        className="w-full bg-surface-raised text-white px-4 py-2 rounded-lg border border-border focus:outline-none focus:border-purple-500 text-xs font-mono resize-none"
                                        onClick={(e) => (e.target as HTMLTextAreaElement).select()}
                                    />
                                    <button
                                        onClick={() => {
                                            copyToClipboard(shareData.embed_code, 'embed');
                                            trackShare('embed');
                                        }}
                                        className="w-full px-4 py-2 bg-surface-raised hover:bg-surface-hover text-white rounded-lg transition-colors flex items-center justify-center gap-2"
                                    >
                                        {copied === 'embed' ? (
                                            <>
                                                <Check className="h-4 w-4" />
                                                Copied Embed Code!
                                            </>
                                        ) : (
                                            <>
                                                <Copy className="h-4 w-4" />
                                                Copy Embed Code
                                            </>
                                        )}
                                    </button>
                                </div>
                            </div>
                        </>
                    )}
                </div>

                {/* Footer */}
                <div className="p-6 border-t border-border">
                    <button
                        onClick={onClose}
                        className="w-full px-4 py-2 bg-surface-raised hover:bg-surface-hover text-white rounded-lg transition-colors"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
}
