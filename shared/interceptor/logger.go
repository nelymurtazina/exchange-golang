package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestID := GetRequestID(ctx)
		startTime := time.Now()
		
		logger.Debug("gRPC request started",
			zap.String("request_id", requestID),
			zap.String("method", info.FullMethod),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		code := status.Code(err)

		if err != nil {
			logFields := []zap.Field{
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("status_code", code.String()),
				zap.Error(err),
			}
			// Если код ошибки равен Internal (ошибка сервера) — пишем Error.
			// Если любой другой (пользователь ошибся) — пишем Warn.
			if code == codes.Internal {
				logger.Error("gRPC request failed with internal error", logFields...)
			} else {
				logger.Warn("gRPC request completed with client error", logFields...)
			}
		} else {
			logger.Debug("gRPC request completed",
                zap.String("request_id", requestID),
                zap.String("method", info.FullMethod),
                zap.Duration("duration", duration),
                zap.String("status_code", code.String()),
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

		logger.Debug("gRPC stream started",
			zap.String("request_id", requestID),
		)

		err := handler(srv, ss)

		duration := time.Since(startTime)
		code := status.Code(err)

		if err != nil {
			logFields := []zap.Field{
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("status_code", code.String()),
				zap.Error(err),
			}
			if code == codes.Internal {
				logger.Error("gRPC request failed with internal error", logFields...)
			} else {
				logger.Warn("gRPC request completed with client error", logFields...)
			}
		} else {
			logger.Debug("gRPC stream completed", 
                zap.String("request_id", requestID),
                zap.String("method", info.FullMethod),
                zap.Duration("duration", duration),
                zap.String("status_code", code.String()),
            )
		}

		return err
	}
}