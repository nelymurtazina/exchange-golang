package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "test-project/api/gen/order"
	"test-project/orderService/config"
	handler "test-project/orderService/internal/adapters/inbound/grpcv1"
	grpcClient "test-project/orderService/internal/adapters/outbound/grpc_client"
	"test-project/orderService/internal/adapters/outbound/repository/postgres"
	"test-project/orderService/internal/core/service"
	"test-project/shared/authorization"
	"test-project/shared/interceptor"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	db         *pgxpool.Pool
	spotConn   *grpc.ClientConn
	grpcServer *grpc.Server
	lis        net.Listener
}

func New(cfg *config.Config) (*App, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	spotConn, err := grpc.NewClient(cfg.Services.SpotServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to spot service: %w", err)
	}

	repo := postgres.NewOrderRepository(db)
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
		spotConn.Close()
		db.Close()
		return nil, fmt.Errorf("listen: %w", err)
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		spotConn:   spotConn,
		grpcServer: grpcServer,
		lis:        lis,
	}, nil
}

func (a *App) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		a.logger.Info("OrderService started", zap.String("address", a.cfg.Services.OrderServicePort))
		if err := a.grpcServer.Serve(a.lis); err != nil {
			a.logger.Error("failed to serve", zap.Error(err))
			select {
			case stop <- syscall.SIGTERM:
			default:
			}
		}
	}()

	<-stop
	a.logger.Info("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdownDone := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		a.logger.Info("Server stopped gracefully")
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout, forcing stop")
		a.grpcServer.Stop()
	}

	return nil
}

func (a *App) Close() {
	if a.spotConn != nil {
		_ = a.spotConn.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
	_ = a.logger.Sync()
}
