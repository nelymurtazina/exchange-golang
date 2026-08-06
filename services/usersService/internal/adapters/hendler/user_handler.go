package hendler

import (
	"context"

	pb "grpc-exchange/gen/user"
	"grpc-exchange/services/usersService/internal/core/domain"
	"grpc-exchange/services/usersService/internal/core/ports"

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
	accessToken, refreshToken, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound, domain.ErrInvalidPassword:
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case domain.ErrUserDisabled:
			return nil, status.Error(codes.PermissionDenied, "user disabled")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		Message:      "Login successful",
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
	userID, role, err := h.service.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		UserId: userID,
		Valid:  true,
		Role:   role,
	}, nil
}

func (h *UserHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	newAccessToken, newRefreshToken, err := h.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.RefreshTokenResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		Message:      "Token refreshed successfully",
	}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.service.Logout(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.LogoutResponse{
		Message: "Logout successful",
	}, nil
}