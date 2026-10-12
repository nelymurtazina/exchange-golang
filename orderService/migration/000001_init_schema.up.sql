CREATE TYPE order_status_enum AS ENUM (
    'CREATED',
    'PENDING',
    'FILLED',
    'CANCELLED',
    'REJECTED'
);

CREATE TYPE order_type_enum AS ENUM (
    'BUY',
    'SELL'
);

CREATE TABLE IF NOT EXISTS orders (
    order_id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    market_id VARCHAR(100) NOT NULL,
    order_name VARCHAR(255), -- Добавлено, так как есть в твоем proto
    order_type order_type_enum NOT NULL,
    
    -- Цена хранится как JSONB, так как в proto это common.v1.Money (Units, Nanos, CurrencyCode)
    price JSONB NOT NULL,
    
    -- Количество. Используем NUMERIC для точности, хотя в proto double. 
    quantity DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    
    status order_status_enum NOT NULL DEFAULT 'CREATED',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_market_id ON orders(market_id);
CREATE INDEX idx_orders_status ON orders(status);