package grpc

import (
	"context"
	"errors"
	"test-project/spotService/internal/core/domain"
	"test-project/spotService/internal/core/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	 pb "test-project/api/gen/spot"
)

type MarketHandler struct {
	pb.UnimplementedSpotInstrumentServiceServer
	service ports.MarketService
}

func NewMarketHandler(service ports.MarketService) *MarketHandler {
	return &MarketHandler{service: service}
}

func (h *MarketHandler) ListMarkets(ctx context.Context, req *pb.ListMarketsRequest) (*pb.ListMarketsResponse, error) {
    input := ports.ListMarketsInput{
        PageSize:  req.PageSize,
        PageToken: req.PageToken,
        UserRoles: req.UserRoles,
    }

    output, err := h.service.ListMarkets(ctx, input)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrMarketNotFound):
            return nil, status.Error(codes.NotFound, err.Error())
        default:
            return nil, status.Error(codes.Internal, "internal error")
        }
    }

    var protoMarkets []*pb.Market
    for _, m := range output.Markets {
        protoMarkets = append(protoMarkets, toProtoMarket(m))
    }

    return &pb.ListMarketsResponse{
        Markets:       protoMarkets,
        NextPageToken: output.NextPageToken,
    }, nil
}

func (h *MarketHandler) GetMarket(ctx context.Context, req *pb.GetMarketRequest) (*pb.GetMarketResponse, error) {
    input := ports.GetMarketInput{
        MarketID: req.MarketId,
    }

    output, err := h.service.GetMarket(ctx, input)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrInvalidMarketID):
            return nil, status.Error(codes.InvalidArgument, err.Error())
        case errors.Is(err, domain.ErrMarketNotFound):
            return nil, status.Error(codes.NotFound, err.Error())
        case errors.Is(err, domain.ErrMarketDisabled):
            return nil, status.Error(codes.NotFound, "market not found or inactive")
        case errors.Is(err, domain.ErrMarketDeleted):
            return nil, status.Error(codes.NotFound, "market not found")
        default:
            return nil, status.Error(codes.Internal, "internal error")
        }
    }

    return &pb.GetMarketResponse{
        Market: toProtoMarket(output.Market),
    }, nil
}

func toProtoMarket(m *domain.Market) *pb.Market {
    if m == nil {
        return nil
    }
    
    var deletedAt *timestamppb.Timestamp
    if m.DeletedAt != nil {
        deletedAt = timestamppb.New(*m.DeletedAt)
    }
    
    return &pb.Market{
        MarketId:   m.MarketID,
        Name:       m.Name,
        BaseAsset:  m.BaseAsset,
        QuoteAsset: m.QuoteAsset,
        Enabled:    m.Enabled,
        CreatedAt:  timestamppb.New(m.CreatedAt),
        UpdatedAt:  timestamppb.New(m.UpdatedAt),
        DeletedAt:  deletedAt,
        Price:      &m.Price,
    }
}