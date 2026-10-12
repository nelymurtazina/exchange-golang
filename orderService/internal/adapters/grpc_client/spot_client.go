package grpc_client

import (
	"context"
	"fmt"
	spotv1 "test-project/api/gen/spot" 
	
	"test-project/orderService/internal/core/ports/outbound"
	
	"google.golang.org/grpc"
)

type SpotClient struct {
	client spotv1.SpotInstrumentServiceClient 
}
 
// Мы передаем сюда уже готовое gRPC-соединение (grpc.ClientConn) 
func NewSpotClient(conn *grpc.ClientConn) outbound.SpotInstrumentClient{
	return &SpotClient{
		client: spotv1.NewSpotInstrumentServiceClient(conn),
	}
}

// GetMarket реализует интерфейс 
func (c *SpotClient) GetMarket(ctx context.Context, marketID string) (*outbound.MarketInfo, error) {
	//  делаем реальный gRPC-вызов к SpotService
	resp, err := c.client.GetMarket(ctx, &spotv1.GetMarketRequest{
		MarketId: marketID,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC call to SpotService failed: %w", err)
	}

	// из proto-структуры в MarketInfo
	return &outbound.MarketInfo{
		MarketID: resp.Market.MarketId,
		Enabled:  resp.Market.Enabled,
	}, nil
}