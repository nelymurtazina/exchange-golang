package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	commonv1 "grpc-exchange/gen/common"
// 	orderv1 "grpc-exchange/gen/order"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// func main() {
// 	conn, err := grpc.NewClient(
// 		"localhost:50052",
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 	)
// 	if err != nil {
// 		log.Fatalf("failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	client := orderv1.NewOrderServiceClient(conn)
// 	ctx := context.Background()

// 	fmt.Println(" Создание заказа ")

// 	createResp, err := client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
// 		Key:      "test-" + time.Now().Format("150405"),
// 		UserId:   "user123",
// 		MarketId: "btc-usd",
// 		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
// 		Price: &commonv1.Money{
// 			Amount: &commonv1.Decimal{
// 				Units: 50000,
// 				Nanos: 0,
// 			},
// 			CurrencyCode: "USD",
// 		},
// 		Quantity: &commonv1.Decimal{
// 			Units: 1,
// 			Nanos: 500000000,
// 		},
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to create order: %v", err)
// 	}

// 	fmt.Printf("Заказ создан!\n")
// 	fmt.Printf("   ID: %s\n", createResp.OrderId)
// 	fmt.Printf("   Статус: %v\n", createResp.Status)

// 	fmt.Println("\n=== Проверка статуса ===")

// 	statusResp, err := client.GetOrderStatus(ctx, &orderv1.GetOrderStatusRequest{
// 		OrderId: createResp.OrderId,
// 		UserId:  "user123",
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to get order status: %v", err)
// 	}

// 	order := statusResp.Order
// 	fmt.Printf("Информация о заказе:\n")
// 	fmt.Printf("   ID: %s\n", order.OrderId)
// 	fmt.Printf("   Статус: %v\n", order.Status)
// 	fmt.Printf("   Цена: %d %s\n",
// 		order.Price.Amount.Units,
// 		order.Price.CurrencyCode)
// 	fmt.Printf("   Количество: %d.%09d\n",
// 		order.Quantity.Units,
// 		order.Quantity.Nanos)
// }