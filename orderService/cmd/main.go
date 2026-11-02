package main

import (
	"log"

	"test-project/orderService/config"
	"test-project/orderService/internal/app"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.New(&cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}
