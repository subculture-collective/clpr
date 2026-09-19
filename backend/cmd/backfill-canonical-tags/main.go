package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type stats struct{ examined, migrated, ambiguous, skipped, failed int }

func main() {
	dryRun := flag.Bool("dry-run", false, "report work without changing tags")
	batchSize := flag.Int("batch-size", 500, "clips processed per batch")
	afterText := flag.String("after", "00000000-0000-0000-0000-000000000000", "resume after this clip UUID")
	flag.Parse()
	if *batchSize < 1 || *batchSize > 5000 {
		log.Fatal("batch-size must be between 1 and 5000")
	}
	after, err := uuid.Parse(*afterText)
	if err != nil {
		log.Fatalf("invalid --after: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	tagRepo := repository.NewTagRepository(db.Pool)
	clipRepo := repository.NewClipRepository(db.Pool)
	tagger := services.NewAutoTagService(tagRepo)
	result := stats{}

	if err := db.Pool.QueryRow(ctx, ambiguousLegacyTagCountQuery).Scan(&result.ambiguous); err != nil {
		log.Fatalf("counting ambiguous legacy tags: %v", err)
	}
	if !*dryRun {
		if err := ensureKnownLegacyAliases(ctx, db.Pool); err != nil {
			log.Fatalf("legacy aliases: %v", err)
		}
	}
	for {
		rows, err := db.Pool.Query(ctx, `SELECT id,twitch_clip_id,title,broadcaster_name,game_id,game_name,language,duration
			FROM clips WHERE id > $1 ORDER BY id LIMIT $2`, after, *batchSize)
		if err != nil {
			log.Fatal(err)
		}
		clips := []models.Clip{}
		for rows.Next() {
			var c models.Clip
			if err := rows.Scan(&c.ID, &c.TwitchClipID, &c.Title, &c.BroadcasterName, &c.GameID, &c.GameName, &c.Language, &c.Duration); err != nil {
				rows.Close()
				log.Fatal(err)
			}
			clips = append(clips, c)
		}
		rows.Close()
		if len(clips) == 0 {
			break
		}
		for i := range clips {
			clip := &clips[i]
			result.examined++
			after = clip.ID
			if *dryRun {
				if len(services.CanonicalTagSlugsForClip(clip)) == 0 {
					result.skipped++
				} else {
					result.migrated++
				}
				continue
			}
			if err := tagger.ApplyAutoTags(ctx, clip); err != nil {
				result.failed++
				log.Printf("clip %s failed: %v", clip.ID, err)
				continue
			}
			if err := migrateKnownLegacyTagsForClip(ctx, db.Pool, clip.ID); err != nil {
				result.failed++
				log.Printf("clip %s legacy mapping failed: %v", clip.ID, err)
				continue
			}
			if err := migrateLegacyGameTagForClip(ctx, db.Pool, clip); err != nil {
				result.failed++
				log.Printf("clip %s legacy game mapping failed: %v", clip.ID, err)
				continue
			}
			if err := clipRepo.MarkAutoTagged(ctx, clip.ID); err != nil {
				result.failed++
				log.Printf("clip %s checkpoint failed: %v", clip.ID, err)
				continue
			}
			result.migrated++
		}
		log.Printf("checkpoint after=%s examined=%d migrated=%d failed=%d", after, result.examined, result.migrated, result.failed)
	}
	if !*dryRun {
		if err := removeRetiredLegacyTags(ctx, db.Pool); err != nil {
			log.Printf("retired legacy cleanup failed: %v", err)
			result.failed++
		}
	}
	fmt.Printf("examined=%d migrated=%d ambiguous=%d skipped=%d failed=%d last=%s\n", result.examined, result.migrated, result.ambiguous, result.skipped, result.failed, after)
	if result.failed > 0 {
		log.Fatal("backfill completed with failures")
	}
}

const knownLegacyMappingCTE = `mapping(old_slug,new_slug) AS (VALUES
	('ace','content/ace'),('clutch','content/clutch'),('fail','content/fail'),
	('rage','content/rage'),('funny','content/funny'),('insane','content/insane'),
	('lucky','content/lucky'),('bug','content/bug'),('toxic','content/toxic'),
	('epic','content/epic'),('noob','content/noob'),('pro','content/pro'),
	('highlight','content/highlight'),('speedrun','content/speedrun'),('tutorial','content/tutorial'),
	('short','duration/short'),('medium','duration/medium'),('long','duration/long'),
	('english','lang/en'),('spanish','lang/es'),('portuguese','lang/pt'),
	('french','lang/fr'),('german','lang/de'),('russian','lang/ru'),
	('japanese','lang/ja'),('korean','lang/ko'),('chinese','lang/zh'),
	('italian','lang/it'),('turkish','lang/tr'),('arabic','lang/ar')
)`

const ambiguousLegacyTagCountQuery = `WITH ` + knownLegacyMappingCTE + `
SELECT COUNT(*) FROM tags t
WHERE t.parent_slug IS NULL
  AND t.slug NOT IN ('game','duration','lang','content','community')
  AND NOT EXISTS (SELECT 1 FROM mapping m WHERE m.old_slug=t.slug)
  AND NOT EXISTS (
      SELECT 1 FROM clips c
      WHERE trim(both '-' from regexp_replace(lower(COALESCE(c.game_name,'')), '[^a-z0-9]+', '-', 'g'))=t.slug
  )
  AND NOT EXISTS (SELECT 1 FROM tag_aliases a WHERE a.alias_slug=t.slug)`

func ensureKnownLegacyAliases(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		WITH `+knownLegacyMappingCTE+`, resolved AS (
			SELECT m.old_slug,m.new_slug,old.id old_id,canonical.id canonical_id
			FROM mapping m JOIN tags old ON old.slug=m.old_slug JOIN tags canonical ON canonical.slug=m.new_slug
		)
		INSERT INTO tag_aliases(alias_slug,canonical_tag_id)
		SELECT old_slug,canonical_id FROM resolved
		ON CONFLICT(alias_slug) DO UPDATE SET canonical_tag_id=EXCLUDED.canonical_tag_id`)
	return err
}

func migrateKnownLegacyTagsForClip(ctx context.Context, pool *pgxpool.Pool, clipID uuid.UUID) error {
	_, err := pool.Exec(ctx, `WITH `+knownLegacyMappingCTE+`, resolved AS (
		SELECT old.id old_id, canonical.id canonical_id
		FROM mapping m JOIN tags old ON old.slug=m.old_slug JOIN tags canonical ON canonical.slug=m.new_slug
	), copied AS (
		INSERT INTO clip_tags(clip_id,tag_id,created_at)
		SELECT ct.clip_id,r.canonical_id,ct.created_at FROM clip_tags ct JOIN resolved r ON r.old_id=ct.tag_id
		WHERE ct.clip_id=$1 ON CONFLICT DO NOTHING
	)
	DELETE FROM clip_tags ct USING resolved r WHERE ct.clip_id=$1 AND ct.tag_id=r.old_id`, clipID)
	return err
}

func migrateLegacyGameTagForClip(ctx context.Context, pool *pgxpool.Pool, clip *models.Clip) error {
	if clip.GameName == nil {
		return nil
	}
	legacySlug := utils.Slugify(*clip.GameName)
	if legacySlug == "" {
		return nil
	}
	target := ""
	for _, slug := range services.CanonicalTagSlugsForClip(clip) {
		if slug == "game/"+legacySlug {
			target = slug
			break
		}
		if target == "" && len(slug) > len("game/") && slug[:len("game/")] == "game/" {
			target = slug
		}
	}
	if target == "" || target == legacySlug {
		return nil
	}
	_, err := pool.Exec(ctx, `WITH resolved AS (
		SELECT old.id old_id, canonical.id canonical_id
		FROM tags old JOIN tags canonical ON canonical.slug=$2 WHERE old.slug=$1
	), aliased AS (
		INSERT INTO tag_aliases(alias_slug,canonical_tag_id)
		SELECT $1,canonical_id FROM resolved
		ON CONFLICT(alias_slug) DO UPDATE SET canonical_tag_id=EXCLUDED.canonical_tag_id
	)
	DELETE FROM clip_tags ct USING resolved r WHERE ct.clip_id=$3 AND ct.tag_id=r.old_id`, legacySlug, target, clip.ID)
	return err
}

func removeRetiredLegacyTags(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DELETE FROM tags t USING tag_aliases a
		WHERE t.slug=a.alias_slug AND NOT EXISTS (SELECT 1 FROM clip_tags ct WHERE ct.tag_id=t.id)`)
	return err
}
