package interceptor

import (
    "context"

    "github.com/google/uuid"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

type contextKey string
const RequestIDKey contextKey = "x-request-id"

func XRequestIDInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    requestID := ""
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        if ids := md.Get("x-request-id"); len(ids) > 0 {
            requestID = ids[0]
        }
    }

    if requestID == "" {
        requestID = uuid.New().String()
    }

    ctx = context.WithValue(ctx, RequestIDKey, requestID)
    return handler(ctx, req)
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}
