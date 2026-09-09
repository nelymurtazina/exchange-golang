package service

import (
	"context"
	"errors"
	"fmt"
	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports"
)

type UserService struct {
	repo   ports.UserRepository
	jwt    JWTManagerInterface
	passwd PasswordManagerInterface
}

type JWTManagerInterface interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(accessToken string) (string, error)
	ValidateRefreshToken(accessToken string) (string, error)
	RefreshToken(refreshToken string) (string, error)
}

type PasswordManagerInterface interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

func NewUserService(repo ports.UserRepository, jwt JWTManagerInterface, passwd PasswordManagerInterface) ports.UserService {
	return &UserService{
		repo:   repo,
		jwt:    jwt,
		passwd: passwd,
	}
}

func (s *UserService) Register(ctx context.Context, input ports.RegisterInput) (*ports.RegisterOutput, error) {
	if err := domain.ValidateUserName(input.UserName); err != nil {
		return nil, domain.ErrInvalidUsername
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, domain.ErrInvalidEmail
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	existingUser, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, domain.ErrInvalidEmail
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPaswword, err := s.passwd.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(domain.NewUserID(), input.UserName, input.Email, hashedPaswword, domain.RoleUser)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
	if err != nil {
		return nil, err
	}

	return &ports.RegisterOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
    if err := domain.ValidateID(userID); err != nil {
        return nil, err
    }

    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if user == nil {
        return nil, domain.ErrUserNotFound
    }

    if user.DeletedAt != nil && !user.DeletedAt.IsZero() {
        return nil, domain.ErrUserNotFound
    }

    return user, nil
}

func (s *UserService) Login(ctx context.Context, input ports.LoginInput) (*ports.LoginOutput, error) {
    if err := domain.ValidateEmail(input.Email); err != nil {
        return nil, domain.ErrInvalidEmail
    }
    if input.Password == "" {
        return nil, domain.ErrInvalidPassword
    }

    user, err := s.repo.GetByEmail(ctx, input.Email)
    if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
        return nil, err
    }

    if user == nil {
        fakeHash := "$2a$12$c0XVyLJxfoQJsv6YLjH1m.39h7J2Jas/BJ..XqBUkonHiYaU6mFOa"
        s.passwd.CheckPassword(input.Password, fakeHash)
        return nil, domain.ErrInvalidCredentials
    }

    if user.DeletedAt != nil && !user.DeletedAt.IsZero() {
        return nil, domain.ErrInvalidCredentials
    }

    if !s.passwd.CheckPassword(input.Password, user.Password) {
        return nil, domain.ErrInvalidCredentials
    }

    accessToken, err := s.jwt.GenerateAccessToken(user.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    return &ports.LoginOutput{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

func (s *UserService) ValidateToken(ctx context.Context, input ports.ValidateTokenInput) (*ports.ValidateTokenOutput, error) {
    if input.Token == "" {
        return nil, domain.ErrInvalidToken
    }

    userID, err := s.jwt.ValidateAccessToken(input.Token)
    if err != nil {
        return nil, domain.ErrInvalidToken
    }

    if err := domain.ValidateID(userID); err != nil {
        return nil, domain.ErrInvalidUserID
    }

    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, domain.ErrInvalidToken
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if user == nil {
        return nil, domain.ErrInvalidToken
    }

    if user.DeletedAt != nil && !user.DeletedAt.IsZero() {
        return nil, domain.ErrInvalidToken
    }

    return &ports.ValidateTokenOutput{
        UserID: userID,
        Role:   user.Role,
    }, nil
}

func (s *UserService) RefreshToken(ctx context.Context, input ports.RefreshTokenInput) (*ports.RefreshTokenOutput, error) {
    if input.RefreshToken == "" {
        return nil, domain.ErrInvalidToken
    }

    userID, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
    if err != nil {
        return nil, domain.ErrInvalidToken
    }

    if err := domain.ValidateID(userID); err != nil {
        return nil, domain.ErrInvalidUserID
    }

    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, domain.ErrInvalidToken
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if user == nil {
        return nil, domain.ErrInvalidToken
    }

    if user.DeletedAt != nil && !user.DeletedAt.IsZero() {
        return nil, domain.ErrInvalidToken
    }

    newAccessToken, err := s.jwt.GenerateAccessToken(userID)
    if err != nil {
        return nil, err
    }

    newRefreshToken, err := s.jwt.GenerateRefreshToken(userID)
    if err != nil {
        return nil, err
    }

    return &ports.RefreshTokenOutput{
        AccessToken:  newAccessToken,
        RefreshToken: newRefreshToken,
    }, nil
}

func (s *UserService) Logout(ctx context.Context, userID string) error {
	if err := domain.ValidateID(userID); err != nil {
		return domain.ErrInvalidUserID
	}

	return nil
}
