package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "test-project/api/gen/spot"
	"test-project/shared/authorization"
	"test-project/shared/interceptor"
	"test-project/spotService/config"
	handler "test-project/spotService/internal/adapters/inbound/grpc"
	redisAdapter "test-project/spotService/internal/adapters/outbound/redis"
	"test-project/spotService/internal/adapters/outbound/repository/postgres"
	"test-project/spotService/internal/core/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	db         *pgxpool.Pool
	redis      *redis.Client
	grpcServer *grpc.Server
	lis        net.Listener
}

func New(cfg *config.Config) (*App, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Database
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelPing()

	if err := rdb.Ping(ctxPing).Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	// Dependencies
	repo := postgres.NewMarketRepository(db)
	marketCache := redisAdapter.NewMarketCache(rdb)
	jwtValidator := authorization.NewTokenValidator(cfg.JWT.Secret)
	marketService := service.NewMarketService(repo, marketCache)
	grpcHandler := handler.NewMarketHandler(marketService)

	skipMethods := map[string]struct{}{
		"/spot.v1.SpotInstrumentService/ListMarkets": {},
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.PanicRecoveryInterceptor(logger),
			interceptor.XRequestIDInterceptor(),
			interceptor.LoggerInterceptor(logger),
			interceptor.AuthInterceptor(jwtValidator, logger, skipMethods),
		),
		grpc.ChainStreamInterceptor(
			interceptor.PanicRecoveryStreamInterceptor(logger),
			interceptor.XRequestIDStreamInterceptor(),
			interceptor.LoggerStreamInterceptor(logger),
			interceptor.AuthStreamInterceptor(jwtValidator, logger, skipMethods),
		),
	)

	pb.RegisterSpotInstrumentServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	// Listener
	lis, err := net.Listen("tcp", cfg.Services.InstrumentServicePort)
	if err != nil {
		db.Close()
		_ = rdb.Close()
		return nil, fmt.Errorf("listen: %w", err)
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		redis:      rdb,
		grpcServer: grpcServer,
		lis:        lis,
	}, nil
}

func (a *App) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		a.logger.Info("SpotService started",
			zap.String("address", a.cfg.Services.InstrumentServicePort),
		)

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

	done := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("Server stopped gracefully")
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout, forcing stop")
		a.grpcServer.Stop()
	}

	return nil
}

func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
	_ = a.logger.Sync()
}