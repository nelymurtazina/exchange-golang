package services

import (
	"context"

	"grpc-exchange/usersService/internal/core/domain"
	"grpc-exchange/usersService/internal/core/ports"
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
	repo   ports.UserRepository
	jwt    JWTManagerInterface
	passwd PasswordManagerInterface
}

func NewUserService(
	repo ports.UserRepository,
	jwt JWTManagerInterface,
	passwd PasswordManagerInterface,
) ports.UserService {
	return &userService{
		repo:   repo,
		jwt:    jwt,
		passwd: passwd,
	}
}

// Register — регистрация (с DTO)
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

	// Проверка существования
	existing, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Хэшируем пароль
	hashedPassword, err := s.passwd.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(domain.NewUserID(), input.Username, input.Email, hashedPassword, domain.RoleUser)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Генерируем токены
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

// Login — вход (защита от атак)
func (s *userService) Login(ctx context.Context, input ports.LoginInput) (*ports.LoginOutput, error) {
	//Валидация
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if input.Password == "" {
		return nil, domain.ErrInvalidPassword
	}

	// Ищем пользователя
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Защита от атак — одинаковая ошибка
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Проверяем пароль
	if !s.passwd.CheckPassword(input.Password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// Проверяем активность
	if !user.Active {
		return nil, domain.ErrInvalidCredentials
	}

	// Генерируем токены
	accessToken, err := s.jwt.GenerateToken(user.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
	if err != nil {
		return nil, err
	}

	return &ports.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUser — получение пользователя
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

// ValidateToken — проверка токена (с DTO)
func (s *userService) ValidateToken(ctx context.Context, input ports.ValidateTokenInput) (*ports.ValidateTokenOutput, error) {
	if input.Token == "" {
		return nil, domain.ErrInvalidToken
	}

	userID, err := s.jwt.ValidateToken(input.Token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if err := domain.ValidateUserID(userID); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Active {
		return nil, domain.ErrUserNotFound
	}

	return &ports.ValidateTokenOutput{
		UserID: userID,
		Role:   user.Role,
	}, nil
}

// RefreshToken — обновление токена (с DTO)
func (s *userService) RefreshToken(ctx context.Context, input ports.RefreshTokenInput) (*ports.RefreshTokenOutput, error) {
	if input.RefreshToken == "" {
		return nil, domain.ErrInvalidToken
	}

	userID, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if err := domain.ValidateUserID(userID); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Active {
		return nil, domain.ErrUserNotFound
	}

	newAccessToken, err := s.jwt.GenerateToken(userID)
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

// Logout — выход
func (s *userService) Logout(ctx context.Context, userID string) error {
	if err := domain.ValidateUserID(userID); err != nil {
		return err
	}
	return nil
}