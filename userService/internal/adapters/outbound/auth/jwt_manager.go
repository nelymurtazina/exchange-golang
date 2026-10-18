package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *JWTManager) GenerateAccessToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(m.accessTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) GenerateRefreshToken(userID string, role string, jti string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(m.refreshTTL).Unix(),
		"iat":     time.Now().Unix(),
		"refresh": true,
        "jti": jti,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (string, string, error) {  
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(m.secret), nil
    })
    if err != nil {
        return "", "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        isRefresh, ok := claims["refresh"].(bool)
        if !ok || !isRefresh {
            return "", "", errors.New("invalid refresh token")
        }
        
        userID, ok := claims["user_id"].(string)
        if !ok {
            return "", "", errors.New("invalid user_id in refresh token")
        }
        
        jti, ok := claims["jti"].(string)
        if !ok || jti == "" {
            return "", "", errors.New("missing jti in refresh token")
        }
        return userID, jti, nil  
    }
    
    return "", "", errors.New("invalid refresh token")
}

func (m *JWTManager) GetAccessTTL() time.Duration {
	return m.accessTTL
}

func (m *JWTManager) GetRefreshTTL() time.Duration {
	return m.refreshTTL
}