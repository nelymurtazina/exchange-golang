package ports

import (
	"context"
	"test-project/userService/internal/core/domain"
)

type RegisterInput struct {
	UserName string
	Email    string
	Password string
}

type RegisterOutput struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 
}

type ChangePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}
type ChangePasswordOutput struct {
	Success bool
}

type GetProfilePreviewInput struct {
	UserID string
}

type GetProfilePreviewOutput struct {
	ID       string
	Username string
}

type LogoutInput struct {
	UserID       string
	RefreshToken string
}

type UserService interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	Logout(ctx context.Context, input LogoutInput) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error)
	ChangePassword(ctx context.Context, input ChangePasswordInput) (*ChangePasswordOutput, error)
	GetProfilePreview(ctx context.Context, input GetProfilePreviewInput) (*GetProfilePreviewOutput, error)
}
