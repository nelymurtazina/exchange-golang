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
	"test-project/orderService/config"
	"test-project/orderService/internal/adapters/grpc_client"
	"test-project/orderService/internal/adapters/inbound/grpcv1"
	"test-project/orderService/internal/adapters/outbound/repository/postgres"
	"test-project/orderService/internal/core/service"
	"test-project/shared/interceptor"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		logger.Fatal("invalid config", zap.Error(err))
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	orderRepo := postgres.NewOrderRepository(db)

	conn, err := grpc.Dial(cfg.Services.SpotServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to SpotService", zap.Error(err))
	}
	defer func() { _ = conn.Close() }()

	spotClient := grpc_client.NewSpotClient(conn)


	orderService := service.NewOrderService(orderRepo, spotClient)

	grpcHandler := grpcv1.NewOrderHandler(orderService)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.PanicRecoveryInterceptor(logger),
			interceptor.XRequestIDInterceptor(),
			interceptor.LoggerInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			interceptor.PanicRecoveryStreamInterceptor(logger),
			interceptor.XRequestIDStreamInterceptor(),
			interceptor.LoggerStreamInterceptor(logger),
		),
	)

	pb.RegisterOrderServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.Services.InstrumentServicePort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Info("SpotService started", zap.String("address", cfg.Services.InstrumentServicePort))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
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