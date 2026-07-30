package repository

import (
	"context"
	"database/sql"
	"fmt"
	"grpc-exchange/services/usersService/config"
	"grpc-exchange/services/usersService/internal/core/domain"
	"grpc-exchange/services/usersService/internal/core/ports"
	
	_ "github.com/jackc/pgx/v5/stdlib"
)

type UserRepository struct {
	db *sql.DB
}

// CreateUser implements [ports.UserRepository].
func (u *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, email, password, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := u.db.ExecContext(ctx, query,
		user.UserID, user.Username,user.Email, user.Password,
		user.Role, user.Active, user.CreatedAt, user.UpdatedAt,
)
	return err
}

// Delete implements [ports.UserRepository].
func (u *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE user_id = $1`
	_, err := u.db.ExecContext(ctx, query, id)
	return err
}

// GetByEmail implements [ports.UserRepository].
func (u *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, active, created_at, updated_at
		FROM users WHERE email = $1
	`
	row := u.db.QueryRowContext(ctx, query, email)
	return u.scanUser(row)
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID implements [ports.UserRepository].
func (u *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, active, created_at, updated_at
		FROM users WHERE user_id = $1
	`
	row := u.db.QueryRowContext(ctx, query, id)
	return u.scanUser(row)
}

// Update implements [ports.UserRepository].
func (u *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password = $3, role = $4, active = $5, updated_at = $6
		WHERE user_id = $7
	`
	_, err := u.db.ExecContext(ctx, query,
		user.Username, user.Email, user.Password, user.Role, user.Active, user.UpdatedAt, user.UserID,
	)
	return err
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &UserRepository{db: db}
}





// NewConnection — подключение к БД (из конфига)
func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
