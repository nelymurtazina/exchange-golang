package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggerInterceptor — логирует запросы, НЕ зависит от сервиса
func LoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestID := GetRequestID(ctx)
		startTime := time.Now()
		

		logger.Info("gRPC request started",
			zap.String("method", info.FullMethod),  //огромный объем бесполезных логов,
			zap.String("request_id", requestID),
		)

		resp, err := handler(ctx, req)
		duration := time.Since(startTime)
		code := status.Code(err).String()

		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC request completed",
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("status_code", code),
			)
		}
		return resp, err
	}
}





func LoggerStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		requestID := GetRequestID(ss.Context())
		startTime := time.Now()

		logger.Info("gRPC stream started",
			zap.String("method", info.FullMethod),
			zap.String("request_id", requestID),
		)

		err := handler(srv, ss)

		duration := time.Since(startTime)
		code := status.Code(err).String()

		if err != nil {
			logger.Error("gRPC stream failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("status_code", code),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC stream completed",
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("status_code", code),
			)
		}

		return err
	}
}