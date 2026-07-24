package domain

import (
	commonv1 "grpc-exchange/gen/common"
	"time"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "PENDING"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

type Order struct {
	OrderID  string
	UserID   string
	MarketID string
	Side     OrderSide
	Price    commonv1.Money
	Quantity commonv1.Decimal
	FilledQuantity commonv1.Decimal
	Status OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}