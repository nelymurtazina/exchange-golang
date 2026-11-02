package grpc_client

import (
	"context"
	"fmt"
	"time"

	spotv1 "test-project/api/gen/spot"
	"test-project/orderService/internal/core/ports/outbound"
	"test-project/shared/money"

	"google.golang.org/grpc"
)

type SpotClient struct {
	client spotv1.SpotInstrumentServiceClient
}

func NewSpotClient(conn *grpc.ClientConn) outbound.SpotInstrumentClient {
	return &SpotClient{
		client: spotv1.NewSpotInstrumentServiceClient(conn),
	}
}

func (c *SpotClient) GetMarket(ctx context.Context, marketID string) (*outbound.MarketInfo, error) {
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.GetMarket(callCtx, &spotv1.GetMarketRequest{
		MarketId: marketID,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC call to SpotService failed: %w", err)
	}

	var minQuantity int64
	if resp.Market.MinQuantity != nil {
		minQuantity = resp.Market.MinQuantity.Units
	}

	var priceMoney money.Money
	if resp.Market.Price != nil && resp.Market.Price.Amount != nil {
		priceMoney = money.Money{
			Units:        resp.Market.Price.Amount.Units,
			Nanos:        resp.Market.Price.Amount.Nanos,
			CurrencyCode: resp.Market.Price.CurrencyCode,
		}
	}
	return &outbound.MarketInfo{
		MarketID:    resp.Market.MarketId,
		Name:        resp.Market.Name,
		BaseAsset:   resp.Market.BaseAsset,
		QuoteAsset:  resp.Market.QuoteAsset,
		Enabled:     resp.Market.Enabled,
		Price:       priceMoney,
		MinQuantity: minQuantity,
		CreatedAt:   resp.Market.CreatedAt.AsTime(),
		UpdatedAt:   resp.Market.UpdatedAt.AsTime(),
	}, nil
}