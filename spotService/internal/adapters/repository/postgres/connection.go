package postgres

import (
	"database/sql"
	"fmt"
	"test-project/spotService/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(cfg config.DatabaseConfig)(*sql.DB, error){
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
    cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil{
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}