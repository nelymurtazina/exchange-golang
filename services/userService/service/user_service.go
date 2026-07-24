package service

import (
	"context"
	"time"

	userv1 "grpc-exchange/gen/user"
	"grpc-exchange/services/userService/auth"
	"grpc-exchange/services/userService/domain"
	"grpc-exchange/services/userService/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	userv1.UnimplementedUserServiceServer
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	// Проверяем обязательные поля
	if req.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	// Проверяем, существует ли пользователь
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check user: %v", err)
	}
	if existing != nil {
		return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
	}

	// Хэшируем пароль
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	user := &domain.User{
		UserID:    uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userv1.RegisterResponse{
		UserId:  user.UserID,
		Message: "User registered successfully",
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	// поиск user
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// проверяем пароль
	if !auth.CheckPassword(req.Password, user.Password) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password")
	}

	if !user.Active {
		return nil, status.Errorf(codes.PermissionDenied, "user account is disabled")
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(user.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &userv1.LoginResponse{
		UserId:  user.UserID,
		Token:   token,
		Message: "Login successful",
	}, nil
}


func (s *UserService) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	return &userv1.LogoutResponse{
		Message: "Logout successful",
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.User, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.repo.GetByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &userv1.User{
		UserId:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Active:   user.Active,
	}, nil
}

// ValidateToken — проверяет JWT токен
func (s *UserService) ValidateToken(ctx context.Context, req *userv1.ValidateTokenRequest) (*userv1.ValidateTokenResponse, error) {
	if req.Token == "" {
		return &userv1.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	userID, err := auth.ValidateToken(req.Token)
	if err != nil {
		return &userv1.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return &userv1.ValidateTokenResponse{
			Valid: false,
		}, nil
	}
	if user == nil || !user.Active {
		return &userv1.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &userv1.ValidateTokenResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}