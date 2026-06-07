package services

import (
	"context"
	"errors"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

const (
	trendingCursorKeyPrefix  = "trending:cursor:"
	trendingGamesCacheKey    = "trending:games:last"
	defaultTrendingCursorTTL = 8 * time.Hour
	defaultTrendingGamesTTL  = 8 * time.Hour
)

// TrendingCursorState holds pagination progress for a game
type TrendingCursorState struct {
	Cursor string `json:"cursor"`
	Page   int    `json:"page"`
}

// TrendingStateStore defines persistence for trending pagination and game lists
type TrendingStateStore interface {
	GetCursor(ctx context.Context, gameID string) (*TrendingCursorState, error)
	SaveCursor(ctx context.Context, gameID string, state *TrendingCursorState) error
	ClearCursor(ctx context.Context, gameID string) error
	SaveGameIDs(ctx context.Context, gameIDs []string) error
	LoadGameIDs(ctx context.Context) ([]string, error)
}

// RedisTrendingStateStore persists trending state in Redis
//
//nolint:revive // exported type needed for wiring
type RedisTrendingStateStore struct {
	client    *redispkg.Client
	cursorTTL time.Duration
	gamesTTL  time.Duration
}

// NewRedisTrendingStateStore constructs a Redis-backed store
func NewRedisTrendingStateStore(client *redispkg.Client) *RedisTrendingStateStore {
	if client == nil {
		return nil
	}

	return &RedisTrendingStateStore{
		client:    client,
		cursorTTL: defaultTrendingCursorTTL,
		gamesTTL:  defaultTrendingGamesTTL,
	}
}

func (s *RedisTrendingStateStore) cursorKey(gameID string) string {
	return trendingCursorKeyPrefix + gameID
}

// GetCursor returns the stored cursor and page for a game, or nil if none exists.
func (s *RedisTrendingStateStore) GetCursor(ctx context.Context, gameID string) (*TrendingCursorState, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}

	var state TrendingCursorState
	if err := s.client.GetJSON(ctx, s.cursorKey(gameID), &state); err != nil {
		if errors.Is(err, redisv9.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if state.Page <= 0 {
		state.Page = 1
	}

	return &state, nil
}

// SaveCursor persists the next cursor/page pair for a game.
func (s *RedisTrendingStateStore) SaveCursor(ctx context.Context, gameID string, state *TrendingCursorState) error {
	if s == nil || s.client == nil || state == nil {
		return nil
	}

	if state.Page <= 0 {
		state.Page = 1
	}

	return s.client.SetJSON(ctx, s.cursorKey(gameID), state, s.cursorTTL)
}

// ClearCursor removes stored pagination for a game.
func (s *RedisTrendingStateStore) ClearCursor(ctx context.Context, gameID string) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Delete(ctx, s.cursorKey(gameID))
}

// SaveGameIDs caches the last successful game set for fallback.
func (s *RedisTrendingStateStore) SaveGameIDs(ctx context.Context, gameIDs []string) error {
	if s == nil || s.client == nil || len(gameIDs) == 0 {
		return nil
	}
	return s.client.SetJSON(ctx, trendingGamesCacheKey, gameIDs, s.gamesTTL)
}

// LoadGameIDs retrieves the cached game set if present.
func (s *RedisTrendingStateStore) LoadGameIDs(ctx context.Context) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}

	var ids []string
	if err := s.client.GetJSON(ctx, trendingGamesCacheKey, &ids); err != nil {
		if errors.Is(err, redisv9.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return ids, nil
}
