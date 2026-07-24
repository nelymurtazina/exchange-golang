package repository

import (
	"context"
	"database/sql"
	"fmt"

	"grpc-exchange/services/userService/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	CheckRole(ctx context.Context, userID, requiredRole string) (bool, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, email, password, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, active, created_at, updated_at
		FROM users WHERE user_id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanUser(row)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password, role, active, created_at, updated_at
		FROM users WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)
	return r.scanUser(row)
}

func (r *PostgresUserRepository) CheckRole(ctx context.Context, userID, requiredRole string) (bool, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.Role == requiredRole && user.Active, nil
}

func (r *PostgresUserRepository) scanUser(row *sql.Row) (*domain.User, error) {
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

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil 
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}