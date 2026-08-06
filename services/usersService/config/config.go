package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxOpenConns    int
	MaxIdleConns    int 
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret       string
	ExpiresHours int
}

type ServerConfig struct {
	Port         string
	// MetricsPort  string
}


func LoadConfig() Config {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	return Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "exchange"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 5)) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:       getEnv("JWT_SECRET", "my-super-secret-key"),
			ExpiresHours: getEnvAsInt("JWT_EXPIRES_HOURS", 24),
		},
		Server: ServerConfig{
			Port:        getEnv("USER_SERVICE_PORT", ":50053"),
		},
	}
}

// getEnv — получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt — получает переменную окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// проверяет обязательные поля
func (c Config) Validate() error {
    if c.Database.Host == "" {
        return errors.New("DB_HOST is required")
    }
    if c.Database.Port == 0 {
        return errors.New("DB_PORT is required")
    }
    if c.Database.DBName == "" {
        return errors.New("DB_NAME is required")
    }
    if c.JWT.Secret == "" {
        return errors.New("JWT_SECRET is required")
    }
    if c.JWT.ExpiresHours <= 0 {
        return errors.New("JWT_EXPIRES_HOURS must be > 0")
    }
    return nil
}