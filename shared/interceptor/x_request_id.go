package interceptor

import (
	"context"
	"test-project/shared/requestid"

	"google.golang.org/grpc"
)


func XRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		reqID := requestid.ExtractOrGenerate(ctx)
		newCtx := requestid.WithContext(ctx, reqID)

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
		reqID := requestid.ExtractOrGenerate(ss.Context())
		newCtx := requestid.WithContext(ss.Context(), reqID)
		
		wrappedStream := NewWrappedServerStream(ss, newCtx)
		return handler(srv, wrappedStream)
	}
}
