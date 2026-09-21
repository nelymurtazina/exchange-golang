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

    return &ports.GetMarketOutput{
        Market: market,
    }, nil
}

const defaultPageSize = 20
const maxPageSize = 100

func (s *MarketService) ListMarkets(ctx context.Context, input ports.ListMarketsInput) (*ports.ListMarketsOutput, error) {
    pageSize := input.PageSize
    if pageSize == 0 {
        pageSize = defaultPageSize
    }
    if pageSize > maxPageSize {
        pageSize = maxPageSize
    }

    markets, err := s.repo.GetAllActive(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list markets: %w", err)
    }

    var nextToken string
    if len(markets) == int(pageSize) {
        nextToken = markets[len(markets)-1].MarketID
    }

    return &ports.ListMarketsOutput{
        Markets:       markets,
        NextPageToken: nextToken,
    }, nil
}