package ports

import (
	"context"
	"grpc-exchange/usersService/internal/core/domain"
)

// RegisterInput — DTO для входа (регистрация)
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// RegisterOutput — DTO для выхода (регистрация)
type RegisterOutput struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

// LoginInput — DTO для входа (логин)
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput — DTO для выхода (логин)
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

// RefreshTokenInput — DTO для обновления токена
type RefreshTokenInput struct {
	RefreshToken string
}

// RefreshTokenOutput — DTO для ответа обновления токена
type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
}

// ValidateTokenInput — DTO для проверки токена
type ValidateTokenInput struct {
	Token string
}

// ValidateTokenOutput — DTO для ответа проверки токена
type ValidateTokenOutput struct {
	UserID string
	Role   string
}

type UserService interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	ValidateToken(ctx context.Context, input ValidateTokenInput) (*ValidateTokenOutput, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error)
	Logout(ctx context.Context, userID string) error
}