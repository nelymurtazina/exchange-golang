package auth

import (
	"errors"
	"fmt"
	outbound "test-project/spotService/internal/core/ports/outbound"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenValidator struct {
	secretKey string
}

func NewTokenValidator(secretKey string) outbound.JWTValidator {
	return &TokenValidator{secretKey: secretKey}
}

func (v *TokenValidator) ValidateAccessToken(tokenStr string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.secretKey), nil
	})

	if err != nil {
		return "", "", err 
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid token claims")
	}

	return claims.UserID, claims.Role, nil
}
