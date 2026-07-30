package main

// import (
//     "context"
//     "fmt"
//     "log"
//     "time"

//     commonv1 "grpc-exchange/gen/common"
//     orderv1 "grpc-exchange/gen/order"
//     userv1 "grpc-exchange/gen/user"

//     "google.golang.org/grpc"
//     "google.golang.org/grpc/credentials/insecure"
//     "google.golang.org/grpc/metadata"
// )

// func main() {
//     // Подключаемся к UserService
//     userConn, err := grpc.NewClient(
//         "localhost:50053",
//         grpc.WithTransportCredentials(insecure.NewCredentials()),
//     )
//     if err != nil {
//         log.Fatalf("failed to connect to user service: %v", err)
//     }
//     defer userConn.Close()

//     // Подключаемся к OrderService
//     orderConn, err := grpc.NewClient(
//         "localhost:50052",
//         grpc.WithTransportCredentials(insecure.NewCredentials()),
//     )
//     if err != nil {
//         log.Fatalf("failed to connect to order service: %v", err)
//     }
//     defer orderConn.Close()

//     userClient := userv1.NewUserServiceClient(userConn)
//     orderClient := orderv1.NewOrderServiceClient(orderConn)

//     ctx := context.Background()

//     testEmail := "testuser1@example.com"

//     fmt.Println("\n---Регистрация ---")
//     registerResp, err := userClient.Register(ctx, &userv1.RegisterRequest{
//         Username: "test_user1",
//         Email:    testEmail,
//         Password: "password1231",
//     })
//     if err != nil {
//         log.Printf("Register error (user may already exist): %v", err)
//     } else {
//         fmt.Printf("\n Регистрация успешна! ID: %s\n", registerResp.UserId)
//     }

//     fmt.Println("\n--- Вход ---")
//     loginResp, err := userClient.Login(ctx, &userv1.LoginRequest{
//         Email:    testEmail,
//         Password: "password1231",
//     })
//     if err != nil {
//         log.Fatalf("Login error: %v", err)
//     }
//     fmt.Printf("Вход успешен!\n")
//     fmt.Printf("   User ID: %s\n", loginResp.UserId)
//     fmt.Printf("   Token: %s...\n", loginResp.Token[:20])

//     fmt.Println("\n--- Создание заказа (с авторизацией) ---")

//     md := metadata.New(map[string]string{
//         "authorization": "Bearer " + loginResp.Token,
//     })
//     ctxWithAuth := metadata.NewOutgoingContext(ctx, md)

//     createResp, err := orderClient.CreateOrder(ctxWithAuth, &orderv1.CreateOrderRequest{
//         Key:      "test-" + time.Now().Format("150405"),
//         UserId:   loginResp.UserId,
//         MarketId: "btc-usd",
//         Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
//         Price: &commonv1.Money{
//             Amount: &commonv1.Decimal{
//                 Units: 50000,
//                 Nanos: 0,
//             },
//             CurrencyCode: "USD",
//         },
//         Quantity: &commonv1.Decimal{
//             Units: 1,
//             Nanos: 500000000,
//         },
//     })
//     if err != nil {
//         log.Fatalf("CreateOrder error: %v", err)
//     }
//     fmt.Printf("Заказ создан!\n")
//     fmt.Printf("   ID: %s\n", createResp.OrderId)
//     fmt.Printf("   Статус: %v\n", createResp.Status)

//     fmt.Println("\n---  Проверка статуса ---")
//     statusResp, err := orderClient.GetOrderStatus(ctxWithAuth, &orderv1.GetOrderStatusRequest{
//         OrderId: createResp.OrderId,
//     })
//     if err != nil {
//         log.Fatalf("GetOrderStatus error: %v", err)
//     }
//     fmt.Printf("Статус заказа: %v\n", statusResp.Order.Status)

//     fmt.Println("\n--- Выход ---")
//     logoutResp, err := userClient.Logout(ctx, &userv1.LogoutRequest{
//         UserId: loginResp.UserId,
//     })
//     if err != nil {
//         log.Printf("Logout error: %v", err)
//     } else {
//         fmt.Printf("Выход успешен: %s\n", logoutResp.Message)
//     }
// }