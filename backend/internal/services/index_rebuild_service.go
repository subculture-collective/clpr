package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	"git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
)

// IndexRebuildService handles rebuilding search indices with zero-downtime swaps
type IndexRebuildService struct {
	db             *database.DB
	osClient       *opensearch.Client
	indexer        *SearchIndexerService
	versionService *IndexVersionService
}

// RebuildConfig contains configuration for index rebuild
type RebuildConfig struct {
	BatchSize       int           `json:"batch_size"`
	KeepOldVersions int           `json:"keep_old_versions"`
	SwapAfterBuild  bool          `json:"swap_after_build"`
	Verbose         bool          `json:"verbose"`
	BatchDelay      time.Duration `json:"batch_delay"` // Delay between batches to control load
}

// RebuildResult contains the result of a rebuild operation
type RebuildResult struct {
	BaseIndex      string        `json:"base_index"`
	NewVersion     int           `json:"new_version"`
	NewIndexName   string        `json:"new_index_name"`
	DocCount       int64         `json:"doc_count"`
	Duration       time.Duration `json:"duration"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	SwappedAlias   bool          `json:"swapped_alias"`
	DeletedIndices []string      `json:"deleted_indices,omitempty"`
	Error          string        `json:"error,omitempty"`
}

// RebuildAllResult contains results for rebuilding all indices
type RebuildAllResult struct {
	Results       []RebuildResult `json:"results"`
	TotalDuration time.Duration   `json:"total_duration"`
	Success       bool            `json:"success"`
	Errors        []string        `json:"errors,omitempty"`
}

// DefaultRebuildConfig returns the default rebuild configuration
func DefaultRebuildConfig() *RebuildConfig {
	return &RebuildConfig{
		BatchSize:       100,
		KeepOldVersions: 2,
		SwapAfterBuild:  true,
		Verbose:         true,
		BatchDelay:      100 * time.Millisecond,
	}
}

// NewIndexRebuildService creates a new IndexRebuildService
func NewIndexRebuildService(db *database.DB, osClient *opensearch.Client) *IndexRebuildService {
	indexer := NewSearchIndexerService(osClient)
	versionService := NewIndexVersionService(osClient)

	return &IndexRebuildService{
		db:             db,
		osClient:       osClient,
		indexer:        indexer,
		versionService: versionService,
	}
}

// RebuildClipsIndex rebuilds the clips index with zero-downtime swap
func (s *IndexRebuildService) RebuildClipsIndex(ctx context.Context, config *RebuildConfig) (*RebuildResult, error) {
	if config == nil {
		config = DefaultRebuildConfig()
	}

	result := &RebuildResult{
		BaseIndex: ClipsIndex,
		StartTime: time.Now(),
	}

	// Get next version
	nextVersion, err := s.versionService.GetNextVersion(ctx, ClipsIndex)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get next version: %v", err)
		return result, fmt.Errorf("failed to get next version: %w", err)
	}

	result.NewVersion = nextVersion
	result.NewIndexName = s.versionService.GetVersionedIndexName(ClipsIndex, nextVersion)

	log.Printf("Starting rebuild of %s index (new version: %d)", ClipsIndex, nextVersion)

	// Create new versioned index
	mapping := getClipIndexMapping()
	if err := s.versionService.CreateVersionedIndex(ctx, ClipsIndex, nextVersion, mapping); err != nil {
		result.Error = fmt.Sprintf("failed to create versioned index: %v", err)
		return result, fmt.Errorf("failed to create versioned index: %w", err)
	}

	// Index all clips to the new index
	docCount, err := s.indexClipsToVersionedIndex(ctx, result.NewIndexName, config)
	if err != nil {
		result.Error = fmt.Sprintf("failed to index clips: %v", err)
		return result, fmt.Errorf("failed to index clips: %w", err)
	}

	result.DocCount = docCount

	// Refresh index to make all documents visible
	if err := s.versionService.RefreshIndex(ctx, result.NewIndexName); err != nil {
		log.Printf("WARNING: Failed to refresh index: %v", err)
	}

	// Swap alias if configured
	if config.SwapAfterBuild {
		if err := s.versionService.SwapAlias(ctx, ClipsIndex, nextVersion); err != nil {
			result.Error = fmt.Sprintf("failed to swap alias: %v", err)
			return result, fmt.Errorf("failed to swap alias: %w", err)
		}
		result.SwappedAlias = true
	}

	// Clean up old versions
	if config.KeepOldVersions > 0 {
		deleted, err := s.versionService.DeleteOldVersions(ctx, ClipsIndex, config.KeepOldVersions)
		if err != nil {
			log.Printf("WARNING: Failed to delete old versions: %v", err)
		}
		result.DeletedIndices = deleted
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Completed rebuild of %s index: %d docs in %v", ClipsIndex, docCount, result.Duration)

	return result, nil
}

// indexClipsToVersionedIndex indexes all clips to a specific versioned index
func (s *IndexRebuildService) indexClipsToVersionedIndex(ctx context.Context, indexName string, config *RebuildConfig) (int64, error) {
	var totalIndexed int64
	offset := 0

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

		rows, err := s.db.Pool.Query(ctx, query, config.BatchSize, offset)
		if err != nil {
			return totalIndexed, fmt.Errorf("failed to fetch clips: %w", err)
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
				return totalIndexed, fmt.Errorf("failed to scan clip: %w", err)
			}
			clips = append(clips, clip)
		}
		rows.Close()

		if len(clips) == 0 {
			break
		}

		// Bulk index to the specific versioned index
		if err := s.bulkIndexClipsToIndex(ctx, indexName, clips); err != nil {
			return totalIndexed, fmt.Errorf("failed to bulk index clips: %w", err)
		}

		totalIndexed += int64(len(clips))

		if config.Verbose {
			log.Printf("Indexed %d clips to %s (total: %d)", len(clips), indexName, totalIndexed)
		}

		offset += config.BatchSize
		if config.BatchDelay > 0 {
			time.Sleep(config.BatchDelay)
		}
	}

	return totalIndexed, nil
}

// bulkIndexClipsToIndex performs bulk indexing to a specific index name
func (s *IndexRebuildService) bulkIndexClipsToIndex(ctx context.Context, indexName string, clips []models.Clip) error {
	if len(clips) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, clip := range clips {
		// Action metadata
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    clip.ID.String(),
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for clip %s: %w", clip.ID, err)
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// Document data
		engagementScore := calculateEngagementScore(&clip)
		recencyScore := calculateRecencyScore(&clip)

		doc := map[string]interface{}{
			"id":               clip.ID.String(),
			"twitch_clip_id":   clip.TwitchClipID,
			"title":            clip.Title,
			"creator_name":     clip.CreatorName,
			"creator_id":       clip.CreatorID,
			"broadcaster_name": clip.BroadcasterName,
			"broadcaster_id":   clip.BroadcasterID,
			"game_id":          clip.GameID,
			"game_name":        clip.GameName,
			"language":         clip.Language,
			"view_count":       clip.ViewCount,
			"vote_score":       clip.VoteScore,
			"comment_count":    clip.CommentCount,
			"favorite_count":   clip.FavoriteCount,
			"is_featured":      clip.IsFeatured,
			"is_nsfw":          clip.IsNSFW,
			"is_removed":       clip.IsRemoved,
			"created_at":       clip.CreatedAt,
			"imported_at":      clip.ImportedAt,
			"engagement_score": engagementScore,
			"recency_score":    recencyScore,
		}
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document for clip %s: %w", clip.ID, err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk request error: %s - %s", res.Status(), string(bodyBytes))
	}

	return nil
}

// RebuildUsersIndex rebuilds the users index with zero-downtime swap
func (s *IndexRebuildService) RebuildUsersIndex(ctx context.Context, config *RebuildConfig) (*RebuildResult, error) {
	if config == nil {
		config = DefaultRebuildConfig()
	}

	result := &RebuildResult{
		BaseIndex: UsersIndex,
		StartTime: time.Now(),
	}

	nextVersion, err := s.versionService.GetNextVersion(ctx, UsersIndex)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get next version: %v", err)
		return result, fmt.Errorf("failed to get next version: %w", err)
	}

	result.NewVersion = nextVersion
	result.NewIndexName = s.versionService.GetVersionedIndexName(UsersIndex, nextVersion)

	log.Printf("Starting rebuild of %s index (new version: %d)", UsersIndex, nextVersion)

	mapping := getUserIndexMapping()
	if err := s.versionService.CreateVersionedIndex(ctx, UsersIndex, nextVersion, mapping); err != nil {
		result.Error = fmt.Sprintf("failed to create versioned index: %v", err)
		return result, fmt.Errorf("failed to create versioned index: %w", err)
	}

	docCount, err := s.indexUsersToVersionedIndex(ctx, result.NewIndexName, config)
	if err != nil {
		result.Error = fmt.Sprintf("failed to index users: %v", err)
		return result, fmt.Errorf("failed to index users: %w", err)
	}

	result.DocCount = docCount

	if err := s.versionService.RefreshIndex(ctx, result.NewIndexName); err != nil {
		log.Printf("WARNING: Failed to refresh index: %v", err)
	}

	if config.SwapAfterBuild {
		if err := s.versionService.SwapAlias(ctx, UsersIndex, nextVersion); err != nil {
			result.Error = fmt.Sprintf("failed to swap alias: %v", err)
			return result, fmt.Errorf("failed to swap alias: %w", err)
		}
		result.SwappedAlias = true
	}

	if config.KeepOldVersions > 0 {
		deleted, err := s.versionService.DeleteOldVersions(ctx, UsersIndex, config.KeepOldVersions)
		if err != nil {
			log.Printf("WARNING: Failed to delete old versions: %v", err)
		}
		result.DeletedIndices = deleted
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Completed rebuild of %s index: %d docs in %v", UsersIndex, docCount, result.Duration)

	return result, nil
}

func (s *IndexRebuildService) indexUsersToVersionedIndex(ctx context.Context, indexName string, config *RebuildConfig) (int64, error) {
	var totalIndexed int64
	offset := 0

	for {
		query := `
			SELECT id, twitch_id, username, display_name, email, avatar_url,
			       bio, karma_points, role, is_banned, created_at, updated_at, last_login_at
			FROM users
			WHERE is_banned = false
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := s.db.Pool.Query(ctx, query, config.BatchSize, offset)
		if err != nil {
			return totalIndexed, fmt.Errorf("failed to fetch users: %w", err)
		}

		var users []models.User
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
				return totalIndexed, fmt.Errorf("failed to scan user: %w", err)
			}
			users = append(users, user)
		}
		rows.Close()

		if len(users) == 0 {
			break
		}

		if err := s.bulkIndexUsersToIndex(ctx, indexName, users); err != nil {
			return totalIndexed, fmt.Errorf("failed to bulk index users: %w", err)
		}

		totalIndexed += int64(len(users))

		if config.Verbose {
			log.Printf("Indexed %d users to %s (total: %d)", len(users), indexName, totalIndexed)
		}

		offset += config.BatchSize
		if config.BatchDelay > 0 {
			time.Sleep(config.BatchDelay)
		}
	}

	return totalIndexed, nil
}

func (s *IndexRebuildService) bulkIndexUsersToIndex(ctx context.Context, indexName string, users []models.User) error {
	if len(users) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, user := range users {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    user.ID.String(),
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for user %s: %w", user.ID, err)
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		doc := map[string]interface{}{
			"id":            user.ID.String(),
			"twitch_id":     user.TwitchID,
			"username":      user.Username,
			"display_name":  user.DisplayName,
			"bio":           user.Bio,
			"karma_points":  user.KarmaPoints,
			"role":          user.Role,
			"is_banned":     user.IsBanned,
			"created_at":    user.CreatedAt,
			"last_login_at": user.LastLoginAt,
		}
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document for user %s: %w", user.ID, err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk request error: %s - %s", res.Status(), string(bodyBytes))
	}

	return nil
}

// RebuildTagsIndex rebuilds the tags index with zero-downtime swap
func (s *IndexRebuildService) RebuildTagsIndex(ctx context.Context, config *RebuildConfig) (*RebuildResult, error) {
	if config == nil {
		config = DefaultRebuildConfig()
	}

	result := &RebuildResult{
		BaseIndex: TagsIndex,
		StartTime: time.Now(),
	}

	nextVersion, err := s.versionService.GetNextVersion(ctx, TagsIndex)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get next version: %v", err)
		return result, fmt.Errorf("failed to get next version: %w", err)
	}

	result.NewVersion = nextVersion
	result.NewIndexName = s.versionService.GetVersionedIndexName(TagsIndex, nextVersion)

	log.Printf("Starting rebuild of %s index (new version: %d)", TagsIndex, nextVersion)

	mapping := getTagIndexMapping()
	if err := s.versionService.CreateVersionedIndex(ctx, TagsIndex, nextVersion, mapping); err != nil {
		result.Error = fmt.Sprintf("failed to create versioned index: %v", err)
		return result, fmt.Errorf("failed to create versioned index: %w", err)
	}

	docCount, err := s.indexTagsToVersionedIndex(ctx, result.NewIndexName, config)
	if err != nil {
		result.Error = fmt.Sprintf("failed to index tags: %v", err)
		return result, fmt.Errorf("failed to index tags: %w", err)
	}

	result.DocCount = docCount

	if err := s.versionService.RefreshIndex(ctx, result.NewIndexName); err != nil {
		log.Printf("WARNING: Failed to refresh index: %v", err)
	}

	if config.SwapAfterBuild {
		if err := s.versionService.SwapAlias(ctx, TagsIndex, nextVersion); err != nil {
			result.Error = fmt.Sprintf("failed to swap alias: %v", err)
			return result, fmt.Errorf("failed to swap alias: %w", err)
		}
		result.SwappedAlias = true
	}

	if config.KeepOldVersions > 0 {
		deleted, err := s.versionService.DeleteOldVersions(ctx, TagsIndex, config.KeepOldVersions)
		if err != nil {
			log.Printf("WARNING: Failed to delete old versions: %v", err)
		}
		result.DeletedIndices = deleted
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Completed rebuild of %s index: %d docs in %v", TagsIndex, docCount, result.Duration)

	return result, nil
}

func (s *IndexRebuildService) indexTagsToVersionedIndex(ctx context.Context, indexName string, config *RebuildConfig) (int64, error) {
	var totalIndexed int64
	offset := 0

	for {
		query := `
			SELECT id, name, slug, description, color, usage_count, created_at
			FROM tags
			ORDER BY id
			LIMIT $1 OFFSET $2
		`

		rows, err := s.db.Pool.Query(ctx, query, config.BatchSize, offset)
		if err != nil {
			return totalIndexed, fmt.Errorf("failed to fetch tags: %w", err)
		}

		var tags []models.Tag
		for rows.Next() {
			var tag models.Tag
			err := rows.Scan(
				&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
				&tag.Color, &tag.UsageCount, &tag.CreatedAt,
			)
			if err != nil {
				rows.Close()
				return totalIndexed, fmt.Errorf("failed to scan tag: %w", err)
			}
			tags = append(tags, tag)
		}
		rows.Close()

		if len(tags) == 0 {
			break
		}

		if err := s.bulkIndexTagsToIndex(ctx, indexName, tags); err != nil {
			return totalIndexed, fmt.Errorf("failed to bulk index tags: %w", err)
		}

		totalIndexed += int64(len(tags))

		if config.Verbose {
			log.Printf("Indexed %d tags to %s (total: %d)", len(tags), indexName, totalIndexed)
		}

		offset += config.BatchSize
		if config.BatchDelay > 0 {
			time.Sleep(config.BatchDelay)
		}
	}

	return totalIndexed, nil
}

func (s *IndexRebuildService) bulkIndexTagsToIndex(ctx context.Context, indexName string, tags []models.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, tag := range tags {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    tag.ID.String(),
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for tag %s: %w", tag.ID, err)
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		doc := map[string]interface{}{
			"id":          tag.ID.String(),
			"name":        tag.Name,
			"slug":        tag.Slug,
			"description": tag.Description,
			"usage_count": tag.UsageCount,
			"created_at":  tag.CreatedAt,
		}
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document for tag %s: %w", tag.ID, err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk request error: %s - %s", res.Status(), string(bodyBytes))
	}

	return nil
}

// RebuildGamesIndex rebuilds the games index with zero-downtime swap
func (s *IndexRebuildService) RebuildGamesIndex(ctx context.Context, config *RebuildConfig) (*RebuildResult, error) {
	if config == nil {
		config = DefaultRebuildConfig()
	}

	result := &RebuildResult{
		BaseIndex: GamesIndex,
		StartTime: time.Now(),
	}

	nextVersion, err := s.versionService.GetNextVersion(ctx, GamesIndex)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get next version: %v", err)
		return result, fmt.Errorf("failed to get next version: %w", err)
	}

	result.NewVersion = nextVersion
	result.NewIndexName = s.versionService.GetVersionedIndexName(GamesIndex, nextVersion)

	log.Printf("Starting rebuild of %s index (new version: %d)", GamesIndex, nextVersion)

	mapping := getGameIndexMapping()
	if err := s.versionService.CreateVersionedIndex(ctx, GamesIndex, nextVersion, mapping); err != nil {
		result.Error = fmt.Sprintf("failed to create versioned index: %v", err)
		return result, fmt.Errorf("failed to create versioned index: %w", err)
	}

	docCount, err := s.indexGamesToVersionedIndex(ctx, result.NewIndexName, config)
	if err != nil {
		result.Error = fmt.Sprintf("failed to index games: %v", err)
		return result, fmt.Errorf("failed to index games: %w", err)
	}

	result.DocCount = docCount

	if err := s.versionService.RefreshIndex(ctx, result.NewIndexName); err != nil {
		log.Printf("WARNING: Failed to refresh index: %v", err)
	}

	if config.SwapAfterBuild {
		if err := s.versionService.SwapAlias(ctx, GamesIndex, nextVersion); err != nil {
			result.Error = fmt.Sprintf("failed to swap alias: %v", err)
			return result, fmt.Errorf("failed to swap alias: %w", err)
		}
		result.SwappedAlias = true
	}

	if config.KeepOldVersions > 0 {
		deleted, err := s.versionService.DeleteOldVersions(ctx, GamesIndex, config.KeepOldVersions)
		if err != nil {
			log.Printf("WARNING: Failed to delete old versions: %v", err)
		}
		result.DeletedIndices = deleted
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Completed rebuild of %s index: %d docs in %v", GamesIndex, docCount, result.Duration)

	return result, nil
}

func (s *IndexRebuildService) indexGamesToVersionedIndex(ctx context.Context, indexName string, config *RebuildConfig) (int64, error) {
	var totalIndexed int64
	offset := 0

	for {
		query := `
			SELECT game_id, game_name, COUNT(*) as clip_count
			FROM clips
			WHERE game_id IS NOT NULL AND game_name IS NOT NULL AND is_removed = false
			GROUP BY game_id, game_name
			ORDER BY game_id
			LIMIT $1 OFFSET $2
		`

		rows, err := s.db.Pool.Query(ctx, query, config.BatchSize, offset)
		if err != nil {
			return totalIndexed, fmt.Errorf("failed to fetch games: %w", err)
		}

		var games []models.GameSearchResult
		for rows.Next() {
			var game models.GameSearchResult
			err := rows.Scan(&game.ID, &game.Name, &game.ClipCount)
			if err != nil {
				rows.Close()
				return totalIndexed, fmt.Errorf("failed to scan game: %w", err)
			}
			games = append(games, game)
		}
		rows.Close()

		if len(games) == 0 {
			break
		}

		if err := s.bulkIndexGamesToIndex(ctx, indexName, games); err != nil {
			return totalIndexed, fmt.Errorf("failed to bulk index games: %w", err)
		}

		totalIndexed += int64(len(games))

		if config.Verbose {
			log.Printf("Indexed %d games to %s (total: %d)", len(games), indexName, totalIndexed)
		}

		offset += config.BatchSize
		if config.BatchDelay > 0 {
			time.Sleep(config.BatchDelay)
		}
	}

	return totalIndexed, nil
}

func (s *IndexRebuildService) bulkIndexGamesToIndex(ctx context.Context, indexName string, games []models.GameSearchResult) error {
	if len(games) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, game := range games {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    game.ID,
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for game %s: %w", game.ID, err)
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		doc := map[string]interface{}{
			"id":         game.ID,
			"name":       game.Name,
			"clip_count": game.ClipCount,
		}
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document for game %s: %w", game.ID, err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk request error: %s - %s", res.Status(), string(bodyBytes))
	}

	return nil
}

// RebuildAllIndices rebuilds all search indices with zero-downtime swaps
func (s *IndexRebuildService) RebuildAllIndices(ctx context.Context, config *RebuildConfig) (*RebuildAllResult, error) {
	if config == nil {
		config = DefaultRebuildConfig()
	}

	startTime := time.Now()
	result := &RebuildAllResult{
		Results: []RebuildResult{},
		Success: true,
	}

	log.Println("Starting rebuild of all search indices...")

	// Rebuild clips index
	clipsResult, err := s.RebuildClipsIndex(ctx, config)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("clips: %v", err))
		result.Success = false
	}
	result.Results = append(result.Results, *clipsResult)

	// Rebuild users index
	usersResult, err := s.RebuildUsersIndex(ctx, config)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("users: %v", err))
		result.Success = false
	}
	result.Results = append(result.Results, *usersResult)

	// Rebuild tags index
	tagsResult, err := s.RebuildTagsIndex(ctx, config)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("tags: %v", err))
		result.Success = false
	}
	result.Results = append(result.Results, *tagsResult)

	// Rebuild games index
	gamesResult, err := s.RebuildGamesIndex(ctx, config)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("games: %v", err))
		result.Success = false
	}
	result.Results = append(result.Results, *gamesResult)

	result.TotalDuration = time.Since(startTime)

	if result.Success {
		log.Printf("Completed rebuild of all indices in %v", result.TotalDuration)
	} else {
		log.Printf("Rebuild completed with errors in %v: %v", result.TotalDuration, result.Errors)
	}

	return result, nil
}

// GetVersionService returns the underlying version service for direct access
func (s *IndexRebuildService) GetVersionService() *IndexVersionService {
	return s.versionService
}
