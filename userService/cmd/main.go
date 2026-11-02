package main

import (
	"log"

	"test-project/userService/internal/app"
	"test-project/userService/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}








// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	pb "test-project/api/gen/user"
// 	"test-project/shared/authorization"
// 	"test-project/shared/interceptor"
// 	"test-project/userService/config"
// 	handler "test-project/userService/internal/adapters/inbound/grpc"
// 	"test-project/userService/internal/adapters/outbound/auth"
// 	"test-project/userService/internal/adapters/outbound/repository/postgres"
// 	"test-project/userService/internal/adapters/outbound/redis"
// 	"test-project/userService/internal/core/service"

// 	redisLib "github.com/redis/go-redis/v9"
// 	_ "github.com/golang-migrate/migrate/v4/database/postgres"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/reflection"
// )

// func main() {
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatalf("failed to create logger: %v", err)
// 	}
// 	defer func() { _ = logger.Sync() }()

// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		logger.Fatal("failed to load config", zap.Error(err))
// 	}

// 	if err := cfg.Validate(); err != nil {
// 		logger.Fatal("invalid config", zap.Error(err))
// 	}

// 	db, err := postgres.NewConnection(cfg.Database)
// 	if err != nil {
// 		logger.Fatal("failed to connect to database", zap.Error(err))
// 	}
// 	defer db.Close()

// 	repo := postgres.NewUserRepository(db)

// 	redisClient := redisLib.NewClient(&redisLib.Options{
// 		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
// 		Password: cfg.Redis.Password,
// 		DB:       cfg.Redis.DB,
// 		//доп настройка для максимального кол-ва подключение и тд. Добавить 
// 	})

// 	if err := redisClient.Ping(context.Background()).Err(); err != nil {
// 		logger.Fatal("failed to connect to redis", zap.Error(err))
// 	}
// 	defer redisClient.Close()

// 	sessionStore := redis.NewSessionStore(redisClient)

// 	jwtValidator := authorization.NewTokenValidator(cfg.JWT.Secret)
	
// 	jwtManager := auth.NewJWTManager(
// 		cfg.JWT.Secret,
// 		cfg.JWT.AccessTTL,
// 		cfg.JWT.RefreshTTL,
// 	)

// 	passwordManager := auth.NewPasswordManager(cfg.Password.BcryptCost)

// 	userService := service.NewUserService(
// 		repo,
// 		jwtManager,
// 		passwordManager,
// 		sessionStore,
// 	)

// 	grpcHandler := handler.NewUserHandler(userService)

// 	skipMethodsMap := make(map[string]struct{})
// 	methodsToSkip := []string{
// 		"/user.v1.UserService/Register",
// 		"/user.v1.UserService/Login",
// 	}
// 	for _, m := range methodsToSkip {
// 		skipMethodsMap[m] = struct{}{}
// 	}

// 	grpcServer := grpc.NewServer(
// 		grpc.ChainUnaryInterceptor(
// 			interceptor.PanicRecoveryInterceptor(logger),
// 			interceptor.XRequestIDInterceptor(),
// 			interceptor.LoggerInterceptor(logger),
// 			interceptor.AuthInterceptor(jwtValidator, logger, skipMethodsMap),
// 		),
// 		grpc.ChainStreamInterceptor(
// 			interceptor.PanicRecoveryStreamInterceptor(logger),
// 			interceptor.XRequestIDStreamInterceptor(),
// 			interceptor.LoggerStreamInterceptor(logger),
// 			interceptor.AuthStreamInterceptor(jwtValidator, logger, skipMethodsMap),
// 		),
// 	)

// 	pb.RegisterUserServiceServer(grpcServer, grpcHandler)
// 	reflection.Register(grpcServer)

// 	lis, err := net.Listen("tcp", cfg.Services.UserServicePort)
// 	if err != nil {
// 		logger.Fatal("failed to listen", zap.Error(err))
// 	}

// 	stop := make(chan os.Signal, 1)
// 	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

// 	go func() {
// 		logger.Info("UserService started", 
// 			zap.String("address", cfg.Services.UserServicePort),
// 			zap.String("database", cfg.Database.DBName),
// 			zap.String("redis", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)),
// 		)
// 		if err := grpcServer.Serve(lis); err != nil {
// 			logger.Error("failed to serve", zap.Error(err))
// 			select {
// 			case stop <- syscall.SIGTERM:
// 			default:
// 			}
// 		}
// 	}()

// 	<-stop
// 	logger.Info("Shutting down gracefully...")

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	shutdownDone := make(chan struct{})
// 	go func() {
// 		grpcServer.GracefulStop()
// 		close(shutdownDone)
// 	}()

// 	select {
// 	case <-shutdownDone:
// 		logger.Info("Server stopped gracefully")
// 	case <-ctx.Done():
// 		logger.Info("Shutdown timeout, forcing stop")
// 		grpcServer.Stop()
// 	}
// }