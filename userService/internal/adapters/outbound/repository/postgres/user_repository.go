package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports/outbound"
	"time"

	"github.com/jackc/pgx/v5/pgconn" 
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) ports.UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (user_id, username, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := u.db.Exec(ctx, query,
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
	query := `UPDATE users SET updated_at = NOW() WHERE user_id = $1`
	result, err := u.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	rows := result.RowsAffected()
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
	row := u.db.QueryRow(ctx, query, email)
	return u.scanUser(row)
}

func (u *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, created_at, updated_at
		FROM users WHERE user_id = $1 
	`
	//когда поменяю контракт изменится логика сервиса 
	row := u.db.QueryRow(ctx, query, id)
	return u.scanUser(row)
}

func (u *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users 
        SET password = $1, updated_at = NOW()
        WHERE user_id = $2
	`
	result, err := u.db.Exec(ctx, query, user.PasswordHash, user.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	} 
	return nil
}

func (u *UserRepository) scanUser(row pgx.Row) (*domain.User, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}


func (u *UserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error{
	query := `UPDATE users 
        SET password = $1, updated_at = NOW()
        WHERE user_id = $2`
	now := time.Now()
	result, err := u.db.Exec(ctx, query,passwordHash,now,userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	} 
	return nil
}