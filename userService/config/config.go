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
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Password  PasswordConfig
	Services  ServicesConfig
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
	MaxOpenConns    int `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime 	 time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5"`
	ConnMaxIdleTime      time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"15"`
}


type JWTConfig struct {
	Secret       string
	ExpiresHours int
	AccessTTL  time.Duration 
	RefreshTTL time.Duration
}

type ServicesConfig struct {
	OrderServicePort      string
	InstrumentServicePort string
	UserServicePort       string
}

type MigrationConfig struct {
	Enabled bool
	Path    string
} //поменять 

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func LoadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	var cfg Config
	var err error

	// Database
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

	idleTime, err := getEnvAsInt("DB_CONN_MAX_IDLE_TIME")
	if err != nil {
		return cfg, err
	}
	cfg.Database.ConnMaxIdleTime = time.Duration(idleTime) * time.Minute

	cfg.Redis.Host, err = getEnv("REDIS_HOST")
	if err != nil {
		return cfg, err
	}
	cfg.Redis.Port, err = getEnvAsInt("REDIS_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.DB, err = getEnvAsInt("REDIS_DB")
	if err != nil {
		return cfg, err
	}

	// JWT
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

	// Services
	cfg.Services.OrderServicePort, err = getEnv("ORDER_SERVICE_PORT")
	if err != nil {
		return cfg, err
	}

	cfg.Services.InstrumentServicePort, err = getEnv("SPOT_SERVICE_PORT")
	if err != nil {
		return cfg, err
	}

	cfg.Services.UserServicePort, err = getEnv("USER_SERVICE_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	if cfg.JWT.Secret == "" {
		return cfg, err
	}
	accessTTLStr, err := getEnv("JWT_ACCESS_TTL")
	if err != nil { return cfg, err }
	cfg.JWT.AccessTTL, err = time.ParseDuration(accessTTLStr)
	if err != nil {
		return cfg, err
	}

	refreshTTLSt, err:= getEnv("JWT_REFRESH_TTL")
	if err != nil { return cfg, err }
	cfg.JWT.RefreshTTL, err = time.ParseDuration(refreshTTLSt)
	if err != nil {
		return cfg, err
	}

	// Migration
	cfg.Migration.Enabled = getEnvAsBool("MIGRATION_ENABLED", true)
	cfg.Migration.Path, err = getEnv("USER_MIGRATION_PATH")
	if err != nil {
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", fmt.Errorf("%s is required in .env", key)
	}
	if value := os.Getenv(key); value != "" {
		return value, nil
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
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.JWT.AccessTTL <= 0 {
		return errors.New("JWT_ACCESS_TTL must be > 0")
	}
	if c.JWT.RefreshTTL <= 0 {
		return errors.New("JWT_REFRESH_TTL must be > 0")
	}
	if c.JWT.RefreshTTL <= c.JWT.AccessTTL {
		return errors.New("JWT_REFRESH_TTL must be greater than JWT_ACCESS_TTL")
	}
	if c.Password.BcryptCost < 10 || c.Password.BcryptCost > 14 {
		return errors.New("BCRYPT_COST must be between 10 and 14")
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

type MigrateConfig struct {
	Database DatabaseConfig
	Path     string
}

func LoadMigrateConfig() (MigrateConfig, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	var cfg MigrateConfig
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
	cfg.Path, err = getEnv("USER_MIGRATION_PATH")
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
