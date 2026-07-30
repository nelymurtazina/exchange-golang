package ports

import (
	"context"
	"grpc-exchange/services/usersService/internal/core/domain"
)

type UserService  interface {
	Register(ctx context.Context, username, email, password string) (*domain.User, string, string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	ValidateToken(ctx context.Context, token string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID string) error
}