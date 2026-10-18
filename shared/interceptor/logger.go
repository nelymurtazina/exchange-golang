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
		requestID := extractOrGenerateRequestID(ctx)
		startTime := time.Now()
		
		logger.Debug("gRPC request started",
			zap.String("request_id", requestID),
			zap.String("method", info.FullMethod),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		
		logGRPCRequest(logger, ctx, info.FullMethod, duration, err)
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
		requestID := extractOrGenerateRequestID(ss.Context())
		startTime := time.Now()
		ctx := ss.Context()

		logger.Debug("gRPC stream started",
			zap.String("request_id", requestID),
		)

		err := handler(srv, ss)

		duration := time.Since(startTime)
		
		logGRPCRequest(logger, ctx, info.FullMethod, duration, err)

		return err
	}
}



func logGRPCRequest(logger *zap.Logger, ctx context.Context, method string, duration time.Duration, err error) {
	requestID := extractOrGenerateRequestID(ctx)
	code := status.Code(err)

	logFields := []zap.Field{
		zap.String("method", method),
		zap.String("request_id", requestID),
		zap.Duration("duration", duration),
		zap.String("status_code", code.String()),
	}

	if err != nil {
		logFields = append(logFields, zap.Error(err))

		if code == codes.Internal || code == codes.Unknown || code == codes.Unavailable || code == codes.DataLoss {
			logger.Error("gRPC request failed with server error", logFields...)
			return
		}

		if isExpectedClientError(code) {
			logger.Warn("gRPC request completed with expected client error", logFields...)
			return
		}

		logger.Warn("gRPC request completed with anomalous client error", logFields...)
		
	} else {
		logger.Debug("gRPC request completed successfully",
			append(logFields, zap.String("status", "OK"))...,
		)
	}
}


func isExpectedClientError(code codes.Code) bool {
	switch code {
	case codes.InvalidArgument,      // Ошибка валидации (неверный email, пароль и т.д.)
		codes.NotFound,              // Ресурс не найден (пользователь, ордер)
		codes.Unauthenticated,       // Токен отсутствует или невалиден
		codes.PermissionDenied,      // Нет прав на действие
		codes.AlreadyExists,         // Попытка создать дубликат
		codes.FailedPrecondition:    // Условие не выполнено (рынок закрыт)
		return true
	default:
		return false
	}
}