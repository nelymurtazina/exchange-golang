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
    UserRole string
}

type DisableMarketInput struct {
    MarketID string
    UserRole string
}

type GetMarketOutput struct {
    Market *domain.Market 
}

type DeleteMarketInput struct {
	MarketID string 
	UserRole string 
}

type MarketService interface {
	ListMarkets(ctx context.Context, input ListMarketsInput) (*ListMarketsOutput, error)
	GetMarket(ctx context.Context, input GetMarketInput) (*GetMarketOutput, error)
	DisableMarket(ctx context.Context, input DisableMarketInput) (*GetMarketOutput, error)

    DeleteMarket(ctx context.Context, input DeleteMarketInput) error
}