CREATE TABLE IF NOT EXISTS markets (
    market_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    base_asset VARCHAR(50) NOT NULL,
    quote_asset VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX IF NOT EXISTS idx_markets_enabled ON markets(enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_markets_deleted_at ON markets(deleted_at);