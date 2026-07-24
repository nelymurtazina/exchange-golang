package mapper

import (
	spotv1 "grpc-exchange/gen/spot"
	"grpc-exchange/services/spotInstrumentService/domain"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoMarket(market *domain.Market) *spotv1.Market {
	if market == nil {
		return nil
	}

	var deletedAt *timestamppb.Timestamp
	if market.DeletedAt != nil {
		deletedAt = timestamppb.New(*market.DeletedAt)
	}

	return &spotv1.Market{
		MarketId:    market.MarketID,
		Name:        market.Name,
		BaseAsset:   market.BaseAsset,
		QuoteAsset:  market.QuoteAsset,
		Enabled:     market.Enabled,
		CreatedAt:   timestamppb.New(market.CreatedAt),
		DeletedAt:   deletedAt,
	}
}

func ToProtoMarkets(markets []*domain.Market) []*spotv1.Market {
    if markets == nil {
        return nil
    }
    
    result := make([]*spotv1.Market, 0, len(markets))
    for _, m := range markets {
        result = append(result, ToProtoMarket(m))
    }
    return result
}