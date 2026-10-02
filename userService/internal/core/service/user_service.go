package service

import (
	"context"
	"errors"
	"fmt"
	"test-project/userService/internal/core/domain"
	ports "test-project/userService/internal/core/ports/outbound"
    portsInbound "test-project/userService/internal/core/ports/inbound"
	"time"
)

type UserService struct {
	repo   ports.UserRepository
	jwt    JWTManagerInterface
	passwd PasswordManagerInterface
}

type JWTManagerInterface interface {
    GenerateAccessToken(userID string, role string) (string, error)          
    GenerateRefreshToken(userID string, role string) (string, error)        
    ValidateAccessToken(accessToken string) (string, string, error)           
    ValidateRefreshToken(accessToken string) (string, string, error)         
    RefreshToken(refreshToken string) (string, error)
}

type PasswordManagerInterface interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

func NewUserService(repo ports.UserRepository, jwt JWTManagerInterface, passwd PasswordManagerInterface) portsInbound.UserService {
	return &UserService{
		repo:   repo,
		jwt:    jwt,
		passwd: passwd,
	}
}

func (s *UserService) Register(ctx context.Context, input portsInbound.RegisterInput) (*portsInbound.RegisterOutput, error) {
	if err := domain.ValidateUserName(input.UserName); err != nil {
		return nil, err
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, domain.ErrInvalidEmail
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	existingUser, err := s.repo.GetByEmail(ctx, input.Email)
    if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
        return nil, fmt.Errorf("failed to check user: %w", err) 
    }
    if existingUser != nil {
        return nil, domain.ErrUserAlreadyExists
    }

	hashedPassword, err := s.passwd.HashPassword(input.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    now := time.Now()

	user, err := domain.NewUser(domain.NewUserID(), input.UserName, input.Email, hashedPassword, domain.RoleUser, now, now)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}

	return &portsInbound.RegisterOutput{
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

    return user, nil
}

func (s *UserService) Login(ctx context.Context, input portsInbound.LoginInput) (*portsInbound.LoginOutput, error) {
    if err := domain.ValidateEmail(input.Email); err != nil {
        return nil, domain.ErrInvalidEmail
    }
    if err := domain.ValidatePassword(input.Password); err != nil {
        return nil, err
    }

    user, err := s.repo.GetByEmail(ctx, input.Email)
    if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
        return nil, err
    }

    if user == nil {
        fakeHash := "$2a$12$c0XVyLJxfoQJsv6YLjH1m.39h7J2Jas/BJ..XqBUkonHiYaU6mFOa"
        _ = s.passwd.CheckPassword(input.Password, fakeHash)
        return nil, domain.ErrInvalidCredentials
    }
    //можно игнорировать резултаты. 

    if !s.passwd.CheckPassword(input.Password, user.PasswordHash) {
        return nil, domain.ErrInvalidCredentials
    }

    accessToken, err := s.jwt.GenerateAccessToken(user.UserID, user.Role)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID, user.Role)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    return &portsInbound.LoginOutput{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

func (s *UserService) ValidateToken(ctx context.Context, input portsInbound.ValidateTokenInput) (*portsInbound.ValidateTokenOutput, error) {
    if input.Token == "" {
        return nil, domain.ErrInvalidToken
    }

    userID, role, err := s.jwt.ValidateAccessToken(input.Token)
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

    return &portsInbound.ValidateTokenOutput{
        UserID: userID,
        Role:   role,
    }, nil
}

func (s *UserService) RefreshToken(ctx context.Context, input portsInbound.RefreshTokenInput) (*portsInbound.RefreshTokenOutput, error) {
    if input.RefreshToken == "" {
        return nil, domain.ErrInvalidToken
    }

    userID, role, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
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

    newAccessToken, err := s.jwt.GenerateAccessToken(userID, role)
    if err != nil {
        return nil, err
    }

    newRefreshToken, err := s.jwt.GenerateRefreshToken(userID, user.Role)
    if err != nil {
        return nil, err
    }

    return &portsInbound.RefreshTokenOutput{
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

func (s *UserService) ChangePassword(ctx context.Context, input portsInbound.ChangePasswordInput) (*portsInbound.ChangePasswordOutput, error) {
    if err := domain.ValidateEmail(input.Email); err != nil {
        return nil, err
    }
    if err := domain.ValidatePassword(input.NewPassword); err != nil {
        return nil, err
    }

    user, err := s.repo.GetByEmail(ctx, input.Email)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, domain.ErrInvalidCredentials
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if !s.passwd.CheckPassword(input.CurrentPassword, user.PasswordHash) {
        return nil, domain.ErrInvalidCredentials
    }

    hashedPassword, err := s.passwd.HashPassword(input.NewPassword)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    if err := s.repo.UpdatePassword(ctx, user.UserID, hashedPassword); err != nil {
        return nil, fmt.Errorf("failed to update password: %w", err)
    }

    return &portsInbound.ChangePasswordOutput{Success: true}, nil
}

func (s *UserService) GetProfilePreview(ctx context.Context, input portsInbound.GetProfilePreviewInput) (*portsInbound.GetProfilePreviewOutput, error) {
    if err := domain.ValidateID(input.UserID); err != nil {
        return nil, err
    }

    user, err := s.repo.GetByID(ctx, input.UserID)
    if err != nil {
        return nil, err  
    }

    return &portsInbound.GetProfilePreviewOutput{
        ID:       user.UserID,
        Username: user.UserName,
    }, nil
}