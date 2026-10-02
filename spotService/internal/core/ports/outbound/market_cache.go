package outbound

import (
	"context"
	"test-project/spotService/internal/core/domain"
	"time"
)

type MarketCache interface {
	// expiration — это "время жизни" записи. Через этот промежуток времени Redis сам её сотрет.
	Set(ctx context.Context, market *domain.Market, expiration time.Duration) error
	Get(ctx context.Context, marketID string) (*domain.Market, error)
}