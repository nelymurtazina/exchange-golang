package service

import (
	"context"

	spotv1 "grpc-exchange/gen/spot"
	"grpc-exchange/services/spotInstrumentService/mapper"
	"grpc-exchange/services/spotInstrumentService/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InstrumentService struct {
	spotv1.UnimplementedSpotInstrumentServiceServer
	repo repository.MarketRepository
}

func NewInstrumentService(repo repository.MarketRepository) *InstrumentService {
	return &InstrumentService{repo: repo}
}

func (s *InstrumentService) ListMarkets(
	ctx context.Context,
	req *spotv1.ListMarketsRequest,
) (*spotv1.ListMarketsResponse, error) {
	if req.MarketId != "" {
		market, err := s.repo.GetActiveByID(ctx, req.MarketId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get market: %v", err)
		}
		if market == nil {
			return nil, status.Errorf(codes.NotFound, "market not found or inactive")
		}
		return &spotv1.ListMarketsResponse{
			Markets: []*spotv1.Market{mapper.ToProtoMarket(market)},
		}, nil
	}

	markets, err := s.repo.GetAllActive(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get markets: %v", err)
	}

	return &spotv1.ListMarketsResponse{
		Markets: mapper.ToProtoMarkets(markets),
	}, nil
}

func (s *InstrumentService) GetMarket(
	ctx context.Context,
	req *spotv1.GetMarketRequest,
) (*spotv1.GetMarketResponse, error) {
	if req.MarketId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "market_id is required")
	}

	market, err := s.repo.GetActiveByID(ctx, req.MarketId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get market: %v", err)
	}
	if market == nil {
		return nil, status.Errorf(codes.NotFound, "market not found")
	}

	return &spotv1.GetMarketResponse{
		Market: mapper.ToProtoMarket(market),
	}, nil
}
