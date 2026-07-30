package main

// import (
// 	"log"
// 	"net"
// 	"net/http"

// 	spotv1 "grpc-exchange/gen/spot"
// 	"grpc-exchange/services/spotInstrumentService/repository"
// 	"grpc-exchange/services/spotInstrumentService/service"
// 	"grpc-exchange/shered/config"
// 	// "grpc-exchange/shered/database"
// 	"grpc-exchange/shered/interceptor"

// 	"github.com/prometheus/client_golang/prometheus/promhttp"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/reflection"
// )

// func main() {
//     logger, err := zap.NewProduction()
//     if err != nil {
//         log.Fatalf("failed to create logger: %v", err)
//     }
//     defer logger.Sync()
    
//     cfg := config.LoadConfig()
    
//     // db, err := database.NewPostgresConnection(database.Config(cfg.Database))
//     // if err != nil {
//     //     logger.Fatal("failed to connect to database", zap.Error(err))
//     // }
//     // defer db.Close()
    
//     // repo := repository.NewPostgresMarketRepository(db)
    
//     // instrumentSrv := service.NewInstrumentService(repo)
    
//     metrics := interceptor.NewMetricsCollector()
    
//     grpcServer := grpc.NewServer(
//         grpc.ChainUnaryInterceptor(
//             interceptor.XRequestIDInterceptor,
//             metrics.UnaryServerInterceptor(),
//         ),
//     )
    
//     spotv1.RegisterSpotInstrumentServiceServer(grpcServer, instrumentSrv)
//     reflection.Register(grpcServer)

// // Запускаем HTTP сервер для Prometheus
// go func() {
//     http.Handle("/metrics", promhttp.Handler())
//     logger.Info("Prometheus metrics server started", 
//         zap.String("address", ":9092")) 
//     if err := http.ListenAndServe(":9092", nil); err != nil {
//         logger.Error("Prometheus metrics server failed", zap.Error(err))
//     }
// }()
    
//     lis, err := net.Listen("tcp", cfg.Services.InstrumentServicePort)
//     if err != nil {
//         logger.Fatal("failed to listen", zap.Error(err))
//     }
    
//     logger.Info("SpotInstrumentService started",
//         zap.String("address", cfg.Services.InstrumentServicePort))
    
//     if err := grpcServer.Serve(lis); err != nil {
//         logger.Fatal("failed to serve", zap.Error(err))
//     }
// }