package cache

import (
    "context"
    "fmt"
    "time"

    "grpc-exchange/shered/config"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(cfg config.RedisConfig) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    return &RedisCache{client: rdb}
}

func (c *RedisCache) Ping(ctx context.Context) error {
    return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
    return c.client.SAdd(ctx, key, members...).Err()
}

func (c *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
    return c.client.SMembers(ctx, key).Result()
}

func (c *RedisCache) Close() error {
    return c.client.Close()
}