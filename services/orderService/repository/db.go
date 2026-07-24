package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    
    "grpc-exchange/services/orderService/domain"
)

// OrderRepository — интерфейс для работы с заказами
type OrderRepository interface {
    Save(ctx context.Context, order *domain.Order) error
    GetByID(ctx context.Context, id string) (*domain.Order, error)
    GetByIDAndUser(ctx context.Context, orderID, userID string) (*domain.Order, error)
}

// PostgresOrderRepository — реализация репозитория для PostgreSQL
type PostgresOrderRepository struct {
    db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
    return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
    // Преобразуем Money и Decimal в JSON 
    priceJSON, err := json.Marshal(order.Price)
    if err != nil {
        return fmt.Errorf("failed to marshal price: %w", err)
    }
    
    quantityJSON, err := json.Marshal(order.Quantity)
    if err != nil {
        return fmt.Errorf("failed to marshal quantity: %w", err)
    }
    
    filledQuantityJSON, err := json.Marshal(order.FilledQuantity)
    if err != nil {
        return fmt.Errorf("failed to marshal filled_quantity: %w", err)
    }
    
    query := `
        INSERT INTO orders (
            order_id, user_id, market_id, side,
            price, quantity, filled_quantity, 
            status, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
    
    _, err = r.db.ExecContext(ctx, query,
        order.OrderID,
        order.UserID,
        order.MarketID,
        string(order.Side),
        priceJSON,
        quantityJSON,
        filledQuantityJSON,
        string(order.Status),
        order.CreatedAt,
        order.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    return nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    query := `
        SELECT 
            order_id, user_id, market_id, side,
            price, quantity, filled_quantity, 
            status, created_at, updated_at
        FROM orders WHERE order_id = $1
    `
    
    row := r.db.QueryRowContext(ctx, query, id)
    return r.scanOrder(row)
}

func (r *PostgresOrderRepository) GetByIDAndUser(ctx context.Context, orderID, userID string) (*domain.Order, error) {
    query := `
        SELECT 
            order_id, user_id, market_id, side,
            price, quantity, filled_quantity, 
            status, created_at, updated_at
        FROM orders WHERE order_id = $1 AND user_id = $2
    `
    
    row := r.db.QueryRowContext(ctx, query, orderID, userID)
    return r.scanOrder(row)
}

// преобразуем строку из БД в доменную сущность Order
func (r *PostgresOrderRepository) scanOrder(row *sql.Row) (*domain.Order, error) {
    var order domain.Order
    var side string
    var status string
    var priceJSON []byte
    var quantityJSON []byte
    var filledQuantityJSON []byte
    
    err := row.Scan(
        &order.OrderID,        
        &order.UserID,         
        &order.MarketID,       
        &side,                 
        &priceJSON,            
        &quantityJSON,         
        &filledQuantityJSON,   
        &status,               
        &order.CreatedAt,      
        &order.UpdatedAt,      
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil 
        }
        return nil, fmt.Errorf("failed to scan order: %w", err)
    }
    
    // Преобразуем JSON обратно в Money и Decimal
    if err := json.Unmarshal(priceJSON, &order.Price); err != nil {
        return nil, fmt.Errorf("failed to unmarshal price: %w", err)
    }
    if err := json.Unmarshal(quantityJSON, &order.Quantity); err != nil {
        return nil, fmt.Errorf("failed to unmarshal quantity: %w", err)
    }
    if err := json.Unmarshal(filledQuantityJSON, &order.FilledQuantity); err != nil {
        return nil, fmt.Errorf("failed to unmarshal filled_quantity: %w", err)
    }
    
    order.Side = domain.OrderSide(side)
    order.Status = domain.OrderStatus(status)
    
    return &order, nil
}