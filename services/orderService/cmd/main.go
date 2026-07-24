package main

import (
	"context"
	"log"
	"net"
	"net/http"

	orderv1 "grpc-exchange/gen/order"
	spotv1 "grpc-exchange/gen/spot"
	userv1 "grpc-exchange/gen/user"
	"grpc-exchange/services/orderService/repository"
	"grpc-exchange/services/orderService/service"
	"grpc-exchange/shered/cache"
	"grpc-exchange/shered/config"
	"grpc-exchange/shered/database"
	"grpc-exchange/shered/interceptor"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)
func main() {
    // Загружаем конфиг
    cfg := config.LoadConfig()

    // Логгер
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("failed to create logger: %v", err)
    }
    defer logger.Sync()

    // Подключаемся к БД
    db, err := database.NewPostgresConnection(database.Config(cfg.Database))
    if err != nil {
        logger.Fatal("failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Подключаемся к Redis
    redisCache := cache.NewRedisCache(cfg.Redis)
    if err := redisCache.Ping(context.Background()); err != nil {
        logger.Warn("Redis not available", zap.Error(err))
    } else {
        logger.Info("Redis connected")
    }

    // Подключаемся к SpotInstrumentService
    spotConn, err := grpc.NewClient(
        "localhost"+cfg.Services.InstrumentServicePort,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        logger.Fatal("failed to connect to spot service", zap.Error(err))
    }
    defer spotConn.Close()
    spotClient := spotv1.NewSpotInstrumentServiceClient(spotConn)

    // Подключаемся к UserService (для авторизации)
    userConn, err := grpc.NewClient(
        "localhost"+cfg.Services.UserServicePort,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        logger.Fatal("failed to connect to user service", zap.Error(err))
    }
    defer userConn.Close()
    userClient := userv1.NewUserServiceClient(userConn)

    // Репозиторий и сервис
    orderRepo := repository.NewPostgresOrderRepository(db)
    orderSrv := service.NewOrderService(orderRepo, spotClient)

    // Метрики
    metrics := interceptor.NewMetricsCollector()

    // gRPC сервер с интерсепторами (ВКЛЮЧАЕМ АВТОРИЗАЦИЮ!)
    grpcServer := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            interceptor.XRequestIDInterceptor,
            interceptor.AuthInterceptor(userClient), 
            interceptor.LoggerInterceptorWithRedis(logger, redisCache),
            interceptor.PanicRecoveryInterceptor,
            metrics.UnaryServerInterceptor(),
        ),
    )

    // 1Регистрация
    orderv1.RegisterOrderServiceServer(grpcServer, orderSrv)
    reflection.Register(grpcServer)

    // HTTP сервер для Prometheus
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        logger.Info("Metrics server started", zap.String("address", cfg.Metrics.OrderPort))
        http.ListenAndServe(cfg.Metrics.OrderPort, nil)
    }()

    // Запуск gRPC
    lis, err := net.Listen("tcp", cfg.Services.OrderServicePort)
    if err != nil {
        logger.Fatal("failed to listen", zap.Error(err))
    }

    logger.Info("OrderService started", zap.String("address", cfg.Services.OrderServicePort))
    if err := grpcServer.Serve(lis); err != nil {
        logger.Fatal("failed to serve", zap.Error(err))
    }
}