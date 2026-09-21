package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports"
	"time"

	"github.com/jackc/pgx/v5/pgconn" 
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
		user.UserID, user.UserName, user.Email, user.PasswordHash,
		user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.Message, "users_email_key") {
                return domain.ErrUserAlreadyExists
            }
            if strings.Contains(pgErr.Message, "users_username_key") {
                return domain.ErrUsernameAlreadyExists
            }
		}
		return fmt.Errorf("database error: %w", err) 
		
	}
	return nil
}

func (u *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET updated_at = $1 WHERE user_id = $2`
	now := time.Now()
	result, err := u.db.ExecContext(ctx, query, now, id)
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
		FROM users WHERE email = $1 
	`
	row := u.db.QueryRowContext(ctx, query, email)
	return u.scanUser(row)
}

func (u *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, created_at, updated_at
		FROM users WHERE user_id = $1 AND
	`
	row := u.db.QueryRowContext(ctx, query, id)
	return u.scanUser(row)
}

func (u *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password = $3, role = $4, updated_at = $5
		WHERE user_id = $6
	`
	_, err := u.db.ExecContext(ctx, query,
		user.UserName, user.Email, user.PasswordHash, user.Role, user.UpdatedAt, user.UserID,
	)
	return err
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.UserID,
		&user.UserName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
