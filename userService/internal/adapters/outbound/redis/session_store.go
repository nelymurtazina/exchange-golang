package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"test-project/userService/internal/core/domain"
	ports "test-project/userService/internal/core/ports/outbound"

	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	client *redis.Client
}

func NewSessionStore(client *redis.Client) ports.RefreshTokenStore {
	return &SessionStore{client: client}
}

func (s *SessionStore) Save(ctx context.Context, userID string, jti string, ttl time.Duration) error {
	pipe := s.client.Pipeline()

	pipe.Set(ctx, "session:"+jti, userID, ttl)

	userSessionsKey := "user_sessions:" + userID
	pipe.SAdd(ctx, userSessionsKey, jti)
	pipe.Expire(ctx, userSessionsKey, ttl) 

	_, err := pipe.Exec(ctx)
	return err
}

func (s *SessionStore) Get(ctx context.Context, userID string) (*domain.User, error) {
	key := fmt.Sprintf("user:%s", userID)

	jsonData, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var user domain.User

	err = json.Unmarshal([]byte(jsonData), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user from cache: %w", err)
	}

	return &user, nil
}

func (s *SessionStore) Set(ctx context.Context, user *domain.User, expiration time.Duration) error {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user for cache: %w", err)
	}

	key := fmt.Sprintf("User:%s", user.UserID)

	err = s.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save user to redis: %w", err)
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


func (s *SessionStore) DeleteAllForUser(ctx context.Context, userID string) error {
	userSessionsKey := "user_sessions:" + userID

	jtis, err := s.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return err
	}

	if len(jtis) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()

	for _, jti := range jtis {
		pipe.Del(ctx, "session:"+jti)
	}

	pipe.Del(ctx, userSessionsKey)

	_, err = pipe.Exec(ctx)
	return err
}