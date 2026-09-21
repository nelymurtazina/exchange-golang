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
    ValidateAccessToken(token string) (userID string, role string, err error)
}

type userIDKey struct{}
type userRoleKey struct{}

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

		userID, role, err := validateTokenFromContext(ctx, validator)
		if err != nil {
			logger.Debug("auth failed",
				zap.String("request_id", GetRequestID(ctx)),
				zap.String("method", info.FullMethod),
				zap.Error(err),  
			)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, userIDKey{}, userID)
		ctx = context.WithValue(ctx, userRoleKey{}, role)

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

		userID, role, err := validateTokenFromContext(ss.Context(), validator)
		
		if err != nil {
			logger.Warn("auth failed",
				zap.String("request_id", GetRequestID(ss.Context())),
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return status.Errorf(codes.Unauthenticated, "invalid token")
		}
		
		ctx := context.WithValue(ss.Context(), userIDKey{}, userID)
		ctx = context.WithValue(ctx, userRoleKey{}, role)
		wrappedStream := NewWrappedServerStream(ss, ctx)

		return handler(srv, wrappedStream)
	}
}

func validateTokenFromContext(ctx context.Context, validator JWTValidator) (string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "" ,status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", "", status.Errorf(codes.Unauthenticated, "authorization token is missing")
	}

	// "Bearer <token>" — не чувствительный к регистру
	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "","", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	return validator.ValidateAccessToken(parts[1])
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

func GetUserRoleFromContext(ctx context.Context) (string, bool) {
    role, ok := ctx.Value(userRoleKey{}).(string)
    return role, ok
}