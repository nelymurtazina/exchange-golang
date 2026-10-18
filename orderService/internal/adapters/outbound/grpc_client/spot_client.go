package grpc_client

import (
	"context"
	"fmt"
	"time"

	spotv1 "test-project/api/gen/spot"
	"test-project/orderService/internal/core/ports/outbound"

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

	return &outbound.MarketInfo{
		MarketID:    resp.Market.MarketId,
		Enabled:     resp.Market.Enabled,
		MinQuantity: minQuantity,
	}, nil
}