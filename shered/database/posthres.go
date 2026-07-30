package database

import (
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/jackc/pgx/v5/stdlib"
)

//мы пока не используем 

// Config — структура с настройками подключения
type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func NewPostgresConnection(config Config) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host,
        config.Port,
        config.User,
        config.Password,
        config.DBName,
        config.SSLMode,
    )
    
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    db.SetMaxOpenConns(25)     // Максимум открытых соединений. А если в сервисе потребуется больше? 
    //должно быть конфигурируемо в (.env)
    db.SetMaxIdleConns(5)      // Максимум простаивающих соединений(все в (.env))
    db.SetConnMaxLifetime(5 * time.Minute) // Время жизни соединения (все в (.env))
    
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return db, nil
}

