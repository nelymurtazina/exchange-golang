package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "test-project/api/gen/spot"
	"test-project/shared/interceptor"
	"test-project/spotService/config"
	handler "test-project/spotService/internal/adapters/inbound/grpc"
	"test-project/spotService/internal/adapters/outbound/auth"
	"test-project/spotService/internal/adapters/outbound/repository/postgres"
	"test-project/spotService/internal/core/service"

	redisAdapter "test-project/spotService/internal/adapters/outbound/redis"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("failed to create logger: %v", err)
    }
    defer func(){
        _ = logger.Sync()
    }()
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
    defer func() { _ = db.Close() }()

    if cfg.Migration.Enabled {
        logger.Info("Running migrations...")

        dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
            cfg.Database.User,
            cfg.Database.Password,
            cfg.Database.Host,
            cfg.Database.Port,
            cfg.Database.DBName,
            cfg.Database.SSLMode,
        )

        m, err := migrate.New(
            "file://"+cfg.Migration.Path,
            dsn,
        )
        if err != nil {
            logger.Fatal("failed to create migrate instance", zap.Error(err))
        }

        if err := m.Up(); err != nil && err != migrate.ErrNoChange {
            logger.Fatal("failed to apply migrations", zap.Error(err))
        }

        logger.Info("Migrations applied successfully")
    }

    repo := postgres.NewMarketRepository(db)
    rdb := redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
    })
    marketCache := redisAdapter.NewMarketCache(rdb)
    jwtValidator := auth.NewTokenValidator(cfg.JWT.Secret)
    marketService := service.NewMarketService(repo, marketCache)  
    skipMethods := []string{
		"/spot.SpotInstrumentService/GetAllActive",  
	}
    grpcHandler := handler.NewMarketHandler(marketService)  

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

    lis, err := net.Listen("tcp", cfg.Services.InstrumentServicePort)
    if err != nil {
        logger.Fatal("failed to listen", zap.Error(err))
    }

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        logger.Info("SpotService started", zap.String("address", cfg.Services.InstrumentServicePort)) 
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