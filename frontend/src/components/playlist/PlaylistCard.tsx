import apiClient from "@/lib/api";

import { Badge } from "@/components/ui";
import { formatRelativeTimestamp, cn } from "@/lib/utils";
import type { Playlist, PlaylistWithClips } from "@/types/playlist";
import {
	useBookmarkPlaylist,
	useLikePlaylist,
	useUnlikePlaylist,
	useUnbookmarkPlaylist,
} from "@/hooks/usePlaylist";
import { useIsAuthenticated, useToast } from "@/hooks";
import { Link } from "react-router-dom";
import {
	Heart,
	Lock,
	Users,
	Globe,
	Bookmark,
	MessageSquare,
} from "lucide-react";
import { ShareButton } from "../clip/ShareButton";
import { PlaylistThumbnail } from "./PlaylistThumbnail";

interface PlaylistCardProps {
	playlist: (Playlist & { clip_count?: number }) | PlaylistWithClips;
}

export function PlaylistCard({ playlist }: PlaylistCardProps) {
	const statusBadges: Array<{ label: string; className: string }> = [];
	const visibilityIconClass = "h-2.5 w-2.5";
	const createdAt = formatRelativeTimestamp(playlist.created_at);
	const isAuthenticated = useIsAuthenticated();
	const toast = useToast();
	const likeMutation = useLikePlaylist();
	const unlikeMutation = useUnlikePlaylist();
	const bookmarkMutation = useBookmarkPlaylist();
	const unbookmarkMutation = useUnbookmarkPlaylist();
	const isLiked = Boolean(playlist.is_liked);
	const likeCount = playlist.like_count;
	const isBookmarked = Boolean(playlist.is_bookmarked);
	const bookmarkCount = playlist.bookmark_count;

	if (playlist.has_processing_clips) {
		statusBadges.push({
			label: "Processing…",
			className:
				"shrink-0 border border-amber-400/40 bg-amber-500/12 text-[11px] text-amber-100 shadow-xs",
		});
	}

	if (playlist.script_id) {
		statusBadges.push({
			label: "Scripted",
			className:
				"shrink-0 border border-violet-400/40 bg-violet-500/12 text-[11px] text-violet-100 shadow-xs",
		});
	}

	const getVisibilityIcon = () => {
		switch (playlist.visibility) {
			case "private":
				return <Lock className={visibilityIconClass} />;
			case "public":
				return <Globe className={visibilityIconClass} />;
			case "unlisted":
				return <Users className={visibilityIconClass} />;
			default:
				return null;
		}
	};

	const getVisibilityLabel = () => {
		switch (playlist.visibility) {
			case "private":
				return "Private";
			case "public":
				return "Public";
			case "unlisted":
				return "Unlisted";
			default:
				return "";
		}
	};

	const trackShare = async (
		platform: "link" | "twitter" | "facebook" | "reddit" | "bluesky",
	) => {
		const trackedPlatform =
			platform === "twitter" || platform === "facebook" ? platform : "link";

		try {
			await apiClient.post(`/playlists/${playlist.id}/track-share`, {
				platform: trackedPlatform,
				referrer: window.location.href,
			});
		} catch (error) {
			console.error("Failed to track share:", error);
		}
	};

	const handleLikeClick = async (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();

		if (!isAuthenticated) {
			toast.info("Please log in to like playlists");
			return;
		}

		try {
			if (isLiked) {
				await unlikeMutation.mutateAsync(playlist.id);
				toast.success("Playlist unliked");
			} else {
				await likeMutation.mutateAsync(playlist.id);
				toast.success("Playlist liked");
			}
		} catch {
			toast.error("Failed to update playlist like");
		}
	};

	const handleBookmarkClick = async (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();

		if (!isAuthenticated) {
			toast.info("Please log in to bookmark playlists");
			return;
		}

		try {
			if (isBookmarked) {
				await unbookmarkMutation.mutateAsync(playlist.id);
				toast.success("Bookmark removed", {
					action: {
						label: "Undo",
						onClick: () => bookmarkMutation.mutate(playlist.id),
					},
				});
			} else {
				await bookmarkMutation.mutateAsync(playlist.id);
				toast.success("Playlist bookmarked");
			}
		} catch {
			toast.error("Failed to update playlist bookmark");
		}
	};

	const clipsData =
		"clips" in playlist &&
		Array.isArray(playlist.clips) &&
		playlist.clips.length > 0
			? playlist.clips
			: "preview_clips" in playlist &&
			  Array.isArray(playlist.preview_clips) &&
			  playlist.preview_clips.length > 0
			? playlist.preview_clips
			: null;
	const previewClips = clipsData ? clipsData.slice(0, 4) : [];
	const hasClips = previewClips.length > 0;
	const getPreviewTileClass = (count: number, idx: number) =>
		cn(
			"h-full w-full rounded-lg overflow-hidden bg-accent",
			count === 1 && "col-span-2 row-span-2",
			count === 2 && "row-span-2",
			count === 3 && idx === 0 && "col-span-2",
		);

	return (
		<>
			<Link
				to={`/playlists/${playlist.id}`}
				className="group flex h-full flex-col rounded-xl border border-border bg-card p-3 transition-all hover:border-primary-500/30 hover:shadow-lg"
			>
				{/* Mosaic or Cover Image */}
				{hasClips && previewClips.length > 0 ? (
					<div className="relative aspect-video shrink-0 overflow-hidden rounded-lg mb-3">
						<div className="grid h-full grid-cols-2 grid-rows-2 gap-1">
							{previewClips.map((clip, idx) => (
								<div
									key={clip.id}
									className={getPreviewTileClass(previewClips.length, idx)}
								>
									{clip.thumbnail_url ? (
										<img
											src={clip.thumbnail_url}
											alt={clip.title}
											className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
											loading="lazy"
										/>
									) : (
										<div className="w-full h-full bg-linear-to-br from-accent to-accent/50" />
									)}
								</div>
							))}
						</div>
					</div>
				) : (
					<div className="relative aspect-video shrink-0 overflow-hidden rounded-lg mb-3">
						{playlist.cover_url ? (
							<img
								src={playlist.cover_url}
								alt={playlist.title}
								className="w-full h-full object-cover"
							/>
						) : (
							<PlaylistThumbnail
								clips={"clips" in playlist ? playlist.clips : undefined}
								className="w-full h-full"
							/>
						)}
					</div>
				)}

				{/* Title */}
				<h3 className="text-base font-semibold text-foreground mb-1 line-clamp-1 group-hover:text-primary-500 transition-colors leading-snug">
					{playlist.title}
				</h3>

				{/* Description */}
				<div className="mb-2 min-h-10">
					{playlist.description && (
						<p className="text-sm text-muted-foreground line-clamp-2">
							{playlist.description}
						</p>
					)}
				</div>

				{/* Comment count + top comment preview */}
				{(playlist.comment_count ?? 0) > 0 && (
					<div className="mb-2 space-y-1">
						<div className="flex items-center gap-1.5 text-text-secondary text-[12px]">
							<MessageSquare size={14} />
							<span>{playlist.comment_count} comments</span>
						</div>
						{playlist.top_comment && (
							<div>
								<p className="text-[12px] text-text-secondary italic line-clamp-1">
									&ldquo;{playlist.top_comment.content}&rdquo;
								</p>
								<p className="text-[11px] text-text-tertiary">
									&mdash; @{playlist.top_comment.username}
								</p>
							</div>
						)}
					</div>
				)}

				{/* Bottom section — pushed to bottom, aligned with thumbnail edges */}
				<div className="mt-4 space-y-2">
					<div className="-mb-1.5 flex min-h-6 items-start justify-between gap-2">
						<div className="flex min-w-0 flex-wrap items-start gap-1.5">
							{statusBadges.map((badge) => (
								<Badge
									key={badge.label}
									variant="secondary"
									size="sm"
									className={badge.className}
								>
									{badge.label}
								</Badge>
							))}
						</div>

						<Badge
							variant="secondary"
							size="sm"
							className="shrink-0 border border-border bg-background/90 text-[11px] text-foreground shadow-xs"
						>
							{getVisibilityIcon()}
							<span>{getVisibilityLabel()}</span>
						</Badge>
					</div>

					<div className="-mb-3 flex flex-no-wrap items-center justify-between gap-1.5 text-xs text-muted-foreground">
						<div className="flex flex-no-wrap items-center gap-x-1.5 gap-y-1">
							{playlist.clip_count !== undefined && (
								<span className="inline-flex items-center gap-1.5">
									<span className="font-medium text-foreground/90">
										{playlist.clip_count}
									</span>
									<span>clips</span>
								</span>
							)}

							<button
								type="button"
								onClick={handleLikeClick}
								disabled={likeMutation.isPending || unlikeMutation.isPending}
								className="inline-flex items-center gap-1 rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500 cursor-pointer disabled:cursor-not-allowed disabled:opacity-60"
								title={
									isAuthenticated
										? isLiked
											? "Unlike playlist"
											: "Like playlist"
										: "Log in to like playlists"
								}
							>
								<Heart
									size={14}
									className={cn(isLiked && "fill-current text-primary-500")}
								/>
								<span className="font-medium text-foreground/90">
									{likeCount}
								</span>
							</button>

							<button
								type="button"
								onClick={handleBookmarkClick}
								disabled={
									bookmarkMutation.isPending || unbookmarkMutation.isPending
								}
								className="inline-flex items-center gap-1 rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500 cursor-pointer disabled:cursor-not-allowed disabled:opacity-60"
								title={
									isAuthenticated
										? isBookmarked
											? "Remove bookmark"
											: "Bookmark playlist"
										: "Log in to bookmark playlists"
								}
							>
								<Bookmark
									size={14}
									className={cn(isBookmarked && "fill-current text-primary-500")}
								/>
								<span className="font-medium text-foreground/90">
									{bookmarkCount}
								</span>
							</button>

							{playlist.visibility !== "private" && (
								<ShareButton
									shareUrl={`${window.location.origin}/playlists/${playlist.id}`}
									shareTitle={
										playlist.title ||
										playlist.description ||
										"Check out this playlist!"
									}
									onShare={trackShare}
									showLabel={false}
									buttonClassName="inline-flex min-h-0 items-center rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500"
									iconClassName="h-3.5 w-3.5"
									preventLinkNavigation={true}
								/>
							)}
						</div>

						<div
							className="shrink-0 whitespace-nowrap text-right text-xs text-muted-foreground"
							title={createdAt.title}
						>
							{createdAt.display}
						</div>
					</div>
				</div>
			</Link>
		</>
	);
}
