package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
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

	log.Println("Backfilling clips...")
	if err := backfillClips(ctx, db, indexer, *batchSize); err != nil {
		log.Fatalf("Failed to backfill clips: %v", err)
	}

	log.Println("Backfilling users...")
	if err := backfillUsers(ctx, db, indexer, *batchSize); err != nil {
		log.Fatalf("Failed to backfill users: %v", err)
	}

	log.Println("Backfilling tags...")
	if err := backfillTags(ctx, db, indexer, *batchSize); err != nil {
		log.Fatalf("Failed to backfill tags: %v", err)
	}

	log.Println("Backfilling games...")
	if err := backfillGames(ctx, db, indexer, *batchSize); err != nil {
		log.Fatalf("Failed to backfill games: %v", err)
	}

	log.Println("Backfill completed successfully!")
}

func backfillClips(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) error {
	offset := 0
	totalIndexed := 0

	for {
		query := `
			SELECT id, twitch_clip_id, twitch_clip_url, embed_url, title, 
			       creator_name, creator_id, broadcaster_name, broadcaster_id,
			       game_id, game_name, language, thumbnail_url, duration,
			       view_count, created_at, imported_at, vote_score,
			       comment_count, favorite_count, is_featured, is_nsfw,
			       is_removed, removed_reason
			FROM clips
			WHERE is_removed = false
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch clips: %w", err)
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
			)
			if err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan clip: %w", err)
			}
			clips = append(clips, clip)
		}
		rows.Close()

		if len(clips) == 0 {
			break
		}

		if err := indexer.BulkIndexClips(ctx, clips); err != nil {
			return fmt.Errorf("failed to index clips batch: %w", err)
		}

		totalIndexed += len(clips)
		log.Printf("Indexed %d clips (total: %d)", len(clips), totalIndexed)

		offset += batchSize
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Completed indexing %d clips", totalIndexed)
	return nil
}

func backfillUsers(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) error {
	offset := 0
	totalIndexed := 0

	for {
		query := `
			SELECT id, twitch_id, username, display_name, email, avatar_url,
			       bio, karma_points, role, is_banned, created_at, updated_at, last_login_at
			FROM users
			WHERE is_banned = false
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch users: %w", err)
		}

		count := 0
		for rows.Next() {
			var user models.User
			err := rows.Scan(
				&user.ID, &user.TwitchID, &user.Username, &user.DisplayName,
				&user.Email, &user.AvatarURL, &user.Bio, &user.KarmaPoints,
				&user.Role, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt,
				&user.LastLoginAt,
			)
			if err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan user: %w", err)
			}

			if err := indexer.IndexUser(ctx, &user); err != nil {
				log.Printf("WARNING: Failed to index user %s: %v", user.ID, err)
			} else {
				count++
			}
		}
		rows.Close()

		if count == 0 {
			break
		}

		totalIndexed += count
		log.Printf("Indexed %d users (total: %d)", count, totalIndexed)

		offset += batchSize
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Completed indexing %d users", totalIndexed)
	return nil
}

func backfillTags(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) error {
	offset := 0
	totalIndexed := 0

	for {
		query := `
			SELECT id, name, slug, description, color, usage_count, created_at
			FROM tags
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch tags: %w", err)
		}

		count := 0
		for rows.Next() {
			var tag models.Tag
			err := rows.Scan(
				&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
				&tag.Color, &tag.UsageCount, &tag.CreatedAt,
			)
			if err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan tag: %w", err)
			}

			if err := indexer.IndexTag(ctx, &tag); err != nil {
				log.Printf("WARNING: Failed to index tag %s: %v", tag.ID, err)
			} else {
				count++
			}
		}
		rows.Close()

		if count == 0 {
			break
		}

		totalIndexed += count
		log.Printf("Indexed %d tags (total: %d)", count, totalIndexed)

		offset += batchSize
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Completed indexing %d tags", totalIndexed)
	return nil
}

func backfillGames(ctx context.Context, db *database.DB, indexer *services.SearchIndexerService, batchSize int) error {
	offset := 0
	totalIndexed := 0

	for {
		query := `
			SELECT game_id, game_name, COUNT(*) as clip_count
			FROM clips
			WHERE game_id IS NOT NULL AND game_name IS NOT NULL AND is_removed = false
			GROUP BY game_id, game_name
			ORDER BY game_id
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch games: %w", err)
		}

		count := 0
		for rows.Next() {
			var game models.GameSearchResult
			err := rows.Scan(&game.ID, &game.Name, &game.ClipCount)
			if err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan game: %w", err)
			}

			if err := indexer.IndexGameSearchResult(ctx, &game); err != nil {
				log.Printf("WARNING: Failed to index game %s: %v", game.ID, err)
			} else {
				count++
			}
		}
		rows.Close()

		if count == 0 {
			break
		}

		totalIndexed += count
		log.Printf("Indexed %d games (total: %d)", count, totalIndexed)

		offset += batchSize
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Completed indexing %d games", totalIndexed)
	return nil
}
