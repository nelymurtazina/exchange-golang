package grpc

import (
	"context"
	"errors"
	"test-project/shared/interceptor"
	"test-project/spotService/internal/core/domain"
	"test-project/spotService/internal/core/ports"

	commonv1 "test-project/api/gen/common"
	pb "test-project/api/gen/spot"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MarketHandler struct {
	pb.UnimplementedSpotInstrumentServiceServer
	service ports.MarketService
}

func NewMarketHandler(service ports.MarketService) *MarketHandler {
	return &MarketHandler{service: service}
}

func (h *MarketHandler) ListMarkets(ctx context.Context, req *pb.ListMarketsRequest) (*pb.ListMarketsResponse, error) {
    role, ok := interceptor.GetUserRoleFromContext(ctx)
    if !ok{
        role = domain.RoleGuest
    }
    input := ports.ListMarketsInput{
        PageSize:  req.PageSize,
        PageToken: req.PageToken,
        UserRole: role,
    }

    output, err := h.service.ListMarkets(ctx, input)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrMarketAlreadyExists):
            return nil, status.Error(codes.AlreadyExists, "market already exists")
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
    
    return &pb.Market{
        MarketId:   m.MarketID,
        Name:       m.Name,
        BaseAsset:  m.BaseAsset,
        QuoteAsset: m.QuoteAsset,
        Enabled:    m.Enabled,
        Price:      toProtoMoney(m.Price),
        CreatedAt:  timestamppb.New(m.CreatedAt),
        UpdatedAt:  timestamppb.New(m.UpdatedAt),
    }
}

// domain.Money -> pb.Money ОНО? так мне конвертировать в money?
func toProtoMoney(m domain.Money) *commonv1.Money {
    return &commonv1.Money{
        Amount: &commonv1.Decimal{
            Units: m.Units,
            Nanos: m.Nanos,
        },
        CurrencyCode: m.CurrencyCode,
    }
}

// pb.Money -> domain.Money  ОНО?
func fromProtoMoney(m *commonv1.Money) domain.Money {
    if m == nil || m.Amount == nil {
        return domain.Money{}
    }
    return domain.Money{
        Units:        m.Amount.Units,
        Nanos:        m.Amount.Nanos,
        CurrencyCode: m.CurrencyCode,
    }
}