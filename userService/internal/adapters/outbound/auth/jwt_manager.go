package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//сделать интерфейс для JWTManager 

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

func (m *JWTManager) GenerateAccessToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Duration(m.expiresHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) GenerateRefreshToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role": role,
		"exp":     time.Now().Add(time.Duration(m.expiresHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"refresh": true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (string, string, error) {  
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
        if isRefresh, ok := claims["refresh"].(bool); ok && isRefresh {
            return "", "", errors.New("refresh token cannot be used as access token")
        }
        userID, ok := claims["user_id"].(string)
        if !ok {
            return "", "", errors.New("invalid user_id in token")
        }
        role, _ := claims["role"].(string) 
        return userID, role, nil 
    }
    return "", "", errors.New("invalid token")
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (string, string, error) {  // ← + role
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
        role, _ := claims["role"].(string) 
        return userID, role, nil  
    }
    return "", "", errors.New("invalid refresh token")
}

func (m *JWTManager) RefreshToken(refreshToken string) (string, error) {
    userID, role, err := m.ValidateRefreshToken(refreshToken) 
    if err != nil {
        return "", err
    }
    return m.GenerateAccessToken(userID, role)  
    //сразу менять оба, а не только аксес
}
