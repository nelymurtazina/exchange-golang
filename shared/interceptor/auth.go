package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"test-project/shared/authorization"
	"test-project/shared/requestid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTValidator = authorization.JWTValidator

type userIDKey struct{}
type userRoleKey struct{}  

func AuthInterceptor(validator JWTValidator, logger *zap.Logger, skipMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	skip := make(map[string]struct{}, len(skipMethods))
	for k := range skipMethods {
		skip[k] = struct{}{}
	}
	
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, ok := skip[info.FullMethod]; ok {
        	return handler(ctx, req)
    	}

		claims, err :=  validateTokenFromContext(ctx, validator)
		if err != nil {
			return nil, handleAuthError(err, logger, info.FullMethod, ctx)
		}

		ctx = context.WithValue(ctx, userIDKey{}, claims.UserID)
		ctx = context.WithValue(ctx, userRoleKey{}, claims.Role)

		return handler(ctx, req)
	}
}

func AuthStreamInterceptor(validator JWTValidator, logger *zap.Logger, skipMethods map[string]struct{}) grpc.StreamServerInterceptor {
	skip := make(map[string]struct{}, len(skipMethods))
	for k := range skipMethods {
		skip[k] = struct{}{}
	}
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if _, ok := skip[info.FullMethod]; ok {
       		return handler(srv, ss)
    	}

		claims, err := validateTokenFromContext(ss.Context(), validator)
		if err != nil {
			return handleAuthError(err, logger, info.FullMethod, ss.Context())
		}

		ctx := context.WithValue(ss.Context(), userIDKey{}, claims.UserID)
		ctx = context.WithValue(ctx, userRoleKey{}, claims.Role)

		wrappedStream := NewWrappedServerStream(ss, ctx)

		return handler(srv, wrappedStream)
	}
}

func validateTokenFromContext(ctx context.Context, validator JWTValidator) (*authorization.JWTClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetadataMissing
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, ErrTokenMissing
	}

	// "Bearer <token>" — не чувствительный к регистру
	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, ErrInvalidHeaderFormat
	}

	claims, err := validator.ValidateAccessToken(parts[1])
	if err != nil{
		return nil, fmt.Errorf("%w: %v", ErrTokenValidation, err)
	}
	return claims, nil
}

func handleAuthError(err error, logger *zap.Logger, method string, ctx context.Context) error {
	logFields := []zap.Field{
		zap.String("request_id", requestid.ExtractOrGenerate(ctx)),
		zap.String("method", method),
		zap.Error(err),
	}

	switch {
	case errors.Is(err, ErrMetadataMissing),
		errors.Is(err, ErrTokenMissing),
		errors.Is(err, ErrInvalidHeaderFormat):
		logger.Warn("auth failed: client side issue (bad format)", logFields...)
		return status.Error(codes.Unauthenticated, "missing or invalid authorization header")

	case errors.Is(err, ErrTokenValidation):
		logger.Warn("auth failed: invalid token", logFields...)
		return status.Error(codes.Unauthenticated, "invalid or expired token")

	default:
		// Сюда почти не должны попадать. Клиенту всё равно отдаём Unauthenticated,
		// а не Internal (чтобы не светить внутренние детали).
		logger.Error("auth failed: unexpected error", logFields...)
		return status.Error(codes.Unauthenticated, "authentication failed")
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey{}).(string)
	return role, ok
}