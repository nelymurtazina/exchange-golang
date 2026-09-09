package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports"
	"time"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (user_id, username, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := u.db.ExecContext(ctx, query,
		user.UserID, user.UserName, user.Email, user.Password,
		user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("database error: %w", err) 
		
	}
	return nil
}

func (u *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = $1, updated_at = $2 WHERE user_id = $3 AND deleted_at IS NULL`
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

func (u *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`
	row := u.db.QueryRowContext(ctx, query, email)
	return u.scanUser(row)
}

func (u *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, created_at, updated_at
		FROM users WHERE user_id = $1 AND deleted_at IS NULL
	`
	row := u.db.QueryRowContext(ctx, query, id)
	return u.scanUser(row)
}

func (u *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password = $3, role = $4, updated_at = $5
		WHERE user_id = $6 AND deleted_at IS NULL
	`
	_, err := u.db.ExecContext(ctx, query,
		user.UserName, user.Email, user.Password, user.Role, user.UpdatedAt, user.UserID,
	)
	return err
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var user domain.User
	var deletedAt sql.NullTime 
	err := row.Scan(
		&user.UserID,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	if deletedAt.Valid {
        user.DeletedAt = &deletedAt.Time
    }

	return &user, nil
}
