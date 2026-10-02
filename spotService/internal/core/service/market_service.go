package service

import (
	"context"
	"fmt"
	"time"

	"test-project/shared/interceptor"
	"test-project/spotService/internal/core/domain"
	portsInbound "test-project/spotService/internal/core/ports/inbound"
	"test-project/spotService/internal/core/ports/outbound"
	ports "test-project/spotService/internal/core/ports/outbound"
)

type MarketService struct {
    repo ports.MarketRepository
    cache outbound.MarketCache
}

func NewMarketService(repo ports.MarketRepository, cache outbound.MarketCache) *MarketService {
    return &MarketService{
        repo: repo,
        cache: cache,
    }
}

func (s *MarketService) GetMarket(ctx context.Context, input portsInbound.GetMarketInput) (*portsInbound.GetMarketOutput, error) {
    if err := domain.ValidateMarketID(input.MarketID); err != nil {
        return nil, err
    }

    market, err := s.repo.GetByID(ctx, input.MarketID)
    if err != nil {
		return nil, err
	}
    if !market.Enabled {
		return nil, domain.ErrMarketDisabled
	}

    return &portsInbound.GetMarketOutput{
        Market: market,
    }, nil
}

func (s *MarketService) GetMarketByID(ctx context.Context, input portsInbound.GetMarketInput) (*portsInbound.GetMarketOutput, error) {
	role, ok := interceptor.GetUserRoleFromContext(ctx)
	if !ok || role == domain.RoleGuest {
		return nil, domain.ErrPermissionDenied
	}

    cachedMarket, err := s.cache.Get(ctx, input.MarketID)
    if err == nil && cachedMarket != nil {
		return &portsInbound.GetMarketOutput{
			Market: cachedMarket,
		}, nil
	}

	market, err := s.repo.GetByID(ctx, input.MarketID) 
	if err != nil {
		return nil, err
	}
    _ = s.cache.Set(ctx, market, 15*time.Minute)

	return &portsInbound.GetMarketOutput{
		Market: market,
	}, nil
}

const defaultPageSize = 20
const maxPageSize = 100

func (s *MarketService) ListMarkets(ctx context.Context, input portsInbound.ListMarketsInput) (*portsInbound.ListMarketsOutput, error) {
    pageSize := input.PageSize
    if pageSize == 0 {
        pageSize = defaultPageSize
    }
    if pageSize > maxPageSize {
        pageSize = maxPageSize
    }

    // Извлекаем PageToken. Если он пустой — база отдаст самые первые записи
	cursor := input.PageToken 

    markets, err := s.repo.GetAllActiveWithCursor(ctx, int(pageSize), cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list markets: %w", err)
	}

    var nextToken string
    if len(markets) == int(pageSize) {
        nextToken = markets[len(markets)-1].MarketID
    }

    return &portsInbound.ListMarketsOutput{
        Markets:       markets,
        NextPageToken: nextToken,
    }, nil
}
