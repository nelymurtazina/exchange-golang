package main

import (
	"context"
	"log"
	"net"
	"net/http"

	userv1 "grpc-exchange/gen/user"
	"grpc-exchange/services/userService/auth"
	"grpc-exchange/services/userService/repository"
	"grpc-exchange/services/userService/service"
	"grpc-exchange/shered/cache"
	"grpc-exchange/shered/config"
	"grpc-exchange/shered/database"
	"grpc-exchange/shered/interceptor"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	auth.Init(cfg.JWT)
	
	
	db, err := database.NewPostgresConnection(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,  
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisCache := cache.NewRedisCache(cfg.Redis)
    if err := redisCache.Ping(context.Background()); err != nil {
        logger.Warn("Redis not available, logs will not be stored", zap.Error(err))
    } else {
        logger.Info("Redis connected successfully")
    }
	
	repo := repository.NewPostgresUserRepository(db)
	userSrv := service.NewUserService(repo)
	
	metrics := interceptor.NewMetricsCollector()
	
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.XRequestIDInterceptor,
			interceptor.LoggerInterceptorWithRedis(logger, redisCache),
			interceptor.LoggerInterceptor,
			interceptor.PanicRecoveryInterceptor,
			metrics.UnaryServerInterceptor(),
		),
	)
	
	userv1.RegisterUserServiceServer(grpcServer, userSrv)
	reflection.Register(grpcServer)


	go func() {
        http.Handle("/metrics", promhttp.Handler())
        logger.Info("Prometheus metrics server started", zap.String("address", cfg.Metrics.UserPort))
        if err := http.ListenAndServe(cfg.Metrics.UserPort, nil); err != nil {
            logger.Error("Prometheus metrics server failed", zap.Error(err))
        }
    }()
	
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	
	logger.Info("UserService started on :50053", zap.String("address", cfg.Services.UserServicePort))
	
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}