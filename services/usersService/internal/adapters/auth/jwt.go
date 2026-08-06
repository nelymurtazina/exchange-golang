package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct { //знает только о токенах , не должен кэшировать 
	secret       string
	expiresHours int
} //разные доменные задачи, должна вынести

func NewJWTManager(secret string, expiresHours int) *JWTManager {
	return &JWTManager{
		secret:       secret,
		expiresHours: expiresHours,
	}
}

// GenerateToken — создаёт Access Token
func (m *JWTManager) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(m.expiresHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// GenerateRefreshToken — создаёт Refresh Token (живёт 7 дней)
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(m.expiresHours*24*7) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"refresh": true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// ValidateToken — проверяет Access Token
func (m *JWTManager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверяем, что это не refresh token
		if isRefresh, ok := claims["refresh"].(bool); ok && isRefresh {
			return "", errors.New("refresh token cannot be used as access token")
		}
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("invalid user_id in token")
		}
		return userID, nil
	}

	return "", errors.New("invalid token")
}

// ValidateRefreshToken — проверяет Refresh Token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		isRefresh, ok := claims["refresh"].(bool)
		if !ok || !isRefresh {
			return "", errors.New("invalid refresh token")
		}
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("invalid user_id in refresh token")
		}
		return userID, nil
	}

	return "", errors.New("invalid refresh token")
}

// RefreshToken — обновляет Access Token по Refresh Token
func (m *JWTManager) RefreshToken(refreshToken string) (string, error) {
	userID, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	return m.GenerateToken(userID)
}