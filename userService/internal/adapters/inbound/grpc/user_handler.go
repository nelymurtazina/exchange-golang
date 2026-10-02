package hendler

import (
	"context"
	"errors"

	pb "test-project/api/gen/user"
	"test-project/shared/interceptor"
	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports/inbound"

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
		UserName: req.Username,
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
		AccessToken:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
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
		AccessToken:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID, ok := interceptor.GetUserIDFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "unauthenticated")
    }
	user, err := h.service.GetUser(ctx, userID)
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

	return &pb.GetUserResponse{
		UserId:   user.UserID,
		Username: user.UserName,
		Email:    user.Email,
		Role:     domainRoleToProto(user.Role),
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	output, err := h.service.ValidateToken(ctx, ports.ValidateTokenInput{
		Token: req.AccessToken,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &pb.ValidateTokenResponse{
		UserId: output.UserID,
		Role:   domainRoleToProto(output.Role),
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
		AccessToken:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	userID, ok := interceptor.GetUserIDFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "user not authenticated")
    }
	err := h.service.Logout(ctx, userID)
    if err != nil {
        return nil, status.Error(codes.Internal, "internal error")
    }

    return &pb.LogoutResponse{}, nil
}


func domainRoleToProto(role string) pb.Role {
    switch role {
    case "admin":
        return pb.Role_ROLE_ADMIN
    case "guest":
        return pb.Role_ROLE_GUEST
    case "user":
        return pb.Role_ROLE_USER
    default:
        return pb.Role_ROLE_UNSPECIFIED
    }
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error){
	output, err := h.service.ChangePassword(ctx, ports.ChangePasswordInput{
		Email: req.Email,
		CurrentPassword: req.CurrentPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		switch{
		case errors.Is(err, domain.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.Is(err, domain.ErrInvalidEmail),
            errors.Is(err, domain.ErrInvalidPassword):
            return nil, status.Error(codes.InvalidArgument, err.Error())
        default:
            return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &pb.ChangePasswordResponse{
		Success: output.Success,
		Message: "Password changed successfully",
	}, nil
}

func (h *UserHandler) GetProfilePreview(ctx context.Context, req *pb.GetUserRequest) (*pb.GetProfilePreviewResponse, error) {
    output, err := h.service.GetProfilePreview(ctx, ports.GetProfilePreviewInput{
        UserID: req.UserId,
    })
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, "internal error")
    }

    return &pb.GetProfilePreviewResponse{
        Profile: &pb.UserProfilePreview{
            Id:       output.ID,
            Username: output.Username,
        },
    }, nil
}