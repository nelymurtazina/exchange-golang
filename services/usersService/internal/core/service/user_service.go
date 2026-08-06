package services

import (
	"context"
	"time"

	"grpc-exchange/services/usersService/internal/core/domain"
	"grpc-exchange/services/usersService/internal/core/ports"
)

type JWTManagerInterface interface {
	GenerateToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(token string) (string, error)
	ValidateRefreshToken(token string) (string, error)
	RefreshToken(refreshToken string) (string, error)
}

type PasswordManagerInterface interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

type userService struct {
	repo ports.UserRepository
	jwt  JWTManagerInterface
	passwd PasswordManagerInterface
}

func NewUserService(repo ports.UserRepository, jwt JWTManagerInterface, passwd PasswordManagerInterface) ports.UserService {
	return &userService{
		repo: repo,
		jwt:  jwt,
		passwd: passwd,
	}
}

func (s *userService) Register(ctx context.Context, input ports.RegisterInput) (*ports.RegisterOutput, error) {
	if err := domain.ValidateUserName(input.Username); err != nil {
		return nil, err
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if input.Password == "" {
		return nil, domain.ErrInvalidPassword
	}

	existing, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := s.passwd.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		UserID:    domain.NewUserID(), // ← UUID!
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GenerateToken(user.UserID)
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

func (s *userService) Login(ctx context.Context, email, password string) (string, string, error) {
	if err := domain.ValidateEmail(email); err != nil {
		return "", "", err
	}
	if password == "" {
		return "", "", domain.ErrInvalidPassword
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", "", domain.ErrUserNotFound
		}
		return "", "", err
	}

	if !s.passwd.CheckPassword(password, user.Password) {
		return "", "", domain.ErrInvalidPassword
	}

	if !user.Active {
		return "", "", domain.ErrUserDisabled
	}

	accessToken, err := s.jwt.GenerateToken(user.UserID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *userService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if err := domain.ValidateUserID(userID); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) ValidateToken(ctx context.Context, tokenString string) (string, string, error) {
	if tokenString == "" {
		return "", "", domain.ErrInvalidToken
	}

	userID, err := s.jwt.ValidateToken(tokenString)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	if err := domain.ValidateUserID(userID); err != nil {
		return "", "", domain.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil || !user.Active {
		return "", "", domain.ErrUserNotFound
	}

	return userID, user.Role, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	if refreshToken == "" {
		return "", "", domain.ErrInvalidToken
	}

	userID, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	if err := domain.ValidateUserID(userID); err != nil {
		return "", "", domain.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil || !user.Active {
		return "", "", domain.ErrUserNotFound
	}

	newAccessToken, err := s.jwt.GenerateToken(userID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *userService) Logout(ctx context.Context, userID string) error {
	if err := domain.ValidateUserID(userID); err != nil {
		return err
	}
	return nil
}
