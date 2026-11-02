package service

import (
	"context"
	"fmt"
	"time"

	"test-project/spotService/internal/core/domain"
	portsInbound "test-project/spotService/internal/core/ports/inbound"
	ports "test-project/spotService/internal/core/ports/outbound"
)

type MarketService struct {
	repo  ports.MarketRepository
	cache ports.MarketCache
}

func NewMarketService(repo ports.MarketRepository, cache ports.MarketCache) *MarketService {
	return &MarketService{
		repo:  repo,
		cache: cache,
	}
}

func denyGuest(role string) error {
	if role == "" || role == domain.RoleGuest {
		return domain.ErrPermissionDenied
	}
	return nil
}

func requireAdmin(role string) error {
	if role != domain.RoleAdmin {
		return domain.ErrPermissionDenied
	}
	return nil
}


func (s *MarketService) GetMarket(ctx context.Context, input portsInbound.GetMarketInput) (*portsInbound.GetMarketOutput, error) {
	if err := denyGuest(input.UserRole); err != nil {
		return nil, err
	}
	if err := domain.ValidateMarketID(input.MarketID); err != nil {
		return nil, err
	}

	cachedMarket, err := s.cache.Get(ctx, input.MarketID)
	if err == nil && cachedMarket != nil {
		if !cachedMarket.Enabled {
			return nil, domain.ErrMarketDisabled
		}
		return &portsInbound.GetMarketOutput{Market: cachedMarket}, nil
	}

	market, err := s.repo.GetByID(ctx, input.MarketID)
	if err != nil {
		return nil, err
	}
	if market.IsDeleted() {
		return nil, domain.ErrMarketNotFound
	}
	if !market.Enabled {
		return nil, domain.ErrMarketDisabled
	}

	_ = s.cache.Set(ctx, market, 15*time.Minute)

	return &portsInbound.GetMarketOutput{Market: market}, nil
}

func (s *MarketService) DisableMarket(ctx context.Context, input portsInbound.DisableMarketInput) (*portsInbound.GetMarketOutput, error) {
	if err := requireAdmin(input.UserRole); err != nil {
		return nil, err
	}
	if err := denyGuest(input.UserRole); err != nil {
		return nil, err
	}
	if err := domain.ValidateMarketID(input.MarketID); err != nil {
		return nil, err
	}

	market, err := s.repo.GetByID(ctx, input.MarketID)
	if err != nil {
		return nil, err
	}
	if market.IsDeleted() {
		return nil, domain.ErrMarketNotFound
	}
	market.SoftDelete()
	if err := s.repo.Update(ctx, market); err != nil {
		return nil, fmt.Errorf("failed to disable market: %w", err)
	}

	market.Enabled = false
	_ = s.cache.Delete(ctx, market.MarketID)
	_ = s.cache.DeleteMarketsList(ctx)

	return &portsInbound.GetMarketOutput{Market: market}, nil
}

const defaultPageSize = 20
const maxPageSize = 100

func (s *MarketService) ListMarkets(ctx context.Context, input portsInbound.ListMarketsInput) (*portsInbound.ListMarketsOutput, error) {
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	cachedMarkets, err := s.cache.GetMarketsList(ctx)
	if err == nil && cachedMarkets != nil {
		var paginatedMarkets []*domain.Market
		start := 0
		end := len(cachedMarkets)

		if input.PageToken != "" {
			for i, m := range cachedMarkets {
				if m.MarketID == input.PageToken {
					start = i + 1
					break
				}
			}
		}

		if start >= len(cachedMarkets) {
			paginatedMarkets = []*domain.Market{}
		} else {
			end = start + int(pageSize)
			if end > len(cachedMarkets) {
				end = len(cachedMarkets)
			}
			paginatedMarkets = cachedMarkets[start:end]
		}

		var nextToken string
		if len(paginatedMarkets) == int(pageSize) && end < len(cachedMarkets) {
			nextToken = paginatedMarkets[len(paginatedMarkets)-1].MarketID
		}

		return &portsInbound.ListMarketsOutput{
			Markets:       paginatedMarkets,
			NextPageToken: nextToken,
		}, nil
	}

	markets, err := s.repo.GetAllActiveWithCursor(ctx, int(pageSize)+1, input.PageToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list markets: %w", err)
	}

	var nextToken string
	if len(markets) > int(pageSize) {
		nextToken = markets[pageSize-1].MarketID
		markets = markets[:pageSize]
	}

	return &portsInbound.ListMarketsOutput{
		Markets:       markets,
		NextPageToken: nextToken,
	}, nil
}


func (s *MarketService) DeleteMarket(ctx context.Context, input portsInbound.DeleteMarketInput) error {
	if err := requireAdmin(input.UserRole); err != nil {
		return err
	}
	if err := denyGuest(input.UserRole); err != nil {
		return err
	}
	if err := domain.ValidateMarketID(input.MarketID); err != nil {
		return err
	}

	isDeleted, err := s.repo.IsDeleted(ctx, input.MarketID)
	if err != nil {
		return err
	}
	if isDeleted {
		return domain.ErrMarketNotFound 
	}
	
	if err := s.repo.SoftDelete(ctx, input.MarketID); err != nil {
		return fmt.Errorf("failed to soft delete market: %w", err)
	}

	_ = s.cache.Delete(ctx, input.MarketID)
	_ = s.cache.DeleteMarketsList(ctx)

	return nil
}
