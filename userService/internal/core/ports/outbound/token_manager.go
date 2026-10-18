package ports
import "time"

type TokenManager interface {
	GenerateAccessToken(userID string, role string) (string, error)
	GenerateRefreshToken(userID string, role string, jti string) (string, error) 
	ValidateRefreshToken(token string) (userID string, jti string, err error)   
	GetAccessTTL() time.Duration
	GetRefreshTTL() time.Duration
}