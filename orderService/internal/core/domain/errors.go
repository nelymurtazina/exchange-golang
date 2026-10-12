package domain

import "errors"

var (
    ErrOrderNotFound      = errors.New("order not found")
    ErrInvalidOrderID     = errors.New("invalid order id")
    ErrInvalidUserID      = errors.New("invalid user id")
    ErrInvalidOrdername   = errors.New("invalid username")
    ErrInvalidMarketID    = errors.New("invalid market id")
    ErrInvalidOrderType   = errors.New("invalid order type")
    ErrInvalidOrderStatus = errors.New("invalid order status")
    ErrInvalidPrice       = errors.New("price must be greater than 0")
    ErrInvalidQuantity    = errors.New("quantity must be greater than 0")
    ErrMarketNotAvailable = errors.New("market is not available for trading")
    ErrUnauthorized       = errors.New("user is not authorized to access this order")
)