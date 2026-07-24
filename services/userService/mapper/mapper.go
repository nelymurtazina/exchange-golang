package mapper

import (
	userv1 "grpc-exchange/gen/user"
	"grpc-exchange/services/userService/domain"
)

func ToProtoUser(user *domain.User) *userv1.User {
    if user == nil {
        return nil
    }
    return &userv1.User{
        UserId:   user.UserID,
        Username: user.Username,
        Email:    user.Email,
        Role:     user.Role,
        Active:   user.Active,
    }
}

func FromProtoUser(req *userv1.GetUserRequest) string {
    return req.UserId
}