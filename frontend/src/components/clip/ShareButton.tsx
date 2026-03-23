import { useState, useRef } from 'react';
import { useToast } from '@/hooks';
import { useClickOutside } from '@/hooks/useClickOutside';
import { cn } from '@/lib/utils';
import { Share2 } from 'lucide-react';

interface ShareButtonProps {
    clipId?: string;
    clipTitle?: string;
    shareUrl?: string;
    shareTitle?: string;
    onShare?: (
        platform: 'link' | 'twitter' | 'facebook' | 'reddit' | 'bluesky',
    ) => void | Promise<void>;
    showLabel?: boolean;
    buttonClassName?: string;
    iconClassName?: string;
    dropdownClassName?: string;
    dropdownAlign?: 'left' | 'right';
    preventLinkNavigation?: boolean;
}

export function ShareButton({
    clipId,
    clipTitle,
    shareUrl,
    shareTitle,
    onShare,
    showLabel = true,
    buttonClassName,
    iconClassName,
    dropdownClassName,
    dropdownAlign = 'left',
    preventLinkNavigation = false,
}: ShareButtonProps) {
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);
    const toast = useToast();

    const resolvedShareUrl =
        shareUrl ||
        (clipId
            ? `${window.location.origin}/clip/${clipId}`
            : window.location.href);
    const shareText = shareTitle || clipTitle || 'Check this out!';
    const dropdownSideClass = dropdownAlign === 'right' ? 'right-0' : 'left-0';

    // Check if Web Share API is available
    const canNativeShare = typeof navigator.share === 'function';

    const trackShare = (
        platform: 'link' | 'twitter' | 'facebook' | 'reddit' | 'bluesky',
    ) => {
        if (!onShare) {
            return;
        }

        void Promise.resolve(onShare(platform)).catch(() => {
            // Tracking failure shouldn't block sharing.
        });
    };

    const interceptParentClick = (e: React.MouseEvent<HTMLButtonElement>) => {
        if (!preventLinkNavigation) {
            return;
        }

        e.preventDefault();
        e.stopPropagation();
    };

    const handleNativeShare = async () => {
        try {
            await navigator.share({
                title: shareText,
                text: shareText,
                url: resolvedShareUrl,
            });
            trackShare('link');
        } catch (err) {
            // User cancelled or share failed - that's okay
            if ((err as Error).name !== 'AbortError') {
                // Fallback to copy
                void handleCopyLink();
            }
        }
    };

    const handleCopyLink = async () => {
        try {
            await navigator.clipboard.writeText(resolvedShareUrl);
            toast.success('Link copied to clipboard');
            trackShare('link');
            setIsOpen(false);
        } catch {
            toast.error('Failed to copy link');
        }
    };

    const handleShareTwitter = () => {
        const twitterUrl = `https://twitter.com/intent/tweet?text=${encodeURIComponent(shareText)}&url=${encodeURIComponent(resolvedShareUrl)}`;
        window.open(twitterUrl, '_blank', 'width=550,height=420');
        trackShare('twitter');
        setIsOpen(false);
    };

    const handleShareReddit = () => {
        const redditUrl = `https://reddit.com/submit?url=${encodeURIComponent(resolvedShareUrl)}&title=${encodeURIComponent(shareText)}`;
        window.open(redditUrl, '_blank', 'width=800,height=600');
        trackShare('reddit');
        setIsOpen(false);
    };

    const handleShareFacebook = () => {
        const facebookUrl = `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(resolvedShareUrl)}`;
        window.open(facebookUrl, '_blank', 'width=550,height=420');
        trackShare('facebook');
        setIsOpen(false);
    };

    const handleShareBluesky = () => {
        const blueskyUrl = `https://bsky.app/intent/compose?text=${encodeURIComponent(`${shareText} ${resolvedShareUrl}`)}`;
        window.open(blueskyUrl, '_blank', 'width=550,height=420');
        trackShare('bluesky');
        setIsOpen(false);
    };

    const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
        interceptParentClick(e);

        // On mobile/supported browsers, use native share directly
        if (canNativeShare) {
            void handleNativeShare();
        } else {
            setIsOpen(!isOpen);
        }
    };

    // Close dropdown when clicking outside
    useClickOutside(dropdownRef, () => setIsOpen(false), isOpen);

    return (
        <div className="relative" ref={dropdownRef}>
            <button
                className={cn(
                    'text-muted-foreground hover:text-foreground flex items-center gap-1.5 transition-colors touch-target min-h-11 cursor-pointer',
                    buttonClassName,
                )}
                onClick={handleClick}
                aria-label="Share"
                aria-haspopup={!canNativeShare}
                aria-expanded={isOpen}
            >
                <Share2 size={18} className={cn('shrink-0', iconClassName)} strokeWidth={1.75} />
                {showLabel && <span>Share</span>}
            </button>

            {/* Dropdown for desktop browsers without Web Share API */}
            {isOpen && !canNativeShare && (
                <div
                    className={cn(
                        'absolute bottom-full mb-2 w-48 overflow-hidden rounded-lg border border-border bg-card shadow-lg z-50',
                        dropdownSideClass,
                        dropdownClassName,
                    )}
                >
                    <div className="py-1">
                        <button
                            onClick={(e) => {
                                interceptParentClick(e);
                                void handleCopyLink();
                            }}
                            className="w-full text-left px-4 py-2.5 text-sm hover:bg-muted transition-colors flex items-center gap-3 cursor-pointer"
                        >
                            <svg
                                className="w-4 h-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                                />
                            </svg>
                            Copy Link
                        </button>

                        <div className="border-t border-border my-1" />

                        <button
                            onClick={(e) => {
                                interceptParentClick(e);
                                handleShareTwitter();
                            }}
                            className="w-full text-left px-4 py-2.5 text-sm hover:bg-muted transition-colors flex items-center gap-3 cursor-pointer"
                        >
                            <svg
                                className="w-4 h-4"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                            >
                                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
                            </svg>
                            Share on X
                        </button>

                        <button
                            onClick={(e) => {
                                interceptParentClick(e);
                                handleShareBluesky();
                            }}
                            className="w-full text-left px-4 py-2.5 text-sm hover:bg-muted transition-colors flex items-center gap-3 cursor-pointer"
                        >
                            <svg
                                className="w-4 h-4"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                            >
                                <path d="M12 10.8c-1.087-2.114-4.046-6.053-6.798-7.995C2.566.944 1.561 1.266.902 1.565.139 1.908 0 3.08 0 3.768c0 .69.378 5.65.624 6.479.815 2.736 3.713 3.66 6.383 3.364.136-.02.275-.039.415-.056-.138.022-.276.04-.415.056-3.912.58-7.387 2.005-2.83 7.078 5.013 5.19 6.87-1.113 7.823-4.308.953 3.195 2.05 9.271 7.733 4.308 4.267-4.308 1.172-6.498-2.74-7.078a8.741 8.741 0 01-.415-.056c.14.017.279.036.415.056 2.67.297 5.568-.628 6.383-3.364.246-.828.624-5.79.624-6.478 0-.69-.139-1.861-.902-2.206-.659-.298-1.664-.62-4.3 1.24C16.046 4.748 13.087 8.687 12 10.8z" />
                            </svg>
                            Share on Bluesky
                        </button>

                        <button
                            onClick={(e) => {
                                interceptParentClick(e);
                                handleShareReddit();
                            }}
                            className="w-full text-left px-4 py-2.5 text-sm hover:bg-muted transition-colors flex items-center gap-3 cursor-pointer"
                        >
                            <svg
                                className="w-4 h-4"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                            >
                                <path d="M12 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0zm5.01 4.744c.688 0 1.25.561 1.25 1.249a1.25 1.25 0 0 1-2.498.056l-2.597-.547-.8 3.747c1.824.07 3.48.632 4.674 1.488.308-.309.73-.491 1.207-.491.968 0 1.754.786 1.754 1.754 0 .716-.435 1.333-1.01 1.614a3.111 3.111 0 0 1 .042.52c0 2.694-3.13 4.87-7.004 4.87-3.874 0-7.004-2.176-7.004-4.87 0-.183.015-.366.043-.534A1.748 1.748 0 0 1 4.028 12c0-.968.786-1.754 1.754-1.754.463 0 .898.196 1.207.49 1.207-.883 2.878-1.43 4.744-1.487l.885-4.182a.342.342 0 0 1 .14-.197.35.35 0 0 1 .238-.042l2.906.617a1.214 1.214 0 0 1 1.108-.701zM9.25 12C8.561 12 8 12.562 8 13.25c0 .687.561 1.248 1.25 1.248.687 0 1.248-.561 1.248-1.249 0-.688-.561-1.249-1.249-1.249zm5.5 0c-.687 0-1.248.561-1.248 1.25 0 .687.561 1.248 1.249 1.248.688 0 1.249-.561 1.249-1.249 0-.687-.562-1.249-1.25-1.249zm-5.466 3.99a.327.327 0 0 0-.231.094.33.33 0 0 0 0 .463c.842.842 2.484.913 2.961.913.477 0 2.105-.056 2.961-.913a.361.361 0 0 0 .029-.463.33.33 0 0 0-.464 0c-.547.533-1.684.73-2.512.73-.828 0-1.979-.196-2.512-.73a.326.326 0 0 0-.232-.095z" />
                            </svg>
                            Share on Reddit
                        </button>

                        <button
                            onClick={(e) => {
                                interceptParentClick(e);
                                handleShareFacebook();
                            }}
                            className="w-full text-left px-4 py-2.5 text-sm hover:bg-muted transition-colors flex items-center gap-3 cursor-pointer"
                        >
                            <svg
                                className="w-4 h-4"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                            >
                                <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
                            </svg>
                            Share on Facebook
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
