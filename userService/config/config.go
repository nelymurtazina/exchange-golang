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
	Services ServicesConfig
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

type ServicesConfig struct {
	OrderServicePort      string
	InstrumentServicePort string
	UserServicePort       string
}

type MigrationConfig struct {
	Enabled bool
	Path    string
}

func LoadConfig() (Config, error){
	if err := godotenv.Load(); err != nil{
		log.Println("Warning: .env file not found, using environment variables")
	}

	var cfg Config
	var err error

	cfg.Database.Host = getEnv("DB_HOST")
	cfg.Database.Port, err = getEnvAsInt("DB_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.Database.User = getEnv("DB_USER")
	cfg.Database.Password = getEnv("DB_PASSWORD")
	cfg.Database.DBName = getEnv("DB_NAME")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE")

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

	// JWT
	cfg.JWT.Secret = getEnv("JWT_SECRET")
	cfg.JWT.ExpiresHours, err = getEnvAsInt("JWT_EXPIRES_HOURS")
	if err != nil {
		return cfg, err
	}

	// Services
	cfg.Services.OrderServicePort = getEnv("ORDER_SERVICE_PORT")
	cfg.Services.InstrumentServicePort = getEnv("INSTRUMENT_SERVICE_PORT")
	cfg.Services.UserServicePort = getEnv("USER_SERVICE_PORT")

	// Migration
	cfg.Migration.Enabled, err = getEnvAsBool("MIGRATION_ENABLED")
	if err != nil {
		// Если ошибка — по умолчанию true
		cfg.Migration.Enabled = true
	}
	cfg.Migration.Path = getEnv("MIGRATION_PATH")

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getEnv(key string) string {
    value, exists := os.LookupEnv(key)
    if !exists || value == "" {
        return ""  
    }
    return value
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
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.JWT.ExpiresHours <= 0 {
		return errors.New("JWT_EXPIRES_HOURS must be > 0")
	}
	if c.Services.UserServicePort == "" {
		return errors.New("USER_SERVICE_PORT is required")
	}
	if c.Migration.Enabled && c.Migration.Path == "" {
		return errors.New("MIGRATION_PATH is required when migrations are enabled")
	}
	return nil
}


func getEnvAsBool(key string) (bool, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return false, errors.New(key + " is required in .env")
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, errors.New(key + " must be true or false, got: " + value)
	}

	return boolValue, nil
}