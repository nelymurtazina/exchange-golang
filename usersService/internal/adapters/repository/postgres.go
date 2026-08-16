package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"grpc-exchange/usersService/config"
	"grpc-exchange/usersService/internal/core/domain"
	"grpc-exchange/usersService/internal/core/ports"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &UserRepository{db: db}
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
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		//  обрыв сети или другая ошибка — возвращаем её
		return fmt.Errorf("database error: %w", err)// Если уникальный email нарушен
		//а если обрыв сети? ломает логику и вводит в заблуждение. Проверять код ошибки. 23550 КОД ОШИБКИ ПОЧИТАТЬ 
	}
	return nil
}

// Delete implements [ports.UserRepository].
func (u *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = $1, updated_at = $2 WHERE user_id = $3 AND deleted_at IS NULL`
	// по делиту таймстемп строго по этому удалять suft Deleted (без жесткого удаления)+
	now := time.Now()
	result, err := u.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// GetByEmail implements [ports.UserRepository].
func (u *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, active, created_at, updated_at
		FROM users WHERE email = $1
	`
	row := u.db.QueryRowContext(ctx, query, email)
	return u.scanUser(row)

	//игнорируем ошибку, добавить обработку ошибку 
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
	
	//антипатер, мне обработать каждую ошибку (исправила)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound 
		}
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
