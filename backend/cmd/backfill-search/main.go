package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	opensearchpkg "git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
)

func main() {
	batchSize := flag.Int("batch", 100, "Number of records to process in each batch")
	flag.Parse()

	log.Println("Starting search index backfill...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	osClient, err := opensearchpkg.NewClient(&opensearchpkg.Config{
		URL:                cfg.OpenSearch.URL,
		Username:           cfg.OpenSearch.Username,
		Password:           cfg.OpenSearch.Password,
		InsecureSkipVerify: cfg.OpenSearch.InsecureSkipVerify,
	})
	if err != nil {
		log.Fatalf("Failed to initialize OpenSearch client: %v", err)
	}

	ctx := context.Background()
	if err := osClient.Ping(ctx); err != nil {
		log.Fatalf("OpenSearch ping failed: %v", err)
	}
	log.Println("OpenSearch connection established")

	indexer := services.NewSearchIndexerService(osClient)

	log.Println("Initializing search indices...")
	if err := indexer.InitializeIndices(ctx); err != nil {
		log.Fatalf("Failed to initialize indices: %v", err)
	}
	log.Println("Indices initialized successfully")

	totalFailed := 0
	steps := []struct {
		name string
		run  func(context.Context, *database.DB, *services.SearchIndexerService, int) (int, error)
	}{
		{"clips", backfillClips},
		{"users", backfillUsers},
		{"tags", backfillTags},
		{"games", backfillGames},
	}
	for _, step := range steps {
		log.Printf("Backfilling %s...", step.name)
		failed, err := step.run(ctx, db, indexer, *batchSize)
		if err != nil {
			log.Fatalf("Failed to backfill %s: %v", step.name, err)
		}
		totalFailed += failed
	}

	if totalFailed > 0 {
		log.Printf("Backfill finished with %d record(s) that failed to index", totalFailed)
		os.Exit(1)
	}

	log.Println("Backfill completed successfully!")
}

// pageDelay is the pause between batches; tests set it to zero.
var pageDelay = 100 * time.Millisecond

// runPaged calls page with increasing offsets until a page returns no rows.
// Paging stops on rows fetched, not rows indexed, so a batch whose rows all
// fail to index does not end the backfill early. It returns the number of rows
// indexed and the number that failed to index.
func runPaged(batchSize int, page func(offset int) (fetched, indexed int, err error)) (int, int, error) {
	offset := 0
	totalIndexed, totalFailed := 0, 0
	for {
		fetched, indexed, err := page(offset)
		if err != nil {
			return totalIndexed, totalFailed, err
		}
		if fetched == 0 {
			return totalIndexed, totalFailed, nil
		}
		totalIndexed += indexed
		totalFailed += fetched - indexed
		offset += batchSize
		if pageDelay > 0 {
			time.Sleep(pageDelay)
		}
	}
}

func backfillClips(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) (int, error) {
	started := time.Now()
	totalIndexed, failed, err := runPaged(batchSize, func(offset int) (int, int, error) {
		query := `
			SELECT id, twitch_clip_id, twitch_clip_url, embed_url, title, 
			       creator_name, creator_id, broadcaster_name, broadcaster_id,
			       game_id, game_name, language, thumbnail_url, duration,
			       view_count, created_at, imported_at, vote_score,
			       comment_count, favorite_count, is_featured, is_nsfw,
			       is_removed, removed_reason, submitted_by_user_id, is_hidden, dmca_removed,
			       (SELECT t.full_text FROM clip_transcripts t WHERE t.clip_id = clips.id)
			FROM clips
			WHERE is_removed = false
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch clips: %w", err)
		}

		var clips []models.Clip
		for rows.Next() {
			var clip models.Clip
			err := rows.Scan(
				&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
				&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
				&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
				&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
				&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
				&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
				&clip.SubmittedByUserID, &clip.IsHidden, &clip.DMCARemoved, &clip.Transcript,
			)
			if err != nil {
				rows.Close()
				return 0, 0, fmt.Errorf("failed to scan clip: %w", err)
			}
			clips = append(clips, clip)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return 0, 0, fmt.Errorf("failed to iterate clips: %w", err)
		}

		if len(clips) == 0 {
			return 0, 0, nil
		}

		if err := indexer.BulkIndexClips(ctx, clips); err != nil {
			return 0, 0, fmt.Errorf("failed to index clips batch: %w", err)
		}

		log.Printf("Indexed %d clips (offset: %d)", len(clips), offset)
		return len(clips), len(clips), nil
	})
	if err != nil {
		return failed, err
	}

	log.Printf("Completed indexing %d clips", totalIndexed)
	// Transcripts written during the run may have been overwritten with the
	// older text read earlier; send them again (#87).
	if err := repository.RequeueTranscriptsSince(ctx, db.Pool, started); err != nil {
		log.Printf("WARNING: Failed to requeue transcripts written during the backfill: %v", err)
	}
	return failed, nil
}

func backfillUsers(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) (int, error) {
	totalIndexed, failed, err := runPaged(batchSize, func(offset int) (int, int, error) {
		query := `
			SELECT id, twitch_id, username, COALESCE(display_name, username), email, avatar_url,
			       bio, karma_points, role, is_banned, created_at, updated_at, last_login_at
			FROM users
			WHERE is_banned = false
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch users: %w", err)
		}

		fetched, count := 0, 0
		for rows.Next() {
			fetched++
			var user models.User
			err := rows.Scan(
				&user.ID, &user.TwitchID, &user.Username, &user.DisplayName,
				&user.Email, &user.AvatarURL, &user.Bio, &user.KarmaPoints,
				&user.Role, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt,
				&user.LastLoginAt,
			)
			if err != nil {
				rows.Close()
				return 0, 0, fmt.Errorf("failed to scan user: %w", err)
			}

			if err := indexer.IndexUser(ctx, &user); err != nil {
				log.Printf("WARNING: Failed to index user %s: %v", user.ID, err)
			} else {
				count++
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return 0, 0, fmt.Errorf("failed to iterate users: %w", err)
		}

		if fetched > 0 {
			log.Printf("Indexed %d of %d users (offset: %d)", count, fetched, offset)
		}
		return fetched, count, nil
	})
	if err != nil {
		return failed, err
	}

	log.Printf("Completed indexing %d users, %d failed", totalIndexed, failed)
	return failed, nil
}

func backfillTags(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) (int, error) {
	totalIndexed, failed, err := runPaged(batchSize, func(offset int) (int, int, error) {
		query := `
			SELECT id, name, slug, description, color, usage_count, created_at
			FROM tags
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch tags: %w", err)
		}

		fetched, count := 0, 0
		for rows.Next() {
			fetched++
			var tag models.Tag
			err := rows.Scan(
				&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
				&tag.Color, &tag.UsageCount, &tag.CreatedAt,
			)
			if err != nil {
				rows.Close()
				return 0, 0, fmt.Errorf("failed to scan tag: %w", err)
			}

			if err := indexer.IndexTag(ctx, &tag); err != nil {
				log.Printf("WARNING: Failed to index tag %s: %v", tag.ID, err)
			} else {
				count++
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return 0, 0, fmt.Errorf("failed to iterate tags: %w", err)
		}

		if fetched > 0 {
			log.Printf("Indexed %d of %d tags (offset: %d)", count, fetched, offset)
		}
		return fetched, count, nil
	})
	if err != nil {
		return failed, err
	}

	log.Printf("Completed indexing %d tags, %d failed", totalIndexed, failed)
	return failed, nil
}

func backfillGames(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) (int, error) {
	totalIndexed, failed, err := runPaged(batchSize, func(offset int) (int, int, error) {
		query := `
			SELECT c.game_id, COALESCE(c.game_name,g.name), COUNT(*) as clip_count
			FROM clips c LEFT JOIN games g ON g.twitch_game_id=c.game_id
			WHERE c.game_id IS NOT NULL AND COALESCE(c.game_name,g.name) IS NOT NULL
			  AND NOT c.is_removed AND NOT c.is_hidden AND NOT c.dmca_removed
			GROUP BY c.game_id, COALESCE(c.game_name,g.name)
			ORDER BY c.game_id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch games: %w", err)
		}

		fetched, count := 0, 0
		for rows.Next() {
			fetched++
			var game models.GameSearchResult
			err := rows.Scan(&game.ID, &game.Name, &game.ClipCount)
			if err != nil {
				rows.Close()
				return 0, 0, fmt.Errorf("failed to scan game: %w", err)
			}

			if err := indexer.IndexGameSearchResult(ctx, &game); err != nil {
				log.Printf("WARNING: Failed to index game %s: %v", game.ID, err)
			} else {
				count++
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return 0, 0, fmt.Errorf("failed to iterate games: %w", err)
		}

		if fetched > 0 {
			log.Printf("Indexed %d of %d games (offset: %d)", count, fetched, offset)
		}
		return fetched, count, nil
	})
	if err != nil {
		return failed, err
	}

	log.Printf("Completed indexing %d games, %d failed", totalIndexed, failed)
	return failed, nil
}
