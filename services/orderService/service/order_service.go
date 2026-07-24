package service

import (
	"context"
	orderv1 "grpc-exchange/gen/order"
	spotv1 "grpc-exchange/gen/spot"
	"grpc-exchange/services/orderService/mapper"
	"grpc-exchange/services/orderService/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService struct {
    orderv1.UnimplementedOrderServiceServer
    orderRepo  repository.OrderRepository
    spotClient spotv1.SpotInstrumentServiceClient
}

func NewOrderService(
    orderRepo repository.OrderRepository,
    spotClient spotv1.SpotInstrumentServiceClient,
) *OrderService {
    return &OrderService{
        orderRepo:  orderRepo,
        spotClient: spotClient,
    }
}

func (s *OrderService) CreateOrder(
    ctx context.Context,
    req *orderv1.CreateOrderRequest,
) (*orderv1.CreateOrderResponse, error) {
    if req.UserId == "" {
        return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
    }
    if req.MarketId == "" {
        return nil, status.Errorf(codes.InvalidArgument, "market_id is required")
    }
    if req.Price == nil {
        return nil, status.Errorf(codes.InvalidArgument, "price is required")
    }
    if req.Quantity == nil {
        return nil, status.Errorf(codes.InvalidArgument, "quantity is required")
    }

    marketResp, err := s.spotClient.GetMarket(ctx, &spotv1.GetMarketRequest{
        MarketId: req.MarketId,
    })
    if err != nil {
        if status.Code(err) == codes.NotFound {
            return nil, status.Errorf(codes.NotFound, "market not found or inactive")
        }
        return nil, status.Errorf(codes.Internal, "failed to check market: %v", err)
    }
    if marketResp.Market == nil || !marketResp.Market.Enabled {
        return nil, status.Errorf(codes.NotFound, "market not found or inactive")
    }

    order := mapper.FromProtoCreateOrder(req)

    if err := s.orderRepo.Save(ctx, order); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to save order: %v", err)
    }

    return &orderv1.CreateOrderResponse{
        OrderId: order.OrderID,
        Status:  orderv1.OrderStatus_ORDER_STATUS_PENDING,
    }, nil
}

func (s *OrderService) GetOrderStatus(
    ctx context.Context,
    req *orderv1.GetOrderStatusRequest,
) (*orderv1.GetOrderStatusResponse, error) {
    if req.OrderId == "" {
        return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
    }
    if req.UserId == "" {
        return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
    }

    order, err := s.orderRepo.GetByIDAndUser(ctx, req.OrderId, req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
    }
    if order == nil {
        return nil, status.Errorf(codes.NotFound, "order not found")
    }

    return &orderv1.GetOrderStatusResponse{
        Order: mapper.ToProtoOrder(order),
    }, nil
}