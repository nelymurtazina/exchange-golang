package outbound

import (
	"context"
	"test-project/orderService/internal/core/domain"
	"time"
)

//Это интерфейс, который использует OrderService, чтобы не зависеть от деталей SpotService

type MarketInfo struct {
    MarketID    string
    Name        string
    BaseAsset   string
    QuoteAsset  string
    Enabled     bool
    Price       domain.Money    
    MinQuantity int64  
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type SpotInstrumentClient interface{
	GetMarket(ctx context.Context, marketID string) (*MarketInfo, error)
}