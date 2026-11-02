package outbound

import (
	"context"
	"test-project/shared/money"
	"time"
)

//Это интерфейс, который использует OrderService, чтобы не зависеть от деталей SpotService

type MarketInfo struct {
    MarketID    string
    Name        string
    BaseAsset   string
    QuoteAsset  string
    Enabled     bool
    Price       money.Money    
    MinQuantity int64  
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type SpotInstrumentClient interface{
	GetMarket(ctx context.Context, marketID string) (*MarketInfo, error)
}