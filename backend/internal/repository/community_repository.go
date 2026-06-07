package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

type CommunityRepository struct {
	pool *pgxpool.Pool
}

func NewCommunityRepository(pool *pgxpool.Pool) *CommunityRepository {
	return &CommunityRepository{pool: pool}
}

// CreateCommunity creates a new community
func (r *CommunityRepository) CreateCommunity(ctx context.Context, community *models.Community) error {
	query := `
		INSERT INTO communities (id, name, slug, description, icon, owner_id, is_public, member_count, rules, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		community.ID, community.Name, community.Slug, community.Description, community.Icon,
		community.OwnerID, community.IsPublic, community.MemberCount, community.Rules,
		community.CreatedAt, community.UpdatedAt,
	).Scan(&community.ID, &community.CreatedAt, &community.UpdatedAt)
}

// GetCommunityByID retrieves a community by ID
func (r *CommunityRepository) GetCommunityByID(ctx context.Context, id uuid.UUID) (*models.Community, error) {
	query := `
		SELECT id, name, slug, description, icon, owner_id, is_public, member_count, rules, created_at, updated_at
		FROM communities
		WHERE id = $1
	`
	community := &models.Community{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&community.ID, &community.Name, &community.Slug, &community.Description, &community.Icon,
		&community.OwnerID, &community.IsPublic, &community.MemberCount, &community.Rules,
		&community.CreatedAt, &community.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("community not found")
	}
	return community, err
}

// GetCommunityBySlug retrieves a community by slug
func (r *CommunityRepository) GetCommunityBySlug(ctx context.Context, slug string) (*models.Community, error) {
	query := `
		SELECT id, name, slug, description, icon, owner_id, is_public, member_count, rules, created_at, updated_at
		FROM communities
		WHERE slug = $1
	`
	community := &models.Community{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&community.ID, &community.Name, &community.Slug, &community.Description, &community.Icon,
		&community.OwnerID, &community.IsPublic, &community.MemberCount, &community.Rules,
		&community.CreatedAt, &community.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("community not found")
	}
	return community, err
}

// ListCommunities retrieves all communities with optional filters
func (r *CommunityRepository) ListCommunities(ctx context.Context, limit, offset int, publicOnly bool, sort string) ([]*models.Community, int, error) {
	// Build query conditions
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if publicOnly {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Determine sort order
	orderBy := "ORDER BY created_at DESC"
	switch sort {
	case "members":
		orderBy = "ORDER BY member_count DESC, created_at DESC"
	case "name":
		orderBy = "ORDER BY name ASC"
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM communities %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query communities
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, icon, owner_id, is_public, member_count, rules, created_at, updated_at
		FROM communities
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	communities := []*models.Community{}
	for rows.Next() {
		community := &models.Community{}
		err := rows.Scan(
			&community.ID, &community.Name, &community.Slug, &community.Description, &community.Icon,
			&community.OwnerID, &community.IsPublic, &community.MemberCount, &community.Rules,
			&community.CreatedAt, &community.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		communities = append(communities, community)
	}
	return communities, total, rows.Err()
}

// SearchCommunities searches communities by name
func (r *CommunityRepository) SearchCommunities(ctx context.Context, query string, limit, offset int) ([]*models.Community, int, error) {
	searchPattern := "%" + strings.ToLower(query) + "%"

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM communities
		WHERE LOWER(name) LIKE $1
	`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query communities
	querySQL := `
		SELECT id, name, slug, description, icon, owner_id, is_public, member_count, rules, created_at, updated_at
		FROM communities
		WHERE LOWER(name) LIKE $1
		ORDER BY member_count DESC, name ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, querySQL, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	communities := []*models.Community{}
	for rows.Next() {
		community := &models.Community{}
		err := rows.Scan(
			&community.ID, &community.Name, &community.Slug, &community.Description, &community.Icon,
			&community.OwnerID, &community.IsPublic, &community.MemberCount, &community.Rules,
			&community.CreatedAt, &community.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		communities = append(communities, community)
	}
	return communities, total, rows.Err()
}

// UpdateCommunity updates a community
func (r *CommunityRepository) UpdateCommunity(ctx context.Context, community *models.Community) error {
	query := `
		UPDATE communities
		SET name = $2, slug = $3, description = $4, icon = $5, is_public = $6, rules = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return r.pool.QueryRow(ctx, query,
		community.ID, community.Name, community.Slug, community.Description, community.Icon, community.IsPublic, community.Rules,
	).Scan(&community.UpdatedAt)
}

// DeleteCommunity deletes a community
func (r *CommunityRepository) DeleteCommunity(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM communities WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// AddMember adds a member to a community
func (r *CommunityRepository) AddMember(ctx context.Context, member *models.CommunityMember) error {
	query := `
		INSERT INTO community_members (id, community_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, joined_at
	`
	return r.pool.QueryRow(ctx, query,
		member.ID, member.CommunityID, member.UserID, member.Role, member.JoinedAt,
	).Scan(&member.ID, &member.JoinedAt)
}

// RemoveMember removes a member from a community
func (r *CommunityRepository) RemoveMember(ctx context.Context, communityID, userID uuid.UUID) error {
	query := `DELETE FROM community_members WHERE community_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, communityID, userID)
	return err
}

// GetMember retrieves a specific member from a community
func (r *CommunityRepository) GetMember(ctx context.Context, communityID, userID uuid.UUID) (*models.CommunityMember, error) {
	query := `
		SELECT id, community_id, user_id, role, joined_at
		FROM community_members
		WHERE community_id = $1 AND user_id = $2
	`
	member := &models.CommunityMember{}
	err := r.pool.QueryRow(ctx, query, communityID, userID).Scan(
		&member.ID, &member.CommunityID, &member.UserID, &member.Role, &member.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // Not a member
	}
	return member, err
}

// ListMembers retrieves all members of a community
func (r *CommunityRepository) ListMembers(ctx context.Context, communityID uuid.UUID, role string, limit, offset int) ([]*models.CommunityMember, int, error) {
	// Build query conditions
	conditions := []string{"community_id = $1"}
	args := []interface{}{communityID}
	argIndex := 2

	if role != "" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, role)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM community_members %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query members
	query := fmt.Sprintf(`
		SELECT id, community_id, user_id, role, joined_at
		FROM community_members
		%s
		ORDER BY joined_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	members := []*models.CommunityMember{}
	for rows.Next() {
		member := &models.CommunityMember{}
		err := rows.Scan(
			&member.ID, &member.CommunityID, &member.UserID, &member.Role, &member.JoinedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		members = append(members, member)
	}
	return members, total, rows.Err()
}

// UpdateMemberRole updates a member's role in a community
func (r *CommunityRepository) UpdateMemberRole(ctx context.Context, communityID, userID uuid.UUID, role string) error {
	query := `
		UPDATE community_members
		SET role = $3
		WHERE community_id = $1 AND user_id = $2
	`
	result, err := r.pool.Exec(ctx, query, communityID, userID, role)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// IsMember checks if a user is a member of a community
func (r *CommunityRepository) IsMember(ctx context.Context, communityID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM community_members WHERE community_id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, communityID, userID).Scan(&exists)
	return exists, err
}

// IsBanned checks if a user is banned from a community
func (r *CommunityRepository) IsBanned(ctx context.Context, communityID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM community_bans WHERE community_id = $1 AND banned_user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, communityID, userID).Scan(&exists)
	return exists, err
}

// BanMember bans a user from a community
func (r *CommunityRepository) BanMember(ctx context.Context, ban *models.CommunityBan) error {
	query := `
		INSERT INTO community_bans (id, community_id, banned_user_id, banned_by_user_id, reason, banned_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, banned_at
	`
	return r.pool.QueryRow(ctx, query,
		ban.ID, ban.CommunityID, ban.BannedUserID, ban.BannedByUserID, ban.Reason, ban.BannedAt,
	).Scan(&ban.ID, &ban.BannedAt)
}

// GetBanByID retrieves a ban record by its ID
func (r *CommunityRepository) GetBanByID(ctx context.Context, banID uuid.UUID) (*models.CommunityBan, error) {
	query := `
		SELECT id, community_id, banned_user_id, banned_by_user_id, reason, banned_at
		FROM community_bans
		WHERE id = $1
	`
	ban := &models.CommunityBan{}
	err := r.pool.QueryRow(ctx, query, banID).Scan(
		&ban.ID, &ban.CommunityID, &ban.BannedUserID, &ban.BannedByUserID, &ban.Reason, &ban.BannedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("ban not found: %s", banID.String())
	}
	return ban, err
}

// UnbanMember unbans a user from a community
func (r *CommunityRepository) UnbanMember(ctx context.Context, communityID, userID uuid.UUID) error {
	query := `DELETE FROM community_bans WHERE community_id = $1 AND banned_user_id = $2`
	_, err := r.pool.Exec(ctx, query, communityID, userID)
	return err
}

// ListBans retrieves all bans for a community
func (r *CommunityRepository) ListBans(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]*models.CommunityBan, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM community_bans WHERE community_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, communityID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query bans
	query := `
		SELECT id, community_id, banned_user_id, banned_by_user_id, reason, banned_at
		FROM community_bans
		WHERE community_id = $1
		ORDER BY banned_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, communityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bans := []*models.CommunityBan{}
	for rows.Next() {
		ban := &models.CommunityBan{}
		err := rows.Scan(
			&ban.ID, &ban.CommunityID, &ban.BannedUserID, &ban.BannedByUserID, &ban.Reason, &ban.BannedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		bans = append(bans, ban)
	}
	return bans, total, rows.Err()
}

// ListAllBans retrieves all bans across all communities with pagination (admin only)
func (r *CommunityRepository) ListAllBans(ctx context.Context, limit, offset int) ([]*models.CommunityBan, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM community_bans`
	var total int
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query bans
	query := `
		SELECT id, community_id, banned_user_id, banned_by_user_id, reason, banned_at
		FROM community_bans
		ORDER BY banned_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bans := []*models.CommunityBan{}
	for rows.Next() {
		ban := &models.CommunityBan{}
		err := rows.Scan(
			&ban.ID, &ban.CommunityID, &ban.BannedUserID, &ban.BannedByUserID, &ban.Reason, &ban.BannedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		bans = append(bans, ban)
	}
	return bans, total, rows.Err()
}

// AddClipToCommunity adds a clip to a community feed
func (r *CommunityRepository) AddClipToCommunity(ctx context.Context, communityClip *models.CommunityClip) error {
	query := `
		INSERT INTO community_clips (id, community_id, clip_id, added_by_user_id, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (community_id, clip_id) DO NOTHING
		RETURNING id, added_at
	`
	err := r.pool.QueryRow(ctx, query,
		communityClip.ID, communityClip.CommunityID, communityClip.ClipID, communityClip.AddedByUserID, communityClip.AddedAt,
	).Scan(&communityClip.ID, &communityClip.AddedAt)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("clip already exists in community")
	}
	return err
}

// RemoveClipFromCommunity removes a clip from a community feed
func (r *CommunityRepository) RemoveClipFromCommunity(ctx context.Context, communityID, clipID uuid.UUID) error {
	query := `DELETE FROM community_clips WHERE community_id = $1 AND clip_id = $2`
	_, err := r.pool.Exec(ctx, query, communityID, clipID)
	return err
}

// GetCommunityClips retrieves clips from a community feed with full clip data
func (r *CommunityRepository) GetCommunityClips(ctx context.Context, communityID uuid.UUID, sort string, limit, offset int) ([]*models.CommunityClipWithClip, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM community_clips WHERE community_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, communityID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Determine sort order
	orderBy := "ORDER BY cc.added_at DESC"
	if sort == "trending" {
		// Use vote score and recency for trending
		orderBy = "ORDER BY c.vote_score DESC, cc.added_at DESC"
	}

	// Query clips with JOIN to get full clip data
	query := fmt.Sprintf(`
		SELECT
			cc.id, cc.community_id, cc.clip_id, cc.added_by_user_id, cc.added_at,
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score,
			c.comment_count, c.favorite_count, c.is_featured, c.is_nsfw,
			c.is_removed, c.removed_reason, c.is_hidden
		FROM community_clips cc
		JOIN clips c ON cc.clip_id = c.id
		WHERE cc.community_id = $1
		%s
		LIMIT $2 OFFSET $3
	`, orderBy)

	rows, err := r.pool.Query(ctx, query, communityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	clips := []*models.CommunityClipWithClip{}
	for rows.Next() {
		communityClip := &models.CommunityClip{}
		clip := &models.Clip{}
		err := rows.Scan(
			&communityClip.ID, &communityClip.CommunityID, &communityClip.ClipID, &communityClip.AddedByUserID, &communityClip.AddedAt,
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL, &clip.Title,
			&clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName, &clip.BroadcasterID,
			&clip.GameID, &clip.GameName, &clip.Language, &clip.ThumbnailURL, &clip.Duration,
			&clip.ViewCount, &clip.CreatedAt, &clip.ImportedAt, &clip.VoteScore,
			&clip.CommentCount, &clip.FavoriteCount, &clip.IsFeatured, &clip.IsNSFW,
			&clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
		)
		if err != nil {
			return nil, 0, err
		}
		clipWithClip := &models.CommunityClipWithClip{
			CommunityClip: *communityClip,
			Clip:          clip,
		}
		clips = append(clips, clipWithClip)
	}
	return clips, total, rows.Err()
}

// CreateDiscussion creates a new discussion thread
func (r *CommunityRepository) CreateDiscussion(ctx context.Context, discussion *models.CommunityDiscussion) error {
	query := `
		INSERT INTO community_discussions (id, community_id, user_id, title, content, is_pinned, is_resolved, vote_score, comment_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		discussion.ID, discussion.CommunityID, discussion.UserID, discussion.Title, discussion.Content,
		discussion.IsPinned, discussion.IsResolved, discussion.VoteScore, discussion.CommentCount,
		discussion.CreatedAt, discussion.UpdatedAt,
	).Scan(&discussion.ID, &discussion.CreatedAt, &discussion.UpdatedAt)
}

// GetDiscussion retrieves a discussion thread by ID
func (r *CommunityRepository) GetDiscussion(ctx context.Context, id uuid.UUID) (*models.CommunityDiscussion, error) {
	query := `
		SELECT id, community_id, user_id, title, content, is_pinned, is_resolved, vote_score, comment_count, created_at, updated_at
		FROM community_discussions
		WHERE id = $1
	`
	discussion := &models.CommunityDiscussion{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&discussion.ID, &discussion.CommunityID, &discussion.UserID, &discussion.Title, &discussion.Content,
		&discussion.IsPinned, &discussion.IsResolved, &discussion.VoteScore, &discussion.CommentCount,
		&discussion.CreatedAt, &discussion.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("discussion not found")
	}
	return discussion, err
}

// ListDiscussions retrieves discussions for a community
func (r *CommunityRepository) ListDiscussions(ctx context.Context, communityID uuid.UUID, sort string, limit, offset int) ([]*models.CommunityDiscussion, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM community_discussions WHERE community_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, communityID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Determine sort order
	orderBy := "ORDER BY is_pinned DESC, created_at DESC"
	switch sort {
	case "trending":
		orderBy = "ORDER BY is_pinned DESC, vote_score DESC, created_at DESC"
	case "recent":
		orderBy = "ORDER BY is_pinned DESC, updated_at DESC"
	}

	// Query discussions
	query := fmt.Sprintf(`
		SELECT id, community_id, user_id, title, content, is_pinned, is_resolved, vote_score, comment_count, created_at, updated_at
		FROM community_discussions
		WHERE community_id = $1
		%s
		LIMIT $2 OFFSET $3
	`, orderBy)

	rows, err := r.pool.Query(ctx, query, communityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	discussions := []*models.CommunityDiscussion{}
	for rows.Next() {
		discussion := &models.CommunityDiscussion{}
		err := rows.Scan(
			&discussion.ID, &discussion.CommunityID, &discussion.UserID, &discussion.Title, &discussion.Content,
			&discussion.IsPinned, &discussion.IsResolved, &discussion.VoteScore, &discussion.CommentCount,
			&discussion.CreatedAt, &discussion.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		discussions = append(discussions, discussion)
	}
	return discussions, total, rows.Err()
}

// UpdateDiscussion updates a discussion thread
func (r *CommunityRepository) UpdateDiscussion(ctx context.Context, discussion *models.CommunityDiscussion) error {
	query := `
		UPDATE community_discussions
		SET title = $2, content = $3, is_pinned = $4, is_resolved = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return r.pool.QueryRow(ctx, query,
		discussion.ID, discussion.Title, discussion.Content, discussion.IsPinned, discussion.IsResolved,
	).Scan(&discussion.UpdatedAt)
}

// DeleteDiscussion deletes a discussion thread
func (r *CommunityRepository) DeleteDiscussion(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM community_discussions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
