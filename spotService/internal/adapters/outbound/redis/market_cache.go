package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"test-project/spotService/internal/core/domain"
	"test-project/spotService/internal/core/ports/outbound"
	"time"

	"github.com/redis/go-redis/v9"
)

type MarketCacheImpl struct {
	client *redis.Client
}

func NewMarketCache(client *redis.Client) outbound.MarketCache {
	return &MarketCacheImpl{
		client: client,
	}
}


func (c *MarketCacheImpl) Get(ctx context.Context, marketID string) (*domain.Market, error) {
	key := fmt.Sprintf("market:%s", marketID)

	jsonData, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil 
		}
		return nil, fmt.Errorf("failed to get market from redis: %w", err)
	}

	var market domain.Market

	err = json.Unmarshal([]byte(jsonData), &market)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal market from cache: %w", err)
	}

	return &market, nil
}

func (c *MarketCacheImpl) Set(ctx context.Context, market *domain.Market, expiration time.Duration) error {
	jsonData, err := json.Marshal(market)
	if err != nil {
		return fmt.Errorf("failed to marshal market for cache: %w", err)
	}

	key := fmt.Sprintf("market:%s", market.MarketID)

	err = c.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save market to redis: %w", err)
	}

	return nil
}