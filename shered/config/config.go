package config

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Services ServicesConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type JWTConfig struct {
	Secret       string
	ExpiresHours int
}

type ServicesConfig struct {
	OrderServicePort      string
	InstrumentServicePort string
	UserServicePort       string
}

// type MetricsConfig struct {
// 	OrderPort      string
// 	InstrumentPort string
// 	UserPort       string
// }

func LoadConfig() Config {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "exchange"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 5),
		},
		JWT: JWTConfig{
			Secret:       getEnv("JWT_SECRET", "my-super-secret-key"),
			ExpiresHours: getEnvAsInt("JWT_EXPIRES_HOURS", 24),
		},
		Services: ServicesConfig{
			OrderServicePort:      getEnv("ORDER_SERVICE_PORT", ":50052"),
			InstrumentServicePort: getEnv("INSTRUMENT_SERVICE_PORT", ":50051"),
			UserServicePort:       getEnv("USER_SERVICE_PORT", ":50053"),
		},
	}
}

// getEnv получаем переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		// Логируем ошибку и возвращаем значение по умолчанию
		log.Printf("WARNING: invalid integer value for %s: %q, using default %d", key, value, defaultValue)
		return defaultValue
	}

	return intValue
}

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
