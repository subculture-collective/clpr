package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

const (
	cacheKeyPrefix = "twitch:"
)

// Cache defines the interface for basic caching operations
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
}

// TwitchCache extends Cache with Twitch-specific methods
type TwitchCache interface {
	Cache
	CachedUser(ctx context.Context, userID string) (*User, error)
	CacheUser(ctx context.Context, user *User, ttl time.Duration) error
	CachedGame(ctx context.Context, gameID string) (*Game, error)
	CacheGame(ctx context.Context, game *Game, ttl time.Duration) error
}

// RedisCache wraps Redis client to implement TwitchCache interface
type RedisCache struct {
	client *redispkg.Client
}

// NewRedisCache creates a new Redis-backed cache
func NewRedisCache(client *redispkg.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Get retrieves a value from cache
func (c *RedisCache) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, false
	}
	return val, true
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) {
	ctx := context.Background()
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		// For complex types, marshal to JSON
		if jsonBytes, err := json.Marshal(v); err == nil {
			strVal = string(jsonBytes)
		} else {
			// Fallback to string formatting only for simple types
			strVal = fmt.Sprintf("%v", v)
		}
	}
	_ = c.client.Set(ctx, key, strVal, ttl)
}

// Delete removes a value from cache
func (c *RedisCache) Delete(key string) {
	ctx := context.Background()
	_ = c.client.Delete(ctx, key)
}

// CachedUser retrieves user data from cache
func (c *RedisCache) CachedUser(ctx context.Context, userID string) (*User, error) {
	cacheKey := fmt.Sprintf("%suser:%s", cacheKeyPrefix, userID)
	val, err := c.client.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CacheUser stores user data in cache
func (c *RedisCache) CacheUser(ctx context.Context, user *User, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("%suser:%s", cacheKeyPrefix, user.ID)
	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cacheKey, string(userData), ttl)
}

// CachedGame retrieves game data from cache
func (c *RedisCache) CachedGame(ctx context.Context, gameID string) (*Game, error) {
	cacheKey := fmt.Sprintf("%sgame:%s", cacheKeyPrefix, gameID)
	val, err := c.client.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var game Game
	if err := json.Unmarshal([]byte(val), &game); err != nil {
		return nil, err
	}

	return &game, nil
}

// CacheGame stores game data in cache
func (c *RedisCache) CacheGame(ctx context.Context, game *Game, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("%sgame:%s", cacheKeyPrefix, game.ID)
	gameData, err := json.Marshal(game)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cacheKey, string(gameData), ttl)
}
