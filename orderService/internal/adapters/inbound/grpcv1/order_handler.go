package grpcv1

import (
	"context"
	"errors"
	pb "test-project/api/gen/order"
	commonv1 "test-project/api/gen/common"
	"test-project/orderService/internal/core/domain"
	"test-project/orderService/internal/core/ports/inbound"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	service inbound.OrderService
}

func NewOrderHandler(service inbound.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error){
	output, err := h.service.CreateOrder(ctx, inbound.CreateOrderInput{
		UserID:   req.UserId,
		MarketID: req.MarketId,
		Type:     mapOrderType(req.OrderType),
		Price:    mapProtoMoneyToDomain(req.Price),
		Quantity: req.Quantity,
	})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.CreateOrderResponse{
		OrderId: output.OrderID,
		Status:  mapStatus(output.Status),
		Message: "Order created successfully",
	}, nil
}

func (h *OrderHandler) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponse, error) {
	order, err := h.service.GetOrderStatus(ctx, inbound.GetOrderStatusInput{
		OrderID: req.OrderId,
		UserID:  req.UserId,
	})
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.GetOrderStatusResponse{
		Order: mapDomainOrderToProto(order),
	}, nil
}

func mapOrderType(orderType pb.OrderType) domain.OrderType {
	switch orderType {
	case pb.OrderType_ORDER_TYPE_BUY:
		return domain.OrderTypeBuy
	case pb.OrderType_ORDER_TYPE_SELL:
		return domain.OrderTypeSell
	default:
		return "" 
	}
}

func mapStatus(status domain.OrderStatus) pb.OrderStatus {
	switch status {
	case domain.StatusCreated:
		return pb.OrderStatus_ORDER_STATUS_CREATED
	case domain.StatusPending:
		return pb.OrderStatus_ORDER_STATUS_PENDING
	case domain.StatusFilled:
		return pb.OrderStatus_ORDER_STATUS_FILLED
	case domain.StatusCancelled:
		return pb.OrderStatus_ORDER_STATUS_CANCELLED
	case domain.StatusRejected:
		return pb.OrderStatus_ORDER_STATUS_REJECTED
	default:
		return pb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func mapProtoMoneyToDomain(money *commonv1.Money) domain.Money {
	if money == nil || money.Amount == nil {
		return domain.Money{}
	}
	return domain.Money{
		Units:        money.Amount.Units,
		Nanos:        money.Amount.Nanos,
		CurrencyCode: money.CurrencyCode,
	}
}

func mapDomainOrderToProto(order *domain.Order) *pb.Order {
	if order == nil {
		return nil
	}
	return &pb.Order{
		OrderId:   order.OrderID,
		UserId:    order.UserID,
		MarketId:  order.MarketID,
		OrderType: mapOrderTypeToProto(order.Type),
		Price:     mapDomainMoneyToProto(order.Price),
		Quantity:  order.Quantity,
		Status:    mapStatus(order.Status),
		CreatedAt: timestamppb.New(order.CreatedAt),
	}
}

func mapOrderTypeToProto(orderType domain.OrderType) pb.OrderType {
	switch orderType {
	case domain.OrderTypeBuy:
		return pb.OrderType_ORDER_TYPE_BUY
	case domain.OrderTypeSell:
		return pb.OrderType_ORDER_TYPE_SELL
	default:
		return pb.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

func mapDomainMoneyToProto(money domain.Money) *commonv1.Money {
	return &commonv1.Money{
		Amount: &commonv1.Decimal{
			Units: money.Units,
			Nanos: money.Nanos,
		},
		CurrencyCode: money.CurrencyCode,
	}
}


func mapDomainErrorToGRPC(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return status.Error(codes.NotFound, "order not found")
	case errors.Is(err, domain.ErrInvalidOrderID),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidMarketID),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidOrderType),
		errors.Is(err, domain.ErrInvalidOrderStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrMarketNotAvailable):
		return status.Error(codes.FailedPrecondition, "market is not available for trading")
	case errors.Is(err, domain.ErrUnauthorized):
		return status.Error(codes.PermissionDenied, "user is not authorized to access this order")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}