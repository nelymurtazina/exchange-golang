package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTManager struct {
	secret       string
	expiresHours int
}

func NewJWTManager(secret string, expiresHours int) *JWTManager {
	return &JWTManager{
		secret:       secret,
		expiresHours: expiresHours,
	}
}

func (m *JWTManager) HashPassword(password string) (string, error) {
	// cost = 12 увеличить c 10!(от 12)")
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (m *JWTManager) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (m *JWTManager) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(m.expiresHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(m.expiresHours*24*7) * time.Hour).Unix(), // 7 дней
		"iat":     time.Now().Unix(),
		"refresh": true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

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
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("invalid user_id in token")
		}
		// Проверяем, не refresh ли это токен
		if isRefresh, ok := claims["refresh"].(bool); ok && isRefresh {
			return "", errors.New("refresh token cannot be used as access token")
		}
		return userID, nil
	}

	return "", errors.New("invalid token")
}

// ValidateRefreshToken проверяет refresh токен
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
			return "", errors.New("invalid user_id in token")
		}
		return userID, nil
	}

	return "", errors.New("invalid refresh token")
}

// RefreshToken обновляет токен
func (m *JWTManager) RefreshToken(oldToken string) (string, error) {
	userID, err := m.ValidateRefreshToken(oldToken)
	if err != nil {
		return "", err
	}
	return m.GenerateToken(userID)
}
