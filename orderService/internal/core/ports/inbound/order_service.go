package inbound

import (
	"context"
	"test-project/orderService/internal/core/domain"
	"test-project/shared/money"
)


type CreateOrderInput struct {
	UserID         string 
	MarketID       string
	Type           domain.OrderType
	Price          money.Money
	Quantity       int64 
	IdempotencyKey string 
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