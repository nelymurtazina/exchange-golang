package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTValidator — интерфейс для валидации JWT 
type JWTValidator interface {
	ValidateToken(token string) (string, error) // возвращает userID
}

type userIDKey struct{}

// AuthInterceptor — проверяет JWT ЛОКАЛЬНО, НЕ зависит от UserService
// Принимает любой объект, реализующий JWTValidator
func AuthInterceptor(validator JWTValidator, skipMethods []string) grpc.UnaryServerInterceptor {
	// Создаём map для быстрого поиска
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
		// Проверяем, нужно ли пропустить метод
		if skipMap[info.FullMethod] {
			return handler(ctx, req)
		}

		// Извлекаем токен из метаданных
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

		// Валидация через переданный validator (локально, без gRPC вызовов)
		userID, err := validator.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Сохраняем user_id в контекст
		ctx = context.WithValue(ctx, userIDKey{}, userID)

		return handler(ctx, req)
	}
}

// GetUserIDFromContext — возвращает user_id из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}