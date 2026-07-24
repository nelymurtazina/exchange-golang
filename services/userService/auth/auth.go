package auth

import (
	"errors"
	"grpc-exchange/shered/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var cfg config.JWTConfig

// Init инициализирует конфигурацию JWT
func Init(config config.JWTConfig) {
    cfg = config
}

// HashPassword хэширует пароль
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// CheckPassword проверяет пароль
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// GenerateToken создает JWT токен
func GenerateToken(userID string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Duration(cfg.ExpiresHours) * time.Hour).Unix(),
        "iat":     time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken проверяет JWT токен и возвращает user_id
func ValidateToken(tokenString string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Проверяем алгоритм подписи
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(cfg.Secret), nil
    })

    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userID, ok := claims["user_id"].(string)
        if !ok {
            return "", errors.New("invalid user_id in token")
        }
        return userID, nil
    }

    return "", errors.New("invalid token")
}

// RefreshToken обновляет токен
func RefreshToken(oldToken string) (string, error) {
    userID, err := ValidateToken(oldToken)
    if err != nil {
        return "", err
    }
    return GenerateToken(userID)
}