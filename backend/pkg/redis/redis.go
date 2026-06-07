package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// Client wraps the Redis client
type Client struct {
	client *redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	return NewClientWithTracing(cfg, false)
}

// NewClientWithTracing creates a new Redis client with optional tracing
func NewClientWithTracing(cfg *config.RedisConfig, enableTracing bool) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Add tracing if enabled
	if enableTracing {
		if err := redisotel.InstrumentTracing(rdb); err != nil {
			return nil, fmt.Errorf("failed to instrument Redis tracing: %w", err)
		}
		log.Println("Redis tracing enabled")
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connection established successfully")

	return &Client{client: rdb}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Set stores a value with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// SetJSON stores a JSON-serialized value with expiration
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetJSON retrieves and unmarshals a JSON value
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *Client) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	pipe := c.client.Pipeline()

	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Increment increments a counter
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrementBy increments a counter by a specific amount
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decrement decrements a counter
func (c *Client) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// Expire sets expiration on a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// SetNX sets a value only if it doesn't exist (for locking)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// GetWithTTL retrieves a value and its remaining TTL
func (c *Client) GetWithTTL(ctx context.Context, key string) (string, time.Duration, error) {
	pipe := c.client.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return "", 0, err
	}

	value, err := getCmd.Result()
	if err != nil {
		return "", 0, err
	}

	ttl, err := ttlCmd.Result()
	if err != nil {
		return "", 0, err
	}

	return value, ttl, nil
}

// MGet retrieves multiple values at once
func (c *Client) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.client.MGet(ctx, keys...).Result()
}

// MSet sets multiple key-value pairs at once
func (c *Client) MSet(ctx context.Context, pairs ...interface{}) error {
	return c.client.MSet(ctx, pairs...).Err()
}

// ZAdd adds a member with score to a sorted set
func (c *Client) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return c.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZRange retrieves members from a sorted set by rank range
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange retrieves members from a sorted set in reverse order
func (c *Client) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScores retrieves members with scores from a sorted set in reverse order
func (c *Client) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// ZIncrBy increments the score of a member in a sorted set
func (c *Client) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	return c.client.ZIncrBy(ctx, key, increment, member).Result()
}

// ZRem removes a member from a sorted set
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.ZRem(ctx, key, members...).Err()
}

// Publish publishes a message to a channel (for cache invalidation)
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.client.Publish(ctx, channel, message).Err()
}

// Subscribe creates a subscription to a channel
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// Pipeline returns a pipeline for batching commands
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// HSet sets a hash field value
func (c *Client) HSet(ctx context.Context, key, field string, value interface{}) error {
	return c.client.HSet(ctx, key, field, value).Err()
}

// HGet gets a hash field value
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HIncrBy increments a hash field by a value
func (c *Client) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return c.client.HIncrBy(ctx, key, field, incr).Result()
}

// HealthCheck checks if Redis is accessible
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetStats returns Redis server stats
func (c *Client) GetStats(ctx context.Context) (map[string]string, error) {
	info, err := c.client.Info(ctx, "stats", "memory", "keyspace").Result()
	if err != nil {
		return nil, err
	}

	// Parse the info string into a map
	stats := make(map[string]string)
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		// Parse key:value
		if idx := strings.IndexByte(line, ':'); idx != -1 {
			key := line[:idx]
			value := line[idx+1:]
			stats[key] = strings.TrimSpace(value)
		}
	}

	return stats, nil
}

// Keys returns all keys matching a pattern
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// GetClient returns the underlying redis client
// WARNING: This breaks encapsulation and should only be used when absolutely necessary.
// Direct access bypasses wrapper logic and makes it harder to change the Redis implementation.
// Prefer adding specific methods to the Client wrapper for new operations.
// Currently used by: embedding service (requires direct client for caching operations)
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// ListPush appends a value to a Redis list using the RPUSH operation
func (c *Client) ListPush(ctx context.Context, key string, value interface{}) error {
	return c.client.RPush(ctx, key, value).Err()
}

// ListRange retrieves a range of elements from a Redis list (LRANGE operation).
// The range is inclusive: both start and stop indices are included.
func (c *Client) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// ListLen returns the length of a Redis list (LLEN operation)
func (c *Client) ListLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// ListTrim trims a list to the specified range (LTRIM operation)
func (c *Client) ListTrim(ctx context.Context, key string, start, stop int64) error {
	return c.client.LTrim(ctx, key, start, stop).Err()
}

// ListPop removes and returns the first element from a Redis list (LPOP operation)
// Returns an empty string and redis.Nil error if the list is empty
func (c *Client) ListPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// SetAdd adds one or more members to a Redis set (SADD operation)
func (c *Client) SetAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SetCard returns the cardinality (number of elements) in a Redis set (SCARD operation)
func (c *Client) SetCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key).Result()
}

// TTL returns the remaining time to live of a key in seconds.
// Returns 0 if the key doesn't exist or has no expiration.
func (c *Client) TTL(ctx context.Context, key string) (int64, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(ttl.Seconds()), nil
}

// SetMembers returns all members of a Redis set (SMEMBERS operation)
func (c *Client) SetMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SetIsMember checks if a member exists in a Redis set (SISMEMBER operation)
func (c *Client) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}
