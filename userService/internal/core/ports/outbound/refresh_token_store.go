package ports

import (
	"context"
	"time"
)

type RefreshTokenStore interface {
	Save(ctx context.Context, userID string, jti string, ttl time.Duration) error
	GetUserID(ctx context.Context, jti string) (string, error)
	Delete(ctx context.Context, jti string) error
}