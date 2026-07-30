package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type LogEntry struct {
    RequestID   string        `json:"request_id"`
    Method      string        `json:"method"`
    StartTime   time.Time     `json:"start_time"`
    Duration    time.Duration `json:"duration_ms"` //int64
    StatusCode  string        `json:"status_code"`
    Error       string        `json:"error,omitempty"`
}

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
			zap.String("method", info.FullMethod),
			zap.String("request_id", requestID),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}



