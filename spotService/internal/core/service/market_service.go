package service

import (
	"context"
	"errors"
	"fmt"
	"test-project/spotService/internal/core/domain"
	"test-project/spotService/internal/core/ports"
)

type MarketService struct {
	repo ports.MarketRepository
}

func NewMarketService(repo ports.MarketRepository) ports.MarketService {
	return &MarketService{
		repo: repo,
	}
}

func (m *MarketService) GetMarket(ctx context.Context, input ports.GetMarketInput) (*ports.GetMarketOutput, error) {
	if err := domain.ValidateMarketID(input.MarketID); err != nil{
		return nil, err
	}
	
	market, err := m.repo.GetByID(ctx, input.MarketID)
	if err != nil {
		if errors.Is(err, domain.ErrMarketNotFound){
			return nil, domain.ErrMarketNotFound
		}
		return nil, fmt.Errorf("error", err)
	}

	if market == nil {
        return nil, domain.ErrMarketNotFound
    }

	if market.DeletedAt != nil && !market.DeletedAt.IsZero() {
        return nil, domain.ErrMarketNotFound
    }
	return market, nil
}

func (m *MarketService) ListMarkets(ctx context.Context, input ports.ListMarketsInput) (*ports.ListMarketsOutput, error) {
	if err := domain.ValidateMarketID(input.MarketID); err != nil{
		return nil, err
	}
	
	market, err := m.repo.GetByID(ctx, input.MarketID)
	if err != nil {
		if errors.Is(err, domain.ErrMarketNotFound){
			return nil, domain.ErrMarketNotFound
		}
		return nil, fmt.Errorf("error", err)
	}

	if market == nil {
        return nil, domain.ErrMarketNotFound
    }

	if market.DeletedAt != nil && !market.DeletedAt.IsZero() {
        return nil, domain.ErrMarketNotFound
    }
	return market, nil
}

