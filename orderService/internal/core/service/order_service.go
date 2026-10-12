package service

import (
	// "context"
	// "test-project/orderService/internal/core/ports/inbound"
	"context"
	"errors"
	"fmt"
	"test-project/orderService/internal/core/domain"
	"test-project/orderService/internal/core/ports/inbound"
	"test-project/orderService/internal/core/ports/outbound"
	"time"
)


type OrderService struct{
	repo outbound.OrderRepository
	spotClient outbound.SpotInstrumentClient
}

func NewOrderService(repo outbound.OrderRepository, spotClient outbound.SpotInstrumentClient) *OrderService{
	return &OrderService{
		repo: repo,
		spotClient: spotClient,
	}
}

func (o *OrderService) CreateOrder(ctx context.Context, input inbound.CreateOrderInput)(*inbound.CreateOrderOutput, error){
	if err := domain.ValidateID(input.UserID, domain.ErrInvalidUserID); err != nil {
		return nil, err
	}
	if err := domain.ValidateMarketID(input.MarketID); err != nil {
		return nil, err
	}
	if err := domain.ValidateType(input.Type); err != nil {
		return nil, err
	}
	if err := domain.ValidatePrice(input.Price); err != nil {
		return nil, err
	}
	if err := domain.ValidateQuantity(input.Quantity); err != nil {
		return nil, err
	}

	marketInfo, err := o.spotClient.GetMarket(ctx, input.MarketID)
	if err != nil {
		return nil, fmt.Errorf("failed to check market availability: %w", err)
	}
	if marketInfo == nil || !marketInfo.Enabled {
		return nil, domain.ErrMarketNotAvailable
	}
	orderID := domain.NewOrderID()
	now := time.Now()
	order, err := domain.NewOrder(
		orderID,
		input.UserID,
		input.MarketID,
		input.Type,
		input.Price,
		input.Quantity,
		domain.StatusCreated, 
		now,
		now,
	)
	if err != nil{
		return nil, err
	}
	if err := o.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order to repository: %w", err)
	}
	return &inbound.CreateOrderOutput{
		OrderID: order.OrderID,
		Status:  order.Status,
	}, nil
}

func (o *OrderService) GetOrderStatus(ctx context.Context, input inbound.GetOrderStatusInput)(*domain.Order, error){
	if err := domain.ValidateID(input.OrderID, domain.ErrInvalidOrderID); err != nil {
		return nil, err
	}
	if err := domain.ValidateID(input.UserID, domain.ErrInvalidUserID); err != nil {
		return nil, err
	}
	order, err := o.repo.GetByID(ctx, input.OrderID)
	if err != nil{
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get status order: %w",err)
	}
	if order.UserID != input.UserID {
		return nil, domain.ErrUnauthorized 
	}
	return order, nil
}