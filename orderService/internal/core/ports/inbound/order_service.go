package inbound

import (
	"context"
	"test-project/orderService/internal/core/domain"
)


type CreateOrderInput struct{
	UserID   string
    MarketID string
    Type     domain.OrderType
    Price    domain.Money
    Quantity float64
}

type CreateOrderOutput struct{
	OrderID string
    Status  domain.OrderStatus
}

type GetOrderStatusInput struct{
	OrderID string
    UserID  string
}

type OrderService interface {
	CreateOrder(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error)
	GetOrderStatus(ctx context.Context, input GetOrderStatusInput) (*domain.Order, error)
}