import { Link } from "react-router-dom";
import { ArrowUp, ArrowDown, MessageSquare, Heart, Eye } from "lucide-react";
import { Badge } from "@/components/ui";
import { ShareButton } from "./ShareButton";
import { useClipVote, useClipFavorite } from "@/hooks/useClips";
import { useIsAuthenticated, useToast } from "@/hooks";
import { cn, formatCompactNumber, formatDuration, formatTimestamp } from "@/lib/utils";
import type { Clip } from "@/types/clip";

interface ClipGridCardProps {
	clip: Clip;
}

export function ClipGridCard({ clip }: ClipGridCardProps) {
	const isAuthenticated = useIsAuthenticated();
	const voteMutation = useClipVote();
	const favoriteMutation = useClipFavorite();
	const toast = useToast();
	const timestamp = formatTimestamp(clip.created_at);

	const handleVote = (e: React.MouseEvent, voteType: 1 | -1) => {
		e.preventDefault();
		e.stopPropagation();
		if (voteMutation.isPending || clip.user_vote === voteType) return;
		if (!isAuthenticated) {
			toast.info("Please log in to vote on clips");
			return;
		}
		voteMutation.mutate({ clip_id: clip.id, vote_type: voteType });
	};

	const handleFavorite = (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();
		if (!isAuthenticated) {
			toast.info("Please log in to favorite clips");
			return;
		}
		favoriteMutation.mutate({ clip_id: clip.id });
	};

	return (
		<article
			className="group flex h-full min-w-0 flex-col border border-border bg-card p-3 transition-colors motion-reduce:transition-none hover:border-line-strong focus-within:border-primary-400"
			data-testid="clip-grid-card"
		>
			<Link
				to={`/clip/${clip.id}`}
				className="focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
				aria-label={`Watch ${clip.title}`}
			>
			{/* Thumbnail */}
			<div className="relative aspect-video shrink-0 overflow-hidden mb-3">
				{clip.thumbnail_url ? (
					<img
						src={clip.thumbnail_url}
						alt={clip.title}
						width={640}
						height={360}
						className="h-full w-full object-cover transition-opacity duration-150 motion-reduce:transition-none group-hover:opacity-90"
						loading="lazy"
					/>
				) : (
					<div className="flex h-full w-full items-center justify-center bg-accent">
						<Eye
							size={32}
							strokeWidth={1.5}
							className="text-muted-foreground"
						/>
					</div>
				)}

				{/* Duration badge */}
				{clip.duration && (
					<span className="burn-in absolute bottom-2 right-2">
						{formatDuration(clip.duration)}
					</span>
				)}

				{/* NSFW badge */}
				{clip.is_nsfw && (
					<span className="absolute left-2 top-2">
						<Badge variant="error" size="sm">
							NSFW
						</Badge>
					</span>
				)}

				{/* Featured badge */}
				{clip.is_featured && (
					<span className="absolute right-2 top-2">
						<Badge
							variant="secondary"
							size="sm"
							className="shrink-0 border border-primary-700 bg-background/80 text-[11px] text-primary-200"
						>
							Featured
						</Badge>
					</span>
				)}

				{/* Watch progress */}
				{clip.watch_progress && (
					<div
						className="absolute bottom-0 left-0 right-0 h-1 bg-black/30"
						role="progressbar"
						aria-label="Watch progress"
						aria-valuemin={0}
						aria-valuemax={100}
						aria-valuenow={Math.min(100, Math.max(0, clip.watch_progress.progress_percent))}
					>
						<div
							className="h-full bg-tally"
							style={{
								width: `${Math.min(
									100,
									Math.max(0, clip.watch_progress.progress_percent),
								)}%`,
							}}
						/>
					</div>
				)}
			</div>

			{/* Title */}
			<h3 className="font-heading text-xl font-bold uppercase leading-none text-foreground mb-1.5 line-clamp-2 group-hover:text-link transition-colors">
				{clip.title}
			</h3>

			{/* Metadata */}
			<div className="mb-2 font-mono text-[11px] uppercase tracking-[0.04em] text-muted-foreground line-clamp-1">
				<span className="font-medium">{clip.broadcaster_name}</span>
				{clip.game_name && (
					<>
						<span className="mx-1">·</span>
						<span>{clip.game_name}</span>
					</>
				)}
			</div>

			{/* Comment count */}
			{clip.comment_count > 0 && (
				<div className="mb-2 flex items-center gap-1.5 text-text-secondary text-xs">
					<MessageSquare size={14} />
					<span>{clip.comment_count} comments</span>
				</div>
			)}
			</Link>

			{/* Stats — pushed to bottom; views and date wrap below the actions on narrow cards */}
			<div className="-mb-3 mt-auto flex flex-wrap items-center justify-between gap-x-1.5 font-mono text-xs text-muted-foreground">
				<div className="flex min-w-0 items-center gap-x-1.5">
					{/* Vote buttons */}
					<button
						type="button"
						onClick={(e) => handleVote(e, 1)}
						disabled={!isAuthenticated || voteMutation.isPending}
						className={cn(
							"inline-flex min-h-[44px] min-w-[44px] items-center justify-center gap-1 rounded px-2 transition-colors motion-reduce:transition-none hover:bg-accent hover:text-link focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 cursor-pointer disabled:cursor-not-allowed disabled:opacity-60",
							clip.user_vote === 1 && "text-link",
						)}
						title={isAuthenticated ? "Upvote" : "Log in to vote"}
						aria-label={isAuthenticated ? `Upvote ${clip.title}` : "Log in to vote"}
						aria-pressed={clip.user_vote === 1}
					>
						<ArrowUp
							size={14}
							className={cn(clip.user_vote === 1 && "fill-current")}
						/>
						<span className="font-medium text-foreground/90">
							{formatCompactNumber(clip.vote_score)}
						</span>
					</button>

					<button
						type="button"
						onClick={(e) => handleVote(e, -1)}
						disabled={!isAuthenticated || voteMutation.isPending}
						className={cn(
							"inline-flex min-h-[44px] min-w-[44px] items-center justify-center rounded px-2 transition-colors motion-reduce:transition-none hover:bg-accent hover:text-link focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 cursor-pointer disabled:cursor-not-allowed disabled:opacity-60",
							clip.user_vote === -1 && "text-link",
						)}
						title={isAuthenticated ? "Downvote" : "Log in to vote"}
						aria-label={isAuthenticated ? `Downvote ${clip.title}` : "Log in to vote"}
						aria-pressed={clip.user_vote === -1}
					>
						<ArrowDown
							size={14}
							className={cn(clip.user_vote === -1 && "fill-current")}
						/>
					</button>

					{/* Favorite */}
					<button
						type="button"
						onClick={handleFavorite}
						disabled={!isAuthenticated}
						className={cn(
							"inline-flex min-h-[44px] min-w-[44px] items-center justify-center gap-1 rounded px-2 transition-colors motion-reduce:transition-none hover:bg-accent hover:text-link focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 cursor-pointer disabled:cursor-not-allowed disabled:opacity-60",
							clip.is_favorited && "text-link",
						)}
						title={
							isAuthenticated
								? clip.is_favorited
									? "Remove from favorites"
									: "Add to favorites"
								: "Log in to favorite clips"
						}
						aria-label={clip.is_favorited ? `Remove ${clip.title} from favorites` : `Add ${clip.title} to favorites`}
						aria-pressed={clip.is_favorited}
					>
						<Heart
							size={14}
							className={cn(clip.is_favorited && "fill-current text-link")}
						/>
						<span className="font-medium text-foreground/90">
							{formatCompactNumber(clip.favorite_count)}
						</span>
					</button>

					{/* Share */}
					<ShareButton
						clipId={clip.id}
						clipTitle={clip.title}
						showLabel={false}
						buttonClassName="inline-flex min-h-0 items-center rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-link"
						iconClassName="h-3.5 w-3.5"
						preventLinkNavigation={true}
					/>
				</div>

				<div className="ml-auto flex min-h-[44px] shrink-0 items-center gap-2 whitespace-nowrap text-right">
					{/* Views: clpr's last Twitch sync, not a live count */}
					<span
						className="inline-flex items-center gap-1"
						title={`${clip.view_count.toLocaleString()} Twitch views at last sync`}
					>
						<Eye size={14} aria-hidden="true" />
						<span>{formatCompactNumber(clip.view_count)}</span>
						<span className="sr-only">views</span>
					</span>
					<span title={timestamp.title}>{timestamp.display}</span>
				</div>
			</div>
		</article>
	);
}
