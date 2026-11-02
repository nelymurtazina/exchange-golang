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

	SetMarketsList(ctx context.Context, markets []*domain.Market, expiration time.Duration) error
    GetMarketsList(ctx context.Context) ([]*domain.Market, error)
    DeleteMarketsList(ctx context.Context) error
}