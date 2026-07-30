package service

import (
	"context"
	"grpc-exchange/services/usersService/internal/core/domain"
	"grpc-exchange/services/usersService/internal/core/ports"
	"time"

	"github.com/google/uuid"
)
type JWTManagerInterface interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	GenerateToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(token string) (string, error)
	ValidateRefreshToken(token string) (string, error) 
	RefreshToken(refreshToken string) (string, error)
}

type userService struct {
	repo ports.UserRepository
	jwt  JWTManagerInterface
}

func NewUserService(repo ports.UserRepository, jwt JWTManagerInterface) ports.UserService {
	return &userService{
		repo: repo,
		jwt:  jwt,
	}
}
func (s *userService) Register(ctx context.Context, username, email, password string) (*domain.User, string, string, error) {
	if username == "" || email == "" || password == "" {
		return nil, "", "", domain.ErrInvalidEmail
	}

	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, "", "", domain.ErrUserAlreadyExists
	}

	hashedPassword, err := s.jwt.HashPassword(password)
	if err != nil {
		return nil, "", "", err
	}

	user := &domain.User{
		UserID:    uuid.New().String(),
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.jwt.GenerateToken(user.UserID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", "", domain.ErrUserNotFound
	}

	if !s.jwt.CheckPassword(password, user.Password) {
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
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// проверка токена
func (s *userService) ValidateToken(ctx context.Context, tokenString string) (string, string, error) {
	userID, err := s.jwt.ValidateToken(tokenString)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil || !user.Active {
		return "", "", domain.ErrUserNotFound
	}

	return userID, user.Role, nil
}

// обновление токена
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	newAccessToken, err := s.jwt.RefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	userID, err := s.jwt.ValidateRefreshToken(refreshToken)
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
	return nil
}