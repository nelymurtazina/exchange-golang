package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis RedisConfig
	JWT      JWTConfig
	Password  PasswordConfig
	Services ServicesConfig
	Migration MigrationConfig
}
type PasswordConfig struct {
    BcryptCost int
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
type ServicesConfig struct {
	OrderServicePort      string
	InstrumentServicePort string
	UserServicePort       string
}

type MigrationConfig struct {
	Enabled bool
	Path    string
}

type RedisConfig struct{
	Host string
	Port int
}


func LoadConfig() (Config, error){
	if err := godotenv.Load(); err != nil{
		log.Println("Warning: .env file not found, using environment variables")
	}

	var cfg Config
	var err error

	cfg.Database.Host, err = getEnv("DB_HOST")
	if err != nil {
        return cfg, err
    }
	cfg.Database.Port, err = getEnvAsInt("DB_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.Database.User, err = getEnv("DB_USER")
	if err != nil {
		return cfg, err
	}
	cfg.Database.Password, err = getEnv("DB_PASSWORD")
	if err != nil {
		return cfg, err
	}
	cfg.Database.DBName, err = getEnv("DB_NAME")
	if err != nil {
		return cfg, err
	}
	cfg.Database.SSLMode, err = getEnv("DB_SSLMODE")
	if err != nil {
		return cfg, err
	}

	cfg.Database.MaxOpenConns, err = getEnvAsInt("DB_MAX_OPEN_CONNS")
	if err != nil {
		return cfg, err
	}
	cfg.Database.MaxIdleConns, err = getEnvAsInt("DB_MAX_IDLE_CONNS")
	if err != nil {
		return cfg, err
	}

	lifetime, err := getEnvAsInt("DB_CONN_MAX_LIFETIME")
	if err != nil {
		return cfg, err
	}
	cfg.Database.ConnMaxLifetime = time.Duration(lifetime) * time.Minute

	cfg.JWT.Secret, err = getEnv("JWT_SECRET")
    if err != nil {
        return cfg, err
    }

    cfg.JWT.ExpiresHours, err = getEnvAsInt("JWT_EXPIRES_HOURS")
    if err != nil {
        return cfg, err
    }

    cfg.Password.BcryptCost, err = getEnvAsInt("BCRYPT_COST")
    if err != nil {
        return cfg, err
    }

	cfg.Services.OrderServicePort, err = getEnv("ORDER_SERVICE_PORT")
	if err != nil {
        return cfg, fmt.Errorf("failed to get ORDER_SERVICE_PORT: %w", err)
    }
	cfg.Services.InstrumentServicePort, err = getEnv("INSTRUMENT_SERVICE_PORT")
	if err != nil {
        return cfg, fmt.Errorf("failed to get INSTRUMENT_SERVICE_PORT: %w", err)
    }
	cfg.Services.UserServicePort, err = getEnv("USER_SERVICE_PORT")
	if err != nil {
        return cfg, fmt.Errorf("failed to get USER_SERVICE_PORT: %w", err)
    }

	// Migration
	cfg.Migration.Enabled = getEnvAsBool("MIGRATION_ENABLED", true)
	cfg.Migration.Path, err = getEnv("SPOT_MIGRATION_PATH")
	if err != nil {
        return cfg, fmt.Errorf("SPOT_MIGRATION_PATH: %w", err)
    }

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	cfg.Redis.Host, err = getEnv("REDIS_HOST")
	if err != nil{
		return cfg, fmt.Errorf("failed to get REDIS_HOST: %w", err)
	}
	cfg.Redis.Port, err = getEnvAsInt("REDIS_PORT")
	if err != nil {
		return cfg, fmt.Errorf("failed to get REDIS_PORT: %w", err)
	}

	return cfg, nil

}

func getEnv(key string) (string, error) {
    value, exists := os.LookupEnv(key)
    if !exists || value == "" {
        return "", fmt.Errorf("%s is required in .env", key)
    }
    return value, nil
}

func getEnvAsInt(key string) (int, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return 0, errors.New(key + " is required in .env")
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(key + " must be a number, got: " + value)
	}

	return intValue, nil
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
	if c.Database.MaxOpenConns <= 0 {
		return errors.New("DB_MAX_OPEN_CONNS must be > 0")
	}
	if c.Database.MaxIdleConns <= 0 {
		return errors.New("DB_MAX_IDLE_CONNS must be > 0")
	}
	if c.Database.ConnMaxLifetime <= 0 {
		return errors.New("DB_CONN_MAX_LIFETIME must be > 0")
	}
	if c.Migration.Enabled && c.Migration.Path == "" {
		return errors.New("MIGRATION_PATH is required when migrations are enabled")
	}
	return nil
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("WARNING: invalid bool for %s: %q, using default %v", key, value, defaultValue)
		return defaultValue
	}
	return boolValue
}