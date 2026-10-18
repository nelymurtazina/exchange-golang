package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "test-project/api/gen/order"
	"test-project/shared/authorization"
	"test-project/shared/interceptor"
	"test-project/orderService/config"
	handler "test-project/orderService/internal/adapters/inbound/grpcv1"
	grpcClient "test-project/orderService/internal/adapters/outbound/grpc_client"
	"test-project/orderService/internal/adapters/outbound/repository/postgres"
	"test-project/orderService/internal/core/service"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	repo := postgres.NewOrderRepository(db)

	spotConn, err := grpc.NewClient(cfg.Services.SpotServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to SpotService", zap.Error(err))
	}
	defer spotConn.Close()

	spotClient := grpcClient.NewSpotClient(spotConn)

	orderService := service.NewOrderService(repo, spotClient)
	grpcHandler := handler.NewOrderHandler(orderService)

	jwtValidator := authorization.NewTokenValidator(cfg.JWT.Secret)

	skipMethodsMap := make(map[string]struct{})

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.PanicRecoveryInterceptor(logger),
			interceptor.XRequestIDInterceptor(),
			interceptor.LoggerInterceptor(logger),
			interceptor.AuthInterceptor(jwtValidator, logger, skipMethodsMap),
		),
		grpc.ChainStreamInterceptor(
			interceptor.PanicRecoveryStreamInterceptor(logger),
			interceptor.XRequestIDStreamInterceptor(),
			interceptor.LoggerStreamInterceptor(logger),
			interceptor.AuthStreamInterceptor(jwtValidator, logger, skipMethodsMap),
		),
	)

	pb.RegisterOrderServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.Services.OrderServicePort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("OrderService started", zap.String("address", cfg.Services.OrderServicePort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", zap.Error(err))
			select {
			case stop <- syscall.SIGTERM:
			default:
			}
		}
	}()

	<-stop
	logger.Info("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdownDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Info("Server stopped gracefully")
	case <-ctx.Done():
		logger.Info("Shutdown timeout, forcing stop")
		grpcServer.Stop()
	}
}