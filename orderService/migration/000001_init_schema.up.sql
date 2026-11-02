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
    order_type order_type_enum NOT NULL,
    price_units BIGINT NOT NULL,
    price_nanos INT NOT NULL,
    price_currency VARCHAR(3) NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    status order_status_enum NOT NULL DEFAULT 'CREATED',
    idempotency_key VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_orders_idempotency_key ON orders(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_market_id ON orders(market_id);
CREATE INDEX idx_orders_status ON orders(status);
