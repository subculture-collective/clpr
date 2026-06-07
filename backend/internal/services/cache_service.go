package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// CacheService handles all caching operations
type CacheService struct {
	redis *redispkg.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redispkg.Client) *CacheService {
	return &CacheService{
		redis: redis,
	}
}

// Cache key builders
const (
	// Feed cache keys
	KeyFeedHot     = "feed:hot:page:%d"
	KeyFeedTop     = "feed:top:%s:page:%d" // timeframe, page
	KeyFeedNew     = "feed:new:page:%d"
	KeyFeedGame    = "feed:game:%s:%s:page:%d"    // gameId, sort, page
	KeyFeedCreator = "feed:creator:%s:%s:page:%d" // creatorId, sort, page

	// Clip cache keys
	KeyClip         = "clip:%s"               // clipId
	KeyClipVotes    = "clip:%s:votes"         // clipId
	KeyClipComments = "clip:%s:comment_count" // clipId

	// Comment cache keys
	KeyCommentTree = "comments:clip:%s:%s" // clipId, sort
	KeyComment     = "comment:%s"          // commentId

	// Metadata cache keys
	KeyGame    = "game:%s" // gameId
	KeyUser    = "user:%s" // userId
	KeyTagsAll = "tags:all"

	// Search cache keys
	KeySearch        = "search:%s:%s:page:%d"  // query, filters, page
	KeySearchSuggest = "search:suggestions:%s" // query

	// Session keys
	KeySession      = "session:%s"       // sessionId
	KeyRefreshToken = "refresh_token:%s" // tokenId

	// Rate limit keys
	KeyRateLimit = "ratelimit:%s:%s" // endpoint, identifier

	// Lock keys
	KeyLock = "lock:%s" // resource
)

// Cache TTL constants
const (
	TTLFeedHot     = 5 * time.Minute
	TTLFeedTop     = 15 * time.Minute
	TTLFeedNew     = 2 * time.Minute
	TTLFeedGame    = 10 * time.Minute
	TTLFeedCreator = 10 * time.Minute

	TTLClip         = 1 * time.Hour
	TTLClipVotes    = 5 * time.Minute
	TTLClipComments = 10 * time.Minute

	TTLCommentTree = 10 * time.Minute
	TTLComment     = 15 * time.Minute

	TTLGame = 24 * time.Hour
	TTLUser = 1 * time.Hour
	TTLTags = 1 * time.Hour

	TTLSearch        = 5 * time.Minute
	TTLSearchSuggest = 1 * time.Hour

	TTLSession      = 7 * 24 * time.Hour // 7 days
	TTLRefreshToken = 7 * 24 * time.Hour // 7 days

	TTLLock = 30 * time.Second
)

// Feed Caching

// GetFeedHot retrieves hot feed from cache
func (s *CacheService) GetFeedHot(ctx context.Context, page int) ([]models.Clip, error) {
	key := fmt.Sprintf(KeyFeedHot, page)
	var clips []models.Clip
	err := s.redis.GetJSON(ctx, key, &clips)
	return clips, err
}

// SetFeedHot stores hot feed in cache
func (s *CacheService) SetFeedHot(ctx context.Context, page int, clips []models.Clip) error {
	key := fmt.Sprintf(KeyFeedHot, page)
	return s.redis.SetJSON(ctx, key, clips, TTLFeedHot)
}

// InvalidateFeedHot clears hot feed cache
func (s *CacheService) InvalidateFeedHot(ctx context.Context) error {
	return s.redis.DeletePattern(ctx, "feed:hot:*")
}

// GetFeedTop retrieves top feed from cache
func (s *CacheService) GetFeedTop(ctx context.Context, timeframe string, page int) ([]models.Clip, error) {
	key := fmt.Sprintf(KeyFeedTop, timeframe, page)
	var clips []models.Clip
	err := s.redis.GetJSON(ctx, key, &clips)
	return clips, err
}

// SetFeedTop stores top feed in cache
func (s *CacheService) SetFeedTop(ctx context.Context, timeframe string, page int, clips []models.Clip) error {
	key := fmt.Sprintf(KeyFeedTop, timeframe, page)
	return s.redis.SetJSON(ctx, key, clips, TTLFeedTop)
}

// InvalidateFeedTop clears top feed cache
func (s *CacheService) InvalidateFeedTop(ctx context.Context) error {
	return s.redis.DeletePattern(ctx, "feed:top:*")
}

// GetFeedNew retrieves new feed from cache
func (s *CacheService) GetFeedNew(ctx context.Context, page int) ([]models.Clip, error) {
	key := fmt.Sprintf(KeyFeedNew, page)
	var clips []models.Clip
	err := s.redis.GetJSON(ctx, key, &clips)
	return clips, err
}

// SetFeedNew stores new feed in cache
func (s *CacheService) SetFeedNew(ctx context.Context, page int, clips []models.Clip) error {
	key := fmt.Sprintf(KeyFeedNew, page)
	return s.redis.SetJSON(ctx, key, clips, TTLFeedNew)
}

// InvalidateFeedNew clears new feed cache
func (s *CacheService) InvalidateFeedNew(ctx context.Context) error {
	return s.redis.DeletePattern(ctx, "feed:new:*")
}

// GetFeedGame retrieves game feed from cache
func (s *CacheService) GetFeedGame(ctx context.Context, gameID, sort string, page int) ([]models.Clip, error) {
	key := fmt.Sprintf(KeyFeedGame, gameID, sort, page)
	var clips []models.Clip
	err := s.redis.GetJSON(ctx, key, &clips)
	return clips, err
}

// SetFeedGame stores game feed in cache
func (s *CacheService) SetFeedGame(ctx context.Context, gameID, sort string, page int, clips []models.Clip) error {
	key := fmt.Sprintf(KeyFeedGame, gameID, sort, page)
	return s.redis.SetJSON(ctx, key, clips, TTLFeedGame)
}

// InvalidateFeedGame clears game-specific feed cache
func (s *CacheService) InvalidateFeedGame(ctx context.Context, gameID string) error {
	pattern := fmt.Sprintf("feed:game:%s:*", gameID)
	return s.redis.DeletePattern(ctx, pattern)
}

// GetFeedCreator retrieves creator feed from cache
func (s *CacheService) GetFeedCreator(ctx context.Context, creatorID, sort string, page int) ([]models.Clip, error) {
	key := fmt.Sprintf(KeyFeedCreator, creatorID, sort, page)
	var clips []models.Clip
	err := s.redis.GetJSON(ctx, key, &clips)
	return clips, err
}

// SetFeedCreator stores creator feed in cache
func (s *CacheService) SetFeedCreator(ctx context.Context, creatorID, sort string, page int, clips []models.Clip) error {
	key := fmt.Sprintf(KeyFeedCreator, creatorID, sort, page)
	return s.redis.SetJSON(ctx, key, clips, TTLFeedCreator)
}

// InvalidateFeedCreator clears creator-specific feed cache
func (s *CacheService) InvalidateFeedCreator(ctx context.Context, creatorID string) error {
	pattern := fmt.Sprintf("feed:creator:%s:*", creatorID)
	return s.redis.DeletePattern(ctx, pattern)
}

// Clip Caching

// GetClip retrieves clip from cache
func (s *CacheService) GetClip(ctx context.Context, clipID uuid.UUID) (*models.Clip, error) {
	key := fmt.Sprintf(KeyClip, clipID.String())
	var clip models.Clip
	err := s.redis.GetJSON(ctx, key, &clip)
	if err != nil {
		return nil, err
	}
	return &clip, nil
}

// SetClip stores clip in cache
func (s *CacheService) SetClip(ctx context.Context, clip *models.Clip) error {
	key := fmt.Sprintf(KeyClip, clip.ID.String())
	return s.redis.SetJSON(ctx, key, clip, TTLClip)
}

// InvalidateClip removes clip from cache
func (s *CacheService) InvalidateClip(ctx context.Context, clipID uuid.UUID) error {
	key := fmt.Sprintf(KeyClip, clipID.String())
	return s.redis.Delete(ctx, key)
}

// GetClipVotes retrieves vote count from cache
func (s *CacheService) GetClipVotes(ctx context.Context, clipID uuid.UUID) (int, error) {
	key := fmt.Sprintf(KeyClipVotes, clipID.String())
	val, err := s.redis.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	count64, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse vote count: %w", err)
	}
	return int(count64), nil
}

// SetClipVotes stores vote count in cache
func (s *CacheService) SetClipVotes(ctx context.Context, clipID uuid.UUID, voteScore int) error {
	key := fmt.Sprintf(KeyClipVotes, clipID.String())
	return s.redis.Set(ctx, key, voteScore, TTLClipVotes)
}

// IncrementClipVotes increments vote count
func (s *CacheService) IncrementClipVotes(ctx context.Context, clipID uuid.UUID, delta int64) error {
	key := fmt.Sprintf(KeyClipVotes, clipID.String())
	_, err := s.redis.IncrementBy(ctx, key, delta)
	if err != nil {
		return err
	}
	return s.redis.Expire(ctx, key, TTLClipVotes)
}

// GetClipCommentCount retrieves comment count from cache
func (s *CacheService) GetClipCommentCount(ctx context.Context, clipID uuid.UUID) (int, error) {
	key := fmt.Sprintf(KeyClipComments, clipID.String())
	val, err := s.redis.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	count64, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse comment count: %w", err)
	}
	return int(count64), nil
}

// SetClipCommentCount stores comment count in cache
func (s *CacheService) SetClipCommentCount(ctx context.Context, clipID uuid.UUID, count int) error {
	key := fmt.Sprintf(KeyClipComments, clipID.String())
	return s.redis.Set(ctx, key, count, TTLClipComments)
}

// IncrementClipCommentCount increments comment count
func (s *CacheService) IncrementClipCommentCount(ctx context.Context, clipID uuid.UUID) error {
	key := fmt.Sprintf(KeyClipComments, clipID.String())
	_, err := s.redis.Increment(ctx, key)
	if err != nil {
		return err
	}
	return s.redis.Expire(ctx, key, TTLClipComments)
}

// Comment Caching

// GetCommentTree retrieves comment tree from cache
func (s *CacheService) GetCommentTree(ctx context.Context, clipID uuid.UUID, sort string) ([]models.Comment, error) {
	key := fmt.Sprintf(KeyCommentTree, clipID.String(), sort)
	var comments []models.Comment
	err := s.redis.GetJSON(ctx, key, &comments)
	return comments, err
}

// SetCommentTree stores comment tree in cache
func (s *CacheService) SetCommentTree(ctx context.Context, clipID uuid.UUID, sort string, comments []models.Comment) error {
	key := fmt.Sprintf(KeyCommentTree, clipID.String(), sort)
	return s.redis.SetJSON(ctx, key, comments, TTLCommentTree)
}

// InvalidateCommentTree clears comment tree cache
func (s *CacheService) InvalidateCommentTree(ctx context.Context, clipID uuid.UUID) error {
	pattern := fmt.Sprintf("comments:clip:%s:*", clipID.String())
	return s.redis.DeletePattern(ctx, pattern)
}

// GetComment retrieves individual comment from cache
func (s *CacheService) GetComment(ctx context.Context, commentID uuid.UUID) (*models.Comment, error) {
	key := fmt.Sprintf(KeyComment, commentID.String())
	var comment models.Comment
	err := s.redis.GetJSON(ctx, key, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// SetComment stores individual comment in cache
func (s *CacheService) SetComment(ctx context.Context, comment *models.Comment) error {
	key := fmt.Sprintf(KeyComment, comment.ID.String())
	return s.redis.SetJSON(ctx, key, comment, TTLComment)
}

// InvalidateComment removes comment from cache
func (s *CacheService) InvalidateComment(ctx context.Context, commentID uuid.UUID) error {
	key := fmt.Sprintf(KeyComment, commentID.String())
	return s.redis.Delete(ctx, key)
}

// Metadata Caching

// GetGame retrieves game from cache
func (s *CacheService) GetGame(ctx context.Context, gameID string) (interface{}, error) {
	key := fmt.Sprintf(KeyGame, gameID)
	var game interface{}
	err := s.redis.GetJSON(ctx, key, &game)
	return game, err
}

// SetGame stores game in cache
func (s *CacheService) SetGame(ctx context.Context, gameID string, game interface{}) error {
	key := fmt.Sprintf(KeyGame, gameID)
	return s.redis.SetJSON(ctx, key, game, TTLGame)
}

// GetUser retrieves user from cache
func (s *CacheService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	key := fmt.Sprintf(KeyUser, userID.String())
	var user models.User
	err := s.redis.GetJSON(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetUser stores user in cache
func (s *CacheService) SetUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf(KeyUser, user.ID.String())
	return s.redis.SetJSON(ctx, key, user, TTLUser)
}

// InvalidateUser removes user from cache
func (s *CacheService) InvalidateUser(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(KeyUser, userID.String())
	return s.redis.Delete(ctx, key)
}

// GetAllTags retrieves all tags from cache
func (s *CacheService) GetAllTags(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.redis.GetJSON(ctx, KeyTagsAll, &tags)
	return tags, err
}

// SetAllTags stores all tags in cache
func (s *CacheService) SetAllTags(ctx context.Context, tags []models.Tag) error {
	return s.redis.SetJSON(ctx, KeyTagsAll, tags, TTLTags)
}

// InvalidateAllTags clears tags cache
func (s *CacheService) InvalidateAllTags(ctx context.Context) error {
	return s.redis.Delete(ctx, KeyTagsAll)
}

// Search Caching

// GetSearchResults retrieves search results from cache
func (s *CacheService) GetSearchResults(ctx context.Context, query, filters string, page int) (interface{}, error) {
	key := fmt.Sprintf(KeySearch, query, filters, page)
	var results interface{}
	err := s.redis.GetJSON(ctx, key, &results)
	return results, err
}

// SetSearchResults stores search results in cache
func (s *CacheService) SetSearchResults(ctx context.Context, query, filters string, page int, results interface{}) error {
	key := fmt.Sprintf(KeySearch, query, filters, page)
	return s.redis.SetJSON(ctx, key, results, TTLSearch)
}

// GetSearchSuggestions retrieves autocomplete suggestions from cache
func (s *CacheService) GetSearchSuggestions(ctx context.Context, query string) ([]string, error) {
	key := fmt.Sprintf(KeySearchSuggest, query)
	var suggestions []string
	err := s.redis.GetJSON(ctx, key, &suggestions)
	return suggestions, err
}

// SetSearchSuggestions stores autocomplete suggestions in cache
func (s *CacheService) SetSearchSuggestions(ctx context.Context, query string, suggestions []string) error {
	key := fmt.Sprintf(KeySearchSuggest, query)
	return s.redis.SetJSON(ctx, key, suggestions, TTLSearchSuggest)
}

// Session Storage

// GetSession retrieves session data from cache
func (s *CacheService) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	key := fmt.Sprintf(KeySession, sessionID)
	var session map[string]interface{}
	err := s.redis.GetJSON(ctx, key, &session)
	return session, err
}

// SetSession stores session data in cache
func (s *CacheService) SetSession(ctx context.Context, sessionID string, session map[string]interface{}) error {
	key := fmt.Sprintf(KeySession, sessionID)
	return s.redis.SetJSON(ctx, key, session, TTLSession)
}

// DeleteSession removes session from cache
func (s *CacheService) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf(KeySession, sessionID)
	return s.redis.Delete(ctx, key)
}

// GetRefreshToken retrieves refresh token data from cache
func (s *CacheService) GetRefreshToken(ctx context.Context, tokenID string) (map[string]interface{}, error) {
	key := fmt.Sprintf(KeyRefreshToken, tokenID)
	var token map[string]interface{}
	err := s.redis.GetJSON(ctx, key, &token)
	return token, err
}

// SetRefreshToken stores refresh token data in cache
func (s *CacheService) SetRefreshToken(ctx context.Context, tokenID string, token map[string]interface{}) error {
	key := fmt.Sprintf(KeyRefreshToken, tokenID)
	return s.redis.SetJSON(ctx, key, token, TTLRefreshToken)
}

// DeleteRefreshToken removes refresh token from cache
func (s *CacheService) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf(KeyRefreshToken, tokenID)
	return s.redis.Delete(ctx, key)
}

// Distributed Locking

// AcquireLock attempts to acquire a distributed lock
func (s *CacheService) AcquireLock(ctx context.Context, resource string, value string) (bool, error) {
	key := fmt.Sprintf(KeyLock, resource)
	return s.redis.SetNX(ctx, key, value, TTLLock)
}

// ReleaseLock releases a distributed lock
func (s *CacheService) ReleaseLock(ctx context.Context, resource string) error {
	key := fmt.Sprintf(KeyLock, resource)
	return s.redis.Delete(ctx, key)
}

// ExtendLock extends the TTL of a lock
func (s *CacheService) ExtendLock(ctx context.Context, resource string) error {
	key := fmt.Sprintf(KeyLock, resource)
	return s.redis.Expire(ctx, key, TTLLock)
}

// Smart Invalidation

// InvalidateOnNewClip invalidates caches when a new clip is created
func (s *CacheService) InvalidateOnNewClip(ctx context.Context, clip *models.Clip) error {
	// Clear hot feed
	if err := s.InvalidateFeedHot(ctx); err != nil {
		return err
	}

	// Clear new feed
	if err := s.InvalidateFeedNew(ctx); err != nil {
		return err
	}

	// Clear game feed if game ID is present
	if clip.GameID != nil && *clip.GameID != "" {
		if err := s.InvalidateFeedGame(ctx, *clip.GameID); err != nil {
			return err
		}
	}

	// Clear creator feed if creator ID is present
	if clip.CreatorID != nil && *clip.CreatorID != "" {
		if err := s.InvalidateFeedCreator(ctx, *clip.CreatorID); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateOnVote invalidates caches when a vote is cast
func (s *CacheService) InvalidateOnVote(ctx context.Context, clipID uuid.UUID) error {
	// Clear hot feed
	if err := s.InvalidateFeedHot(ctx); err != nil {
		return err
	}

	// Clear top feed
	if err := s.InvalidateFeedTop(ctx); err != nil {
		return err
	}

	// Clear clip votes cache
	key := fmt.Sprintf(KeyClipVotes, clipID.String())
	if err := s.redis.Delete(ctx, key); err != nil {
		return err
	}

	// Invalidate the clip itself
	return s.InvalidateClip(ctx, clipID)
}

// InvalidateOnComment invalidates caches when a comment is added
func (s *CacheService) InvalidateOnComment(ctx context.Context, clipID uuid.UUID) error {
	// Clear comment tree
	if err := s.InvalidateCommentTree(ctx, clipID); err != nil {
		return err
	}

	// Clear comment count
	key := fmt.Sprintf(KeyClipComments, clipID.String())
	return s.redis.Delete(ctx, key)
}

// Publish cache invalidation event
func (s *CacheService) PublishInvalidation(ctx context.Context, event string, data interface{}) error {
	return s.redis.Publish(ctx, "cache:invalidation", fmt.Sprintf("%s:%v", event, data))
}
