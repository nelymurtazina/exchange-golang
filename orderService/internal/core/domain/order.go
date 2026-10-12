package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

type OrderType string

const (
	OrderTypeBuy  OrderType = "BUY"
	OrderTypeSell OrderType = "SELL"
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusPending   OrderStatus = "PENDING"
	StatusFilled    OrderStatus = "FILLED"
	StatusCancelled OrderStatus = "CANCELLED"
	StatusRejected  OrderStatus = "REJECTED"
)

var ValidUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)


type Money struct {
    Units        int64
    Nanos        int32
    CurrencyCode string
}

type Order struct {
	OrderID   string
	UserID    string
	MarketID  string
	Type      OrderType
	Price 	  Money  
	Quantity  float64
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrder(orderID, userID, marketID string, orderType OrderType, price Money, quantity float64, status OrderStatus,createdAt, updatedAt time.Time) (*Order, error){
	if err := ValidateID(userID, ErrInvalidUserID); err != nil{
		return nil, err
	}
	if err := ValidateID(orderID, ErrInvalidOrderID); err != nil{
		return nil, err
	}
	if err := ValidateID(marketID, ErrInvalidMarketID); err != nil{
		return nil, err
	}

	if err := ValidateType(orderType); err != nil {
		return nil, err
	}
	if err := ValidateStatus(status); err != nil{
		return nil, err
	}
	if err := ValidatePrice(price); err != nil{
		return nil, err
	}
	if err := ValidateQuantity(quantity); err != nil {
		return nil, err
	}
	return &Order{
		OrderID: orderID,
		UserID: userID,
		MarketID: marketID,
		Type: orderType,
		Price: price,
		Quantity: quantity,
		Status: status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func ValidateID(id string, errToReturn error) error{
	if id == ""{
		return errToReturn
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return errToReturn
	}
	return nil
}

func NewOrderID() string{
	return uuid.New().String()
}

func ValidateType(orderType OrderType) error{
	switch orderType {
	case OrderTypeBuy, OrderTypeSell:
		return nil
	default:
		return ErrInvalidOrderType
	}
}

func ValidateStatus(status OrderStatus) error{
	switch status {
	case StatusCreated, StatusPending, StatusFilled, StatusCancelled, StatusRejected:
		return nil
	default:
		return ErrInvalidOrderStatus
	}
}

func ValidatePrice(price Money) error {
    if price.CurrencyCode == "" {
        return ErrInvalidPrice
    }
    if price.Nanos < 0 || price.Nanos > 999999999 {
        return ErrInvalidPrice
    }
    if price.Units < 0 || (price.Units == 0 && price.Nanos == 0) {
        return ErrInvalidPrice
    }
    return nil
}

func ValidateQuantity(quantity float64) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

// проверяет, что market_id имеет формат "BTC-USD"
func ValidateMarketID(marketID string) error {
	if marketID == "" {
		return ErrInvalidMarketID
	}
	
	matched, _ := regexp.MatchString(`^[A-Z]+-[A-Z]+$`, marketID)
	if !matched {
		return ErrInvalidMarketID
	}
	
	return nil
}