package main

import (
	"fmt"
	"log"
	"test-project/userService/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	m, err := migrate.New(
			"file://"+cfg.Migration.Path,
			dsn,
		)
		if err != nil {
			logger.Fatal("failed to create migrate instance", zap.Error(err))
		}
	
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Fatal("failed to apply migrations", zap.Error(err))
		}

		logger.Info("Migrations applied successfully")
}