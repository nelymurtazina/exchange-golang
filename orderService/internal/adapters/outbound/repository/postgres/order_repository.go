package postgres

import (
	"context"
	"errors"
	"fmt"

	"test-project/orderService/internal/core/domain"
	"test-project/orderService/internal/core/ports/outbound"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) outbound.OrderRepository {
	return &OrderRepository{db: db}
}

func (o *OrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
	query := `SELECT 
		order_id, user_id, market_id, order_type, price_units, price_nanos, price_currency, quantity, status, idempotency_key, created_at, updated_at 
	FROM orders WHERE order_id = $1`
	
	row := o.db.QueryRow(ctx, query, orderID)
	return o.scanOrder(row)
}

func (o *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	query := `SELECT 
		order_id, user_id, market_id, order_type, price_units, price_nanos, price_currency, quantity, status, idempotency_key, created_at, updated_at 
	FROM orders WHERE idempotency_key = $1`
	
	row := o.db.QueryRow(ctx, query, key)
	return o.scanOrder(row)
}

func (o *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (
		order_id, user_id, market_id, order_type, price_units, price_nanos, price_currency, quantity, status, idempotency_key, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := o.db.Exec(ctx, query,
		order.OrderID,
		order.UserID,
		order.MarketID,
		order.Type,
		order.Price.Units,        
		order.Price.Nanos,       
		order.Price.CurrencyCode, 
		order.Quantity,           
		order.Status,
		order.IdempotencyKey,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}

func (o *OrderRepository) scanOrder(row pgx.Row) (*domain.Order, error) {
	var order domain.Order

	err := row.Scan(
		&order.OrderID,
		&order.UserID,
		&order.MarketID,
		&order.Type,
		&order.Price.Units,        
		&order.Price.Nanos,       
		&order.Price.CurrencyCode,
		&order.Quantity,
		&order.Status,
		&order.IdempotencyKey,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to scan order: %w", err)
	}

	return &order, nil
}