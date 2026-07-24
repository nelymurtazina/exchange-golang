package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Services ServicesConfig
	Metrics  MetricsConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
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

type MetricsConfig struct {
	OrderPort      string
	InstrumentPort string
	UserPort       string
}

func LoadConfig() Config {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "exchange"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
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
		Metrics: MetricsConfig{
			OrderPort:      getEnv("ORDER_METRICS_PORT", ":9091"),
			InstrumentPort: getEnv("INSTRUMENT_METRICS_PORT", ":9092"),
			UserPort:       getEnv("USER_METRICS_PORT", ":9093"),
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
    if value, exists := os.LookupEnv(key); exists {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}