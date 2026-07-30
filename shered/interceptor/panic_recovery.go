package interceptor

import (
    "context"
    "runtime/debug"

    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// PanicRecoveryInterceptor — ловит панику и возвращает Internal ошибку
// ПЕРЕДАЁМ ЛОГГЕР КАК ПАРАМЕТР (не создаём внутри!)
func PanicRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (resp interface{}, err error) {
        requestID := GetRequestID(ctx)

        defer func() {
            if r := recover(); r != nil {
                // Используем переданный логгер
                logger.Error("panic recovered",
                    zap.Any("panic", r),
                    zap.String("method", info.FullMethod),
                    zap.String("request_id", requestID),
                    zap.String("stack", string(debug.Stack())),
                )
                err = status.Errorf(codes.Internal, "internal server error")
                resp = nil
            }
        }()

        return handler(ctx, req)
    }
}

// детали паники, мы должны логировать исключительно на сервере, клиенту без деталей

//передавать эксепляр zap логгер в интерсептор при инициализации 