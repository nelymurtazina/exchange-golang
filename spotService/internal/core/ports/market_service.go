package ports

import (
	"context"
	"test-project/spotService/internal/core/domain"
)

type ListMarketsInput struct {
    PageSize  int32
    PageToken string
    UserRole string
}

type ListMarketsOutput struct {
    Markets       []*domain.Market
    NextPageToken string
}

type GetMarketInput struct {
    MarketID string
}

type GetMarketOutput struct {
    Market *domain.Market 
}

type MarketService interface{
	ListMarkets(ctx context.Context, input ListMarketsInput) (*ListMarketsOutput, error)
	GetMarket(ctx context.Context, input GetMarketInput) (*GetMarketOutput, error)
}