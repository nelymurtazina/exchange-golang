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
	Database  DatabaseConfig
	JWT       JWTConfig
	Server    ServerConfig
	Migration MigrationConfig
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
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret       string
	ExpiresHours int
}

type ServerConfig struct {
	Port        string
	MetricsPort string
}

type MigrationConfig struct {
	Enabled bool
	Path    string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	return Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST"),
			Port:            getEnvAsInt("DB_PORT"),
			User:            getEnv("DB_USER"),
			Password:        getEnv("DB_PASSWORD"),
			DBName:          getEnv("DB_NAME"),
			SSLMode:         getEnv("DB_SSLMODE"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME")) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:       getEnv("JWT_SECRET"),
			ExpiresHours: getEnvAsInt("JWT_EXPIRES_HOURS"),
		},
		Server: ServerConfig{
			Port:        getEnv("USER_SERVICE_PORT"),
			MetricsPort: getEnv("USER_METRICS_PORT"),
		},
		Migration: MigrationConfig{
			Enabled: getEnvAsBool("MIGRATION_ENABLED"),
			Path:    getEnv("MIGRATION_PATH"),
		},
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("WARNING: %s is not set in .env", key)
		return ""
	}
	return value
}

func getEnvAsInt(key string) int {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("WARNING: %s is not set in .env", key)
		return 0
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("WARNING: invalid integer for %s: %q", key, value)
		return 0
	}

	return intValue
}

func getEnvAsBool(key string) bool {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("WARNING: %s is not set in .env", key)
		return false
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("WARNING: invalid bool for %s: %q", key, value)
		return false
	}

	return boolValue
}

func (c Config) Validate() error {
	if c.Database.Host == "" {
		return errors.New("DB_HOST is required in .env")
	}
	if c.Database.Port == 0 {
		return errors.New("DB_PORT is required in .env")
	}
	if c.Database.DBName == "" {
		return errors.New("DB_NAME is required in .env")
	}
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET is required in .env")
	}
	if c.JWT.ExpiresHours <= 0 {
		return errors.New("JWT_EXPIRES_HOURS must be > 0 in .env")
	}
	if c.Server.Port == "" {
		return errors.New("USER_SERVICE_PORT is required in .env")
	}
	if c.Migration.Enabled && c.Migration.Path == "" {
		return errors.New("MIGRATION_PATH is required in .env when migrations are enabled")
	}
	return nil
}