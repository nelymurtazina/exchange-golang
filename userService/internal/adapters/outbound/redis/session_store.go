package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports/outbound"

	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	client *redis.Client
}

func NewSessionStore(client *redis.Client) ports.RefreshTokenStore {
	return &SessionStore{client: client}
}

func (s *SessionStore) Save(ctx context.Context, userID string, jti string, ttl time.Duration) error {
	key := "session:" + jti
	return s.client.Set(ctx, key, userID, ttl).Err()
}

func (s *SessionStore) Get(ctx context.Context, userID string) (*domain.User, error) {
	key := fmt.Sprintf("market:%s", userID)

	jsonData, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil 
		}
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var market domain.User

	err = json.Unmarshal([]byte(jsonData), &market)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal market from cache: %w", err)
	}

	return &market, nil
}

func (s *SessionStore) Set(ctx context.Context, user *domain.User, expiration time.Duration) error {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal market for cache: %w", err)
	}

	key := fmt.Sprintf("market:%s", user.UserID)

	err = s.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save market to redis: %w", err)
	}

	return nil
}

func (s *SessionStore) GetUserID(ctx context.Context, jti string) (string, error) {
	key := "session:" + jti
	userID, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrInvalidToken 
	}
	if err != nil {
		return "", fmt.Errorf("redis error: %w", err)
	}
	return userID, nil
}

func (s *SessionStore) Delete(ctx context.Context, jti string) error {
	key := "session:" + jti
	return s.client.Del(ctx, key).Err()
}