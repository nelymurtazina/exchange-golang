package interceptor

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTValidator interface {
	ValidateAccessToken(token string) (string, error) // возвращает userID
}

type userIDKey struct{}

func AuthInterceptor(validator JWTValidator, skipMethods []string) grpc.UnaryServerInterceptor {
	skipMap := make(map[string]bool, len(skipMethods))
	for _, m := range skipMethods {
		skipMap[m] = true
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if skipMap[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is missing")
		}

		// "Bearer <token>" — не чувствительный к регистру
		parts := strings.SplitN(authHeader[0], " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		token := parts[1]

		userID, err := validator.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, userIDKey{}, userID)

		return handler(ctx, req)
	}
}

// GetUserIDFromContext — возвращает user_id из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}




func AuthStreamInterceptor(validator JWTValidator, logger *zap.Logger, skipMethods []string) grpc.StreamServerInterceptor {
	skipMap := make(map[string]bool, len(skipMethods))
	for _, m := range skipMethods {
		skipMap[m] = true
	}

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		if skipMap[info.FullMethod] {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return status.Errorf(codes.Unauthenticated, "authorization token is missing")
		}

		parts := strings.SplitN(authHeader[0], " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		token := parts[1]

		userID, err := validator.ValidateAccessToken(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Сохраняем user_id в контекст стрима
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          context.WithValue(ctx, userIDKey{}, userID),
		}

		return handler(srv, wrappedStream)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}