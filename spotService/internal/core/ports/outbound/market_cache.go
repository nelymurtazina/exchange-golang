package outbound

import (
	"context"
	"test-project/spotService/internal/core/domain"
	"time"
)

type MarketCache interface {
	Set(ctx context.Context, market *domain.Market, expiration time.Duration) error
	Get(ctx context.Context, marketID string) (*domain.Market, error)
	Delete(ctx context.Context, marketID string) error
}