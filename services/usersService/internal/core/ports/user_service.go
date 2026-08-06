package ports

import (
	"context"
	"grpc-exchange/services/usersService/internal/core/domain"
)

// DTO для регистрации
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// DTO для ответа регистрации
type RegisterOutput struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

type UserService interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	ValidateToken(ctx context.Context, token string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID string) error
}
