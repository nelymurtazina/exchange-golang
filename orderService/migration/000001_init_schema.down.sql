DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_market_id;
DROP INDEX IF EXISTS idx_orders_user_id;

DROP TABLE IF EXISTS orders;

DROP TYPE IF EXISTS order_type_enum;
DROP TYPE IF EXISTS order_status_enum;