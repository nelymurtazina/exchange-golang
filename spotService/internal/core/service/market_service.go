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

func (s *MarketService) GetMarket(ctx context.Context, input ports.GetMarketInput) (*ports.GetMarketOutput, error) {
    if err := domain.ValidateMarketID(input.MarketID); err != nil {
        return nil, err
    }

    market, err := s.repo.GetByID(ctx, input.MarketID)
    if err != nil {
        if errors.Is(err, domain.ErrMarketNotFound) {
            return nil, domain.ErrMarketNotFound
        }
        return nil, fmt.Errorf("failed to get market: %w", err)
    }
    if market == nil {
        return nil, domain.ErrMarketNotFound
    }

    return &ports.GetMarketOutput{
        Market: market,
    }, nil
}

func (s *MarketService) ListMarkets(ctx context.Context, input ports.ListMarketsInput) (*ports.ListMarketsOutput, error) {
    markets, err := s.repo.GetAllActive(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list markets: %w", err)
    }

    return &ports.ListMarketsOutput{
        Markets:       markets,
        NextPageToken: "", // ← ПОКА ПРОСТО
    }, nil
}