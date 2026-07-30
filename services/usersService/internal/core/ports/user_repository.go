package ports

import (
	"context"
	"grpc-exchange/services/usersService/internal/core/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}