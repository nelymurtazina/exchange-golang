package main

import (
	"context"
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

	db, err := repository.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresHours)

	userService := service.NewUserService(repo, jwtManager)
	
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

	
	//  GRACEFUL SHUTDOWN
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	lis, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

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