package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "test-project/api/gen/user"
	"test-project/shared/authorization"
	"test-project/shared/interceptor"
	"test-project/userService/config"
	handler "test-project/userService/internal/adapters/inbound/grpc"
	"test-project/userService/internal/adapters/outbound/auth"
	"test-project/userService/internal/adapters/outbound/redis"
	"test-project/userService/internal/adapters/outbound/repository/postgres"
	"test-project/userService/internal/core/service"

	"github.com/jackc/pgx/v5/pgxpool"
	redisLib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	db         *pgxpool.Pool
	redis      *redisLib.Client
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

	redisClient := redisLib.NewClient(&redisLib.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	repo := postgres.NewUserRepository(db)
	sessionStore := redis.NewSessionStore(redisClient)

	jwtValidator := authorization.NewTokenValidator(cfg.JWT.Secret)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	passwordManager := auth.NewPasswordManager(cfg.Password.BcryptCost)

	userService := service.NewUserService(repo, jwtManager, passwordManager, sessionStore)
	grpcHandler := handler.NewUserHandler(userService)

	skipMethods := map[string]struct{}{
		"/user.v1.UserService/Register": {},
		"/user.v1.UserService/Login":    {},
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

	pb.RegisterUserServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	// Listener
	lis, err := net.Listen("tcp", cfg.Services.UserServicePort)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("listen: %w", err)
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		redis:      redisClient,
		grpcServer: grpcServer,
		lis:        lis,
	}, nil
}

// Run запускает сервер и ждёт сигнала завершения
func (a *App) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		a.logger.Info("UserService started",
			zap.String("address", a.cfg.Services.UserServicePort),
			zap.String("database", a.cfg.Database.DBName),
			zap.String("redis", fmt.Sprintf("%s:%d", a.cfg.Redis.Host, a.cfg.Redis.Port)),
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

// освобождает ресурсы
func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
	_ = a.logger.Sync()
}