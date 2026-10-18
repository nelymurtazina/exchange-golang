package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTValidator interface {
    ValidateAccessToken(token string) (claims map[string]any, err error)
} 

type userIDKey struct{}
type userRoleKey struct{}  

func AuthInterceptor(validator JWTValidator, logger *zap.Logger, skipMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, ok := skipMethods[info.FullMethod]; ok {
        	return handler(ctx, req)
    	}

		claims, err :=  validateTokenFromContext(ctx, validator)
		if err != nil {
			logFields := []zap.Field{
				zap.String("request_id", extractOrGenerateRequestID(ctx)),
				zap.String("method", info.FullMethod),
				zap.Error(err),
			}

			if errors.Is(err, ErrMetadataMissing) || errors.Is(err, ErrTokenMissing) || errors.Is(err, ErrInvalidHeaderFormat) {
				logger.Warn("auth failed: client side issue (bad format)", logFields...)
				return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization header")
			}

			if errors.Is(err, ErrTokenValidation) {
				logger.Warn("auth failed: invalid token", logFields...)
				return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
			}

			logger.Error("auth failed: internal system error", logFields...)
			return nil, status.Error(codes.Internal, "internal authentication error")
		}

		userID, err := getStringClaim(claims, "user_id")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid token: %v", err)
		}

		role, err := getStringClaim(claims, "role")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, userIDKey{}, userID)
		ctx = context.WithValue(ctx, userRoleKey{}, role)

		return handler(ctx, req)
	}
}

func AuthStreamInterceptor(validator JWTValidator, logger *zap.Logger, skipMethods map[string]struct{}) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if _, ok := skipMethods[info.FullMethod]; ok {
       		return handler(srv, ss)
    	}

		claims, err := validateTokenFromContext(ss.Context(), validator)
		if err != nil {
			logFields := []zap.Field{
				zap.String("request_id", extractOrGenerateRequestID(ss.Context())),
				zap.String("method", info.FullMethod),
				zap.Error(err),
			}

			if errors.Is(err, ErrMetadataMissing) || errors.Is(err, ErrTokenMissing) || errors.Is(err, ErrInvalidHeaderFormat) {
				logger.Warn("auth failed: client side issue (bad format)", logFields...)
				return status.Error(codes.Unauthenticated, "missing or invalid authorization header")
			}

			if errors.Is(err, ErrTokenValidation) {
				logger.Warn("auth failed: invalid token", logFields...)
				return status.Error(codes.Unauthenticated, "invalid or expired token")
			}

			logger.Error("auth failed: internal system error", logFields...)
			return status.Errorf(codes.Internal, "internal authentication error")
		}

		userID, err := getStringClaim(claims, "user_id")
		if err != nil {
			return status.Errorf(codes.Internal, "invalid token: %v", err)
		}

		role, err := getStringClaim(claims, "role")
		if err != nil {
			return status.Errorf(codes.Internal, "invalid token: %v", err)
		}
		
		ctx := context.WithValue(ss.Context(), userIDKey{}, userID)
		ctx = context.WithValue(ctx, userRoleKey{}, role)
		wrappedStream := NewWrappedServerStream(ss, ctx)

		return handler(srv, wrappedStream)
	}
}

func validateTokenFromContext(ctx context.Context, validator JWTValidator) (map[string]any, error) {
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

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

func GetUserRoleFromContext(ctx context.Context) (string, bool) {
    role, ok := ctx.Value(userRoleKey{}).(string)
    return role, ok
}


func getStringClaim(claims map[string]any, key string) (string, error){
	val, exists := claims[key]
	if !exists{
		return "", fmt.Errorf("claim '%s' is missing", key)
	}

	strVal, ok := val.(string)
	if !ok{
		return "", fmt.Errorf("claim '%s' has invalid type: expected string, got %T", key, val)
	}

	if strVal == ""{
		return "", fmt.Errorf("claim '%s' is empty", key)
	}
	return strVal, nil
}