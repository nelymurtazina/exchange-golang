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
	User *domain.User
	AccessToken string
	RefreshToken string
}

type LoginInput struct{
	Email string
	Password string
}

type LoginOutput struct {
	User *domain.User
	AccessToken  string
	RefreshToken string
}

type ValidateTokenInput struct {
	Token string
}

type ValidateTokenOutput struct {
	UserID string
	Role   string
}

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshTokenOutput struct {
	User domain.User
	AccessToken  string
	RefreshToken string
}

type ChangePasswordInput struct{
	Email string
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

type UserService interface{
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	Logout(ctx context.Context, userID string) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	ValidateToken(ctx context.Context, input ValidateTokenInput) (*ValidateTokenOutput, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error)
	ChangePassword(ctx context.Context, input ChangePasswordInput) (*ChangePasswordOutput, error)
	GetProfilePreview(ctx context.Context, input GetProfilePreviewInput) (*GetProfilePreviewOutput, error)
}
