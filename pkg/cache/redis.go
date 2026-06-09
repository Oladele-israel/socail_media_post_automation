package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // empty string = no password (local dev)
		DB:       0,                           // Redis has 16 logical DBs (0-15), use 0 by default
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) SetRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s", token) // namespaced key — best practice
	return r.client.Set(ctx, key, userID, ttl).Err()
}

func (r *RedisClient) GetRefreshToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh:%s", token)

	userID, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// redis.Nil means key doesn't exist (expired or never set)
		// This is Go's pattern — no exceptions, explicit error values
		return "", fmt.Errorf("refresh token not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	return userID, nil
}

// DeleteRefreshToken invalidates a token on logout
func (r *RedisClient) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh:%s", token)
	return r.client.Del(ctx, key).Err()
}

// BlacklistAccessToken blocks a JWT before it naturally expires (for logout)
// This is an optional but important security layer
func (r *RedisClient) BlacklistAccessToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return r.client.Set(ctx, key, "1", ttl).Err()
}

// IsTokenBlacklisted checks if an access token has been revoked
func (r *RedisClient) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// ─────────────────────────────────────────
// OAuth State methods
// ─────────────────────────────────────────

// SetOAuthState stores: "oauth_state:{state}" → "platform:userID" with TTL
// The state is a random UUID used to prevent CSRF attacks
// Value format: "linkedin:user-uuid-here"
func (r *RedisClient) SetOAuthState(ctx context.Context, state, value string, ttl time.Duration) error {
	key := fmt.Sprintf("oauth_state:%s", state)

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store oauth state: %w", err)
	}

	return nil
}

// GetOAuthState retrieves the stored value for an oauth state token
// Returns the "platform:userID" string stored during connect
func (r *RedisClient) GetOAuthState(ctx context.Context, state string) (string, error) {
	key := fmt.Sprintf("oauth_state:%s", state)

	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist — either expired (10 min TTL) or never existed
		// This could be a CSRF attack attempt
		return "", fmt.Errorf("oauth state not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get oauth state: %w", err)
	}

	return value, nil
}

// DeleteOAuthState removes the state after it has been used
// OAuth states are one-time use — delete immediately after verification
func (r *RedisClient) DeleteOAuthState(ctx context.Context, state string) error {
	key := fmt.Sprintf("oauth_state:%s", state)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete oauth state: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────
// Helper
// ─────────────────────────────────────────

// getEnv reads an env variable with a fallback default
// keeps NewRedisClient clean — no os.Getenv noise inline
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
