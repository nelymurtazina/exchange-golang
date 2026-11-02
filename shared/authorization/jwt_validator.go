package authorization

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTValidator interface {
	ValidateAccessToken(token string) (*JWTClaims, error) 
}

// отвечает за локальную проверку подписи и срока действия Access-токенов.
type TokenValidator struct {
	secretKey string
}

func NewTokenValidator(secretKey string) *TokenValidator {
	return &TokenValidator{secretKey: secretKey}
}

func (v *TokenValidator) ValidateAccessToken(tokenStr string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UserID == "" {
		return nil, errors.New("claim 'user_id' is missing or empty")
	}
	if claims.Role == "" {
		return nil, errors.New("claim 'role' is missing or empty")
	}

	return claims, nil
}