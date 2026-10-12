package postgres

import (
	"context"
	"encoding/json"
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
	return &OrderRepository{
		db: db,
	}
}

func (o *OrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
	query := `SELECT order_id, user_id, market_id, order_type, price, quantity, status, created_at, updated_at 
              FROM orders WHERE order_id = $1`
	
	row := o.db.QueryRow(ctx, query, orderID)
	return o.scanOrder(row)
}

func (o *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	priceJSON, err := json.Marshal(order.Price)
	if err !=nil{
		return fmt.Errorf("failed to marshal price to JSON: %w", err)
	}
	query := `INSERT INTO orders (order_id, user_id, market_id, order_type, price, quantity, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = o.db.Exec(ctx, query,
		order.OrderID,
		order.UserID,
		order.MarketID,
		order.Type,
		priceJSON,
		order.Quantity,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}

func (r *OrderRepository) scanOrder(row pgx.Row) (*domain.Order, error) {
	var order domain.Order
	var priceBytes []byte

	err := row.Scan(
		&order.OrderID,
		&order.UserID,
		&order.MarketID,
		&order.Type,
		&priceBytes,       // Читаем JSONB как []byte
		&order.Quantity,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to scan order: %w", err)
	}

	if len(priceBytes) > 0 {
		if err := json.Unmarshal(priceBytes, &order.Price); err != nil {
			return nil, fmt.Errorf("failed to unmarshal price JSONB: %w", err)
		}
	}
	return &order, nil
}