package interceptor

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestIDKey struct{}

func XRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		reqID := extractOrGenerateRequestID(ctx)
		newCtx := context.WithValue(ctx, requestIDKey{}, reqID)

		return handler(newCtx, req)
	}
}

func XRequestIDStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		reqID := extractOrGenerateRequestID(ss.Context())
		newCtx := context.WithValue(ss.Context(), requestIDKey{}, reqID)
		
		wrappedStream := NewWrappedServerStream(ss, newCtx)
		return handler(srv, wrappedStream)
	}
}

func extractOrGenerateRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.New().String()
	}

	values := md.Get("x-request-id")
	if len(values) == 0 || values[0] == "" {
		return uuid.New().String() 
	}

	clientID := values[0]

	if len(clientID) > 64 || strings.ContainsAny(clientID, " ,;") {
		return uuid.New().String() 
	}

	return clientID
}