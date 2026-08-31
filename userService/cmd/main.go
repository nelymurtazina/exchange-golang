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

	pb "test-project/api/gen/user"
	"test-project/shered/interceptor"
	"test-project/userService/config"
	"test-project/userService/internal/adapters/auth"
	hendler "test-project/userService/internal/adapters/handler/grpc"
	"test-project/userService/internal/adapters/repository/postgres"
	"test-project/userService/internal/core/service"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

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

	repo := postgres.NewUserRepository(db)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresHours)
	passwordManager := auth.NewPasswordManager()
	userService := service.NewUserService(repo, jwtManager, passwordManager)
	
	grpcHandler := hendler.NewUserHandler(userService)
	skipMethods := []string{
		"/user.v1.UserService/Register",
		"/user.v1.UserService/Login",
	}
	grpcServer := grpc.NewServer( 
		grpc.ChainUnaryInterceptor(
			interceptor.XRequestIDInterceptor(),
			interceptor.PanicRecoveryInterceptor(logger),
			interceptor.AuthInterceptor(jwtManager, skipMethods), 
			interceptor.LoggerInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			interceptor.PanicRecoveryStreamInterceptor(logger),
			interceptor.XRequestIDStreamInterceptor(),
			interceptor.LoggerStreamInterceptor(logger),
			interceptor.AuthStreamInterceptor(jwtManager, logger, skipMethods),
		),

	)

	pb.RegisterUserServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.Services.UserServicePort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)


	go func() {
		logger.Info("UserService started", zap.String("address", cfg.Services.UserServicePort))
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

	if err := db.Close(); err != nil {
		logger.Error("failed to close database connection", zap.Error(err))
	}
}