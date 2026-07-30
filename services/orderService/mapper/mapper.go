package mapper

import (
	orderv1 "grpc-exchange/gen/order"
	"grpc-exchange/services/orderService/domain"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoOrder(order *domain.Order) *orderv1.Order{
	if order == nil{
		return nil
	}

	return &orderv1.Order{
		OrderId: order.OrderID,
		UserId:         order.UserID,
        MarketId:       order.MarketID,
        OrderSide:      convertSide(order.Side),
        Price:          &order.Price,
        Quantity:       &order.Quantity,
        FilledQuantity: &order.FilledQuantity,
        Status:         convertStatus(order.Status),
        CreatedAt:      timestamppb.New(order.CreatedAt),
        UpdatedAt:      timestamppb.New(order.UpdatedAt),
	}
}

// Когда приходит запрос от клиента(FROM PROTO in TO DOMAIN)
func FromProtoCreateOrder(req *orderv1.CreateOrderRequest) *domain.Order {
    if req == nil {
        return nil
    }
    
    return &domain.Order{
        OrderID:  uuid.New().String(),
        UserID:   req.UserId,
        MarketID: req.MarketId,
        Side:     ConvertProtoSide(req.OrderSide),
        Price:    *req.Price,    
        Quantity: *req.Quantity, 
        Status:   domain.OrderStatusPending,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

func convertSide(side domain.OrderSide) orderv1.OrderSide{
	switch side {
    case domain.OrderSideBuy:
        return orderv1.OrderSide_ORDER_SIDE_BUY
    case domain.OrderSideSell:
        return orderv1.OrderSide_ORDER_SIDE_SELL
    default:
        return orderv1.OrderSide_ORDER_SIDE_UNSPECIFIED
    }
}

func convertStatus(status domain.OrderStatus) orderv1.OrderStatus {
    switch status {
    case domain.OrderStatusPending:
        return orderv1.OrderStatus_ORDER_STATUS_PENDING
    case domain.OrderStatusFilled:
        return orderv1.OrderStatus_ORDER_STATUS_FILLED
    case domain.OrderStatusPartiallyFilled:
        return orderv1.OrderStatus_ORDER_STATUS_PARTIALLY_FILLED
    case domain.OrderStatusCancelled:
        return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
    case domain.OrderStatusRejected:
        return orderv1.OrderStatus_ORDER_STATUS_REJECTED
    default:
        return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
    }
}


func ConvertProtoSide(side orderv1.OrderSide) domain.OrderSide {
    switch side {
    case orderv1.OrderSide_ORDER_SIDE_BUY:
        return domain.OrderSideBuy
    case orderv1.OrderSide_ORDER_SIDE_SELL:
        return domain.OrderSideSell
    default:
        return domain.OrderSideBuy 
    }
}

func convertDomainSide(side domain.OrderSide) orderv1.OrderSide {
    switch side {
    case domain.OrderSideBuy:
        return orderv1.OrderSide_ORDER_SIDE_BUY
    case domain.OrderSideSell:
        return orderv1.OrderSide_ORDER_SIDE_SELL
    default:
        return orderv1.OrderSide_ORDER_SIDE_UNSPECIFIED
    }
}

func convertDomainStatus(status domain.OrderStatus) orderv1.OrderStatus {
    switch status {
    case domain.OrderStatusPending:
        return orderv1.OrderStatus_ORDER_STATUS_PENDING
    case domain.OrderStatusFilled:
        return orderv1.OrderStatus_ORDER_STATUS_FILLED
    case domain.OrderStatusPartiallyFilled:
        return orderv1.OrderStatus_ORDER_STATUS_PARTIALLY_FILLED
    case domain.OrderStatusCancelled:
        return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
    case domain.OrderStatusRejected:
        return orderv1.OrderStatus_ORDER_STATUS_REJECTED
    default:
        return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
    }
}