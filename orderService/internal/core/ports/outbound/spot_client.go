package outbound

import "context"

//Это интерфейс, который использует OrderService, чтобы не зависеть от деталей SpotService

type MarketInfo struct {
    MarketID string
    Enabled  bool
}

type SpotInstrumentClient interface{
	GetMarket(ctx context.Context, marketID string) (*MarketInfo, error)
}