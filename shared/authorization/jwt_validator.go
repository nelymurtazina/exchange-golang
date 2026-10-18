package authorization

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)


type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// отвечает за локальную проверку подписи и срока действия Access-токенов.
type TokenValidator struct {
	secretKey string
}

func NewTokenValidator(secretKey string) *TokenValidator {
	return &TokenValidator{secretKey: secretKey}
}

func (v *TokenValidator) ValidateAccessToken(tokenStr string) (map[string]any, error) {
	if len(tokenStr) < 50 || strings.Count(tokenStr, ".") != 2 {
		return nil, errors.New("invalid token format or length")
	}

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
		return nil, errors.New("user_id is empty in token")
	}
	if claims.Role == "" {
		return nil, errors.New("role is empty in token")
	}

	return map[string]any{
		"user_id": claims.UserID,
		"role":    claims.Role,
	}, nil
}