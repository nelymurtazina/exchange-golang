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

func AuthInterceptor(validator JWTValidator, logger *zap.Logger, skipMethods []string) grpc.UnaryServerInterceptor {
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

		userID, err := validateTokenFromContext(ctx, validator)
		if err != nil {
			logger.Warn("auth failed",
				zap.String("request_id", GetRequestID(ctx)),
				zap.String("method", info.FullMethod),
			)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, userIDKey{}, userID)

		return handler(ctx, req)
	}
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
		if skipMap[info.FullMethod] {
			return handler(srv, ss)
		}

		userID, err := validateTokenFromContext(ss.Context(), validator)
		
		if err != nil {
			logger.Warn("auth failed",
				zap.String("request_id", GetRequestID(ss.Context())),
				zap.String("method", info.FullMethod),
			)
			return status.Errorf(codes.Unauthenticated, "invalid token")
		}
		
		ctx := context.WithValue(ss.Context(), userIDKey{}, userID)
		wrappedStream := NewWrappedServerStream(ss, ctx)

		return handler(srv, wrappedStream)
	}
}

func validateTokenFromContext(ctx context.Context, validator JWTValidator) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is missing")
	}

	// "Bearer <token>" — не чувствительный к регистру
	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	return validator.ValidateAccessToken(parts[1])
}

// GetUserIDFromContext — возвращает user_id из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}
