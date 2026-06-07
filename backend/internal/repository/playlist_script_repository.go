package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// playlistScriptColumns is the full column list for playlist_scripts queries
const playlistScriptColumns = `id, name, description, sort, timeframe, clip_limit, visibility,
               is_active, schedule, strategy, game_id, game_ids, broadcaster_id, tag, exclude_tags,
               language, min_vote_score, min_view_count, exclude_nsfw, top_10k_streamers,
               seed_clip_id, retention_days, title_template,
               created_by, created_at, updated_at, last_run_at, last_generated_playlist_id`

// PlaylistScriptRepository handles database operations for playlist scripts
type PlaylistScriptRepository struct {
    pool *pgxpool.Pool
}

// NewPlaylistScriptRepository creates a new PlaylistScriptRepository
func NewPlaylistScriptRepository(pool *pgxpool.Pool) *PlaylistScriptRepository {
    return &PlaylistScriptRepository{pool: pool}
}

// scanPlaylistScript scans a row into a PlaylistScript struct
func scanPlaylistScript(row pgx.Row) (*models.PlaylistScript, error) {
    var script models.PlaylistScript
    err := row.Scan(
        &script.ID,
        &script.Name,
        &script.Description,
        &script.Sort,
        &script.Timeframe,
        &script.ClipLimit,
        &script.Visibility,
        &script.IsActive,
        &script.Schedule,
        &script.Strategy,
        &script.GameID,
        &script.GameIDs,
        &script.BroadcasterID,
        &script.Tag,
        &script.ExcludeTags,
        &script.Language,
        &script.MinVoteScore,
        &script.MinViewCount,
        &script.ExcludeNSFW,
        &script.Top10kStreamers,
        &script.SeedClipID,
        &script.RetentionDays,
        &script.TitleTemplate,
        &script.CreatedBy,
        &script.CreatedAt,
        &script.UpdatedAt,
        &script.LastRunAt,
        &script.LastGeneratedPlaylistID,
    )
    return &script, err
}

// scanPlaylistScriptRows scans multiple rows into PlaylistScript slice
func scanPlaylistScriptRows(rows pgx.Rows) ([]*models.PlaylistScript, error) {
    var scripts []*models.PlaylistScript
    for rows.Next() {
        var script models.PlaylistScript
        err := rows.Scan(
            &script.ID,
            &script.Name,
            &script.Description,
            &script.Sort,
            &script.Timeframe,
            &script.ClipLimit,
            &script.Visibility,
            &script.IsActive,
            &script.Schedule,
            &script.Strategy,
            &script.GameID,
            &script.GameIDs,
            &script.BroadcasterID,
            &script.Tag,
            &script.ExcludeTags,
            &script.Language,
            &script.MinVoteScore,
            &script.MinViewCount,
            &script.ExcludeNSFW,
            &script.Top10kStreamers,
            &script.SeedClipID,
            &script.RetentionDays,
            &script.TitleTemplate,
            &script.CreatedBy,
            &script.CreatedAt,
            &script.UpdatedAt,
            &script.LastRunAt,
            &script.LastGeneratedPlaylistID,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan playlist script: %w", err)
        }
        scripts = append(scripts, &script)
    }
    return scripts, nil
}

// List retrieves all playlist scripts
func (r *PlaylistScriptRepository) List(ctx context.Context) ([]*models.PlaylistScript, error) {
    query := fmt.Sprintf(`SELECT %s FROM playlist_scripts ORDER BY created_at DESC`, playlistScriptColumns)

    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to list playlist scripts: %w", err)
    }
    defer rows.Close()

    return scanPlaylistScriptRows(rows)
}

// ListByUser retrieves playlist scripts owned by a specific user
func (r *PlaylistScriptRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.PlaylistScript, error) {
    query := fmt.Sprintf(`SELECT %s FROM playlist_scripts WHERE created_by = $1 ORDER BY created_at DESC`, playlistScriptColumns)

    rows, err := r.pool.Query(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to list user playlist scripts: %w", err)
    }
    defer rows.Close()

    return scanPlaylistScriptRows(rows)
}

// GetByID retrieves a playlist script by ID
func (r *PlaylistScriptRepository) GetByID(ctx context.Context, scriptID uuid.UUID) (*models.PlaylistScript, error) {
    query := fmt.Sprintf(`SELECT %s FROM playlist_scripts WHERE id = $1`, playlistScriptColumns)

    script, err := scanPlaylistScript(r.pool.QueryRow(ctx, query, scriptID))
    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get playlist script: %w", err)
    }

    return script, nil
}

// Create inserts a new playlist script
func (r *PlaylistScriptRepository) Create(ctx context.Context, script *models.PlaylistScript) error {
    query := `
        INSERT INTO playlist_scripts (
            id, name, description, sort, timeframe, clip_limit, visibility, is_active,
            schedule, strategy, game_id, game_ids, broadcaster_id, tag, exclude_tags,
            language, min_vote_score, min_view_count, exclude_nsfw, top_10k_streamers,
            seed_clip_id, retention_days, title_template, created_by
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
        RETURNING created_at, updated_at
    `

    err := r.pool.QueryRow(ctx, query,
        script.ID,
        script.Name,
        script.Description,
        script.Sort,
        script.Timeframe,
        script.ClipLimit,
        script.Visibility,
        script.IsActive,
        script.Schedule,
        script.Strategy,
        script.GameID,
        script.GameIDs,
        script.BroadcasterID,
        script.Tag,
        script.ExcludeTags,
        script.Language,
        script.MinVoteScore,
        script.MinViewCount,
        script.ExcludeNSFW,
        script.Top10kStreamers,
        script.SeedClipID,
        script.RetentionDays,
        script.TitleTemplate,
        script.CreatedBy,
    ).Scan(&script.CreatedAt, &script.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to create playlist script: %w", err)
    }

    return nil
}

// Update updates an existing playlist script
func (r *PlaylistScriptRepository) Update(ctx context.Context, script *models.PlaylistScript) error {
    query := `
        UPDATE playlist_scripts
        SET name = $1, description = $2, sort = $3, timeframe = $4,
            clip_limit = $5, visibility = $6, is_active = $7,
            schedule = $8, strategy = $9, game_id = $10, game_ids = $11,
            broadcaster_id = $12, tag = $13, exclude_tags = $14, language = $15,
            min_vote_score = $16, min_view_count = $17, exclude_nsfw = $18,
            top_10k_streamers = $19, seed_clip_id = $20, retention_days = $21,
            title_template = $22
        WHERE id = $23
        RETURNING updated_at
    `

    err := r.pool.QueryRow(ctx, query,
        script.Name,
        script.Description,
        script.Sort,
        script.Timeframe,
        script.ClipLimit,
        script.Visibility,
        script.IsActive,
        script.Schedule,
        script.Strategy,
        script.GameID,
        script.GameIDs,
        script.BroadcasterID,
        script.Tag,
        script.ExcludeTags,
        script.Language,
        script.MinVoteScore,
        script.MinViewCount,
        script.ExcludeNSFW,
        script.Top10kStreamers,
        script.SeedClipID,
        script.RetentionDays,
        script.TitleTemplate,
        script.ID,
    ).Scan(&script.UpdatedAt)

    if err == pgx.ErrNoRows {
        return fmt.Errorf("playlist script not found")
    }
    if err != nil {
        return fmt.Errorf("failed to update playlist script: %w", err)
    }

    return nil
}

// Delete removes a playlist script
func (r *PlaylistScriptRepository) Delete(ctx context.Context, scriptID uuid.UUID) error {
    query := `DELETE FROM playlist_scripts WHERE id = $1`

    result, err := r.pool.Exec(ctx, query, scriptID)
    if err != nil {
        return fmt.Errorf("failed to delete playlist script: %w", err)
    }
    if result.RowsAffected() == 0 {
        return fmt.Errorf("playlist script not found")
    }

    return nil
}

// CreateGeneratedPlaylist records a generated playlist instance
func (r *PlaylistScriptRepository) CreateGeneratedPlaylist(ctx context.Context, scriptID, playlistID uuid.UUID) error {
    query := `
        INSERT INTO generated_playlists (id, script_id, playlist_id, generated_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (playlist_id) DO NOTHING
    `

    _, err := r.pool.Exec(ctx, query, uuid.New(), scriptID, playlistID, time.Now())
    if err != nil {
        return fmt.Errorf("failed to record generated playlist: %w", err)
    }

    return nil
}

// UpdateLastRun updates script run metadata
func (r *PlaylistScriptRepository) UpdateLastRun(ctx context.Context, scriptID, playlistID uuid.UUID) error {
    query := `
        UPDATE playlist_scripts
        SET last_run_at = $1, last_generated_playlist_id = $2
        WHERE id = $3
    `

    result, err := r.pool.Exec(ctx, query, time.Now(), playlistID, scriptID)
    if err != nil {
        return fmt.Errorf("failed to update playlist script run metadata: %w", err)
    }
    if result.RowsAffected() == 0 {
        return fmt.Errorf("playlist script not found")
    }

    return nil
}

// ListDueForExecution returns active, non-manual scripts where enough time has elapsed since last_run_at
func (r *PlaylistScriptRepository) ListDueForExecution(ctx context.Context) ([]*models.PlaylistScript, error) {
    query := fmt.Sprintf(`
        SELECT %s
        FROM playlist_scripts
        WHERE is_active = true
          AND schedule != 'manual'
          AND (
            last_run_at IS NULL
            OR (schedule = 'hourly'  AND last_run_at < NOW() - INTERVAL '1 hour')
            OR (schedule = 'daily'   AND last_run_at < NOW() - INTERVAL '1 day')
            OR (schedule = 'weekly'  AND last_run_at < NOW() - INTERVAL '7 days')
            OR (schedule = 'monthly' AND last_run_at < NOW() - INTERVAL '30 days')
          )
        ORDER BY last_run_at ASC NULLS FIRST
    `, playlistScriptColumns)

    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to list scripts due for execution: %w", err)
    }
    defer rows.Close()

    return scanPlaylistScriptRows(rows)
}

// DeleteStaleGeneratedPlaylists deletes generated playlists older than their script's retention_days.
// It soft-deletes the playlists and removes the generated_playlists tracking records.
func (r *PlaylistScriptRepository) DeleteStaleGeneratedPlaylists(ctx context.Context) (int64, error) {
    query := `
        WITH stale AS (
            SELECT gp.id AS gp_id, gp.playlist_id
            FROM generated_playlists gp
            JOIN playlist_scripts ps ON ps.id = gp.script_id
            WHERE gp.generated_at < NOW() - (ps.retention_days || ' days')::INTERVAL
        ),
        soft_deleted AS (
            UPDATE playlists
            SET deleted_at = NOW()
            WHERE id IN (SELECT playlist_id FROM stale)
              AND deleted_at IS NULL
        )
        DELETE FROM generated_playlists
        WHERE id IN (SELECT gp_id FROM stale)
    `

    result, err := r.pool.Exec(ctx, query)
    if err != nil {
        return 0, fmt.Errorf("failed to delete stale generated playlists: %w", err)
    }

    return result.RowsAffected(), nil
}
