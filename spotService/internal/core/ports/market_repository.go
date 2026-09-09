package ports

import (
	"context"
	"test-project/spotService/internal/core/domain"
)

// интерфейс для работы с БД
type MarketRepository interface{
	GetAll(ctx context.Context) ([]*domain.Market, error)
	GetAllActive(ctx context.Context) ([]*domain.Market, error)
	GetByID(ctx context.Context, id string) (*domain.Market, error)
	Create(ctx context.Context, market *domain.Market) error
	Update(ctx context.Context, market *domain.Market) error
	Delete(ctx context.Context, id string) error
}