package interceptor

import (
	"context"
	userv1 "grpc-exchange/gen/user"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor проверяет JWT токен
func AuthInterceptor(userClient userv1.UserServiceClient) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // Методы, которые не требуют авторизации
        skipMethods := map[string]bool{
            "/user.v1.UserService/Register": true,
            "/user.v1.UserService/Login":    true,
        }
//передаю userd без контекст value
        if skip, ok := skipMethods[info.FullMethod]; ok && skip {
            return handler(ctx, req)
        }

        // Извлекаем токен из метаданных
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
        }

        authHeader := md.Get("authorization")
        if len(authHeader) == 0 { //добавить проверку 
            return nil, status.Errorf(codes.Unauthenticated, "authorization token is missing")
        }

        // "Bearer <token>" - не чуствительный к регистру 
        parts := strings.Split(authHeader[0], " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
        }

        token := parts[1]

        // Проверяем токен через UserService
        resp, err := userClient.ValidateToken(ctx, &userv1.ValidateTokenRequest{
            Token: token,
        })
        if err != nil {
            return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
        }

        if !resp.Valid {
            return nil, status.Errorf(codes.Unauthenticated, "invalid token")
        }

        ctx = context.WithValue(ctx, "user_id", resp.UserId)

        return handler(ctx, req)
    }
}

// GetUserIDFromContext получает user_id из контекста
func GetUserIDFromContext(ctx context.Context) string {
    if userID, ok := ctx.Value("user_id").(string); ok {
        return userID
    }
    return ""
}
//нет user_id. сделать как x-requst-id


//userService упадет - упадет все. Не должна быть зависимость user
//jwt -stcret передаем через. извлекаем jwt-heder -> передаем контексту 



//передавать список конкретных методов из shered 