-- Создаем таблицу с плоскими колонками для цены (без JSONB!)
CREATE TABLE IF NOT EXISTS markets (
    market_id       VARCHAR(255) PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    base_asset      VARCHAR(50)  NOT NULL,
    quote_asset     VARCHAR(50)  NOT NULL,
    enabled         BOOLEAN      NOT NULL DEFAULT true,
    
    -- Плоские колонки для Money (быстро и надежно)
    price_units     BIGINT       NOT NULL DEFAULT 0,
    price_nanos     INT          NOT NULL DEFAULT 0,
    price_currency  VARCHAR(10)  NOT NULL,
    
    -- Минимальное количество в наименьших единицах (избегаем float)
    min_quantity    BIGINT       NOT NULL DEFAULT 1 CHECK (min_quantity > 0),
    
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ  NULL -- Soft Delete
);

-- Частичный уникальный индекс: одна активная пара (base, quote) не может дублироваться
CREATE UNIQUE INDEX IF NOT EXISTS unique_market_pair_active
    ON markets (base_asset, quote_asset)
    WHERE deleted_at IS NULL;

-- Частичный индекс для быстрых запросов только активных рынков
CREATE INDEX IF NOT EXISTS idx_markets_enabled_active
    ON markets (enabled)
    WHERE deleted_at IS NULL AND enabled = true;

-- Индекс для курсорной пагинации (по market_id для активных)
CREATE INDEX IF NOT EXISTS idx_markets_cursor
    ON markets (market_id)
    WHERE deleted_at IS NULL AND enabled = true;