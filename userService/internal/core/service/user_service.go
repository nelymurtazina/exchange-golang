package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"test-project/userService/internal/core/domain"
	portsInbound "test-project/userService/internal/core/ports/inbound"
	ports "test-project/userService/internal/core/ports/outbound"

	"github.com/google/uuid"
)

type UserService struct {
	repo     ports.UserRepository
	jwt      ports.TokenManager
	passwd   ports.PasswordHasher
	sessions ports.RefreshTokenStore
}

func NewUserService(
	repo ports.UserRepository,
	jwt ports.TokenManager,
	passwd ports.PasswordHasher,
	sessions ports.RefreshTokenStore,
) portsInbound.UserService {
	return &UserService{
		repo:     repo,
		jwt:      jwt,
		passwd:   passwd,
		sessions: sessions,
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

	accessToken, refreshToken, err := s.issueTokens(ctx, user)
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
		return nil, domain.ErrInvalidPassword
	}

	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		fakeHash := "$2a$12$c0XVyLJxfoQJsv6YLjH1m.39h7J2Jas/BJ..XqBUkonHiYaU6mFOa"
		_ = s.passwd.CheckPassword(input.Password, fakeHash)
		return nil, domain.ErrInvalidCredentials
	}

	if !s.passwd.CheckPassword(input.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &portsInbound.LoginOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwt.GetAccessTTL().Seconds()), 
	}, nil
}

func (s *UserService) RefreshToken(ctx context.Context, input portsInbound.RefreshTokenInput) (*portsInbound.RefreshTokenOutput, error) {
	if input.RefreshToken == "" {
		return nil, domain.ErrInvalidToken
	}

	userID, jti, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if err := domain.ValidateID(userID); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	sessionUserID, err := s.sessions.GetUserID(ctx, jti)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if sessionUserID != userID {
		return nil, domain.ErrInvalidToken
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

	accessToken, refreshToken, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := s.sessions.Delete(ctx, jti); err != nil {
		return nil, fmt.Errorf("failed to delete old session: %w", err)
	}

	return &portsInbound.RefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwt.GetAccessTTL().Seconds()),
	}, nil
}


func (s *UserService) Logout(ctx context.Context, input portsInbound.LogoutInput) error {
	if err := domain.ValidateID(input.UserID); err != nil {
		return domain.ErrInvalidUserID
	}
	if input.RefreshToken == "" {
		return domain.ErrInvalidToken
	}

	tokenUserID, jti, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return domain.ErrInvalidToken
	}
	if tokenUserID != input.UserID {
		return domain.ErrInvalidToken
	}

	return s.sessions.Delete(ctx, jti)
}

func (s *UserService) ChangePassword(ctx context.Context, input portsInbound.ChangePasswordInput) (*portsInbound.ChangePasswordOutput, error) {
	if err := domain.ValidatePassword(input.NewPassword); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, input.UserID)
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

	if err := s.sessions.DeleteAllForUser(ctx, user.UserID); err != nil {
		return nil, fmt.Errorf("password updated, but failed to invalidate sessions: %w", err)
	}

	return &portsInbound.ChangePasswordOutput{Success: true}, nil
}


func (s *UserService) GetProfilePreview(ctx context.Context, input portsInbound.GetProfilePreviewInput) (*portsInbound.GetProfilePreviewOutput, error) {
	if err := domain.ValidateID(input.UserID); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &portsInbound.GetProfilePreviewOutput{
		ID:       user.UserID,
		Username: user.UserName,
	}, nil
}

func (s *UserService) issueTokens(ctx context.Context, user *domain.User) (accessToken, refreshToken string, err error) {
	jti := uuid.New().String()

	accessToken, err = s.jwt.GenerateAccessToken(user.UserID, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.jwt.GenerateRefreshToken(user.UserID, user.Role, jti)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.sessions.Save(ctx, user.UserID, jti, s.jwt.GetRefreshTTL()); err != nil {
		return "", "", fmt.Errorf("failed to save session: %w", err)
	}

	return accessToken, refreshToken, nil
}

