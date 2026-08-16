package hendler

import (
	"context"

	pb "grpc-exchange/gen/user"
	"grpc-exchange/usersService/internal/core/domain"
	"grpc-exchange/usersService/internal/core/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service ports.UserService
}

func NewUserHandler(service ports.UserService) *UserHandler{
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	output, err := h.service.Register(ctx, ports.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch err {
		case domain.ErrInvalidUsername, domain.ErrInvalidEmail:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case domain.ErrUserAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.RegisterResponse{
		UserId:       output.User.UserID,
		Token:        output.AccessToken,
		RefreshToken: output.RefreshToken,
		Message:      "User registered successfully",
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	//Использую DTO
	output, err := h.service.Login(ctx, ports.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials, domain.ErrInvalidEmail, domain.ErrInvalidPassword:
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.LoginResponse{
		Token:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

// gRPC бизнес-логика
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := h.service.GetUser(ctx, req.UserId)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		case domain.ErrInvalidUserID:
			return nil, status.Error(codes.InvalidArgument, "invalid user_id")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.User{
		UserId:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Active:   user.Active,
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	output, err := h.service.ValidateToken(ctx, ports.ValidateTokenInput{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &pb.ValidateTokenResponse{
		UserId: output.UserID,
		Role:   output.Role,
	}, nil
}

func (h *UserHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	output, err := h.service.RefreshToken(ctx, ports.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &pb.RefreshTokenResponse{
		Token:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.service.Logout(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.LogoutResponse{}, nil
}