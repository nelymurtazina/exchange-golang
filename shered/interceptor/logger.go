package interceptor

import (
	"context"
	"encoding/json"
	"grpc-exchange/shered/cache"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type LogEntry struct {
    RequestID   string        `json:"request_id"`
    Method      string        `json:"method"`
    StartTime   time.Time     `json:"start_time"`
    Duration    time.Duration `json:"duration_ms"`
    StatusCode  string        `json:"status_code"`
    Error       string        `json:"error,omitempty"`
}

func LoggerInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    logger := zap.L()
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







func LoggerInterceptorWithRedis(logger *zap.Logger, redisCache *cache.RedisCache) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        startTime := time.Now()
        requestID := GetRequestID(ctx)

        logger.Info("gRPC request started",
            zap.String("method", info.FullMethod),
            zap.String("request_id", requestID),
        )

        resp, err := handler(ctx, req)

        duration := time.Since(startTime)
        statusCode := status.Code(err).String()

        logEntry := LogEntry{
            RequestID:  requestID,
            Method:     info.FullMethod,
            StartTime:  startTime,
            Duration:   time.Duration(duration.Milliseconds()),
            StatusCode: statusCode,
        }

        if err != nil {
            logEntry.Error = err.Error()
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

        // Сохраняем лог в Redis 
        if redisCache != nil {
            go saveLogToRedis(redisCache, logEntry)
        }

        return resp, err
    }
}

func saveLogToRedis(redisCache *cache.RedisCache, entry LogEntry) {
    ctx := context.Background()

    // Ключ: log:request_id
    key := "log:" + entry.RequestID

    // Преобразуем в JSON
    data, err := json.Marshal(entry)
    if err != nil {
        return
    }

    // Сохраняем в Redis с TTL 7 дней
    redisCache.Set(ctx, key, data, 7*24*time.Hour)

    // Добавляем request_id в список всех логов
    redisCache.SAdd(ctx, "logs:all", entry.RequestID)

    // Добавляем в список по методу
    redisCache.SAdd(ctx, "logs:method:"+entry.Method, entry.RequestID)
}