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

	pb "grpc-exchange/gen/user"
	"grpc-exchange/services/usersService/config"
	"grpc-exchange/services/usersService/internal/adapters/auth"
	"grpc-exchange/services/usersService/internal/adapters/hendler"
	"grpc-exchange/services/usersService/internal/adapters/repository"
	"grpc-exchange/services/usersService/internal/core/service"
	"grpc-exchange/shered/interceptor"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	if err := cfg.Validate(); err != nil {
		logger.Fatal("invalid config", zap.Error(err))
	}

	db, err := repository.NewConnection(cfg.Database)
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

	repo := repository.NewUserRepository(db)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresHours)
	passwordManager := auth.NewPasswordManager()
	userService := services.NewUserService(repo, jwtManager, passwordManager)
	
	grpcHandler := hendler.NewUserHandler(userService)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.XRequestIDInterceptor,
			interceptor.LoggerInterceptor(logger),
			interceptor.PanicRecoveryInterceptor(logger),
			
		),
	)

	pb.RegisterUserServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	//Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)


	go func() {
		logger.Info("UserService started", 
			zap.String("address", cfg.Server.Port),
		)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
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