package outbound

import (
	"context"
	"test-project/orderService/internal/core/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, orderID string) (*domain.Order, error)
}