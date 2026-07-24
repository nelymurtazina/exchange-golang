package repository

import (
    "context"
    "database/sql"
    "fmt"
    "grpc-exchange/services/spotInstrumentService/domain"
)

type MarketRepository interface {
    GetAll(ctx context.Context) ([]*domain.Market, error)
    GetByID(ctx context.Context, id string) (*domain.Market, error)
    GetAllActive(ctx context.Context) ([]*domain.Market, error)
    GetActiveByID(ctx context.Context, id string) (*domain.Market, error)
}

type PostgresMarketRepository struct {
    db *sql.DB
}

func NewPostgresMarketRepository(db *sql.DB) *PostgresMarketRepository {
    return &PostgresMarketRepository{db: db}
}

func (r *PostgresMarketRepository) GetAll(ctx context.Context) ([]*domain.Market, error) {
    query := `
        SELECT 
            market_id, name, base_asset, quote_asset, enabled,
            created_at, deleted_at
        FROM markets
    `
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query markets: %w", err)
    }
    defer rows.Close()
    
    return r.scanMarkets(rows)
}

func (r *PostgresMarketRepository) GetAllActive(ctx context.Context) ([]*domain.Market, error) {
    query := `
        SELECT 
            market_id, name, base_asset, quote_asset, enabled,
            created_at, deleted_at
        FROM markets
        WHERE enabled = true AND deleted_at IS NULL
    `
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query active markets: %w", err)
    }
    defer rows.Close()
    
    return r.scanMarkets(rows)
}

func (r *PostgresMarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
    query := `
        SELECT 
            market_id, name, base_asset, quote_asset, enabled,
            created_at, deleted_at
        FROM markets WHERE market_id = $1
    `
    
    row := r.db.QueryRowContext(ctx, query, id)
    return r.scanMarket(row)
}

func (r *PostgresMarketRepository) GetActiveByID(ctx context.Context, id string) (*domain.Market, error) {
    query := `
        SELECT 
            market_id, name, base_asset, quote_asset, enabled,
            created_at, deleted_at
        FROM markets 
        WHERE market_id = $1 AND enabled = true AND deleted_at IS NULL
    `
    
    row := r.db.QueryRowContext(ctx, query, id)
    return r.scanMarket(row)
}

func (r *PostgresMarketRepository) scanMarket(row *sql.Row) (*domain.Market, error) {
    var market domain.Market
    var deletedAt sql.NullTime
    
    err := row.Scan(
        &market.MarketID,
        &market.Name,
        &market.BaseAsset,
        &market.QuoteAsset,
        &market.Enabled,
        &market.CreatedAt,
        &deletedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil 
        }
        return nil, fmt.Errorf("failed to scan market: %w", err)
    }
    
    if deletedAt.Valid {
        market.DeletedAt = &deletedAt.Time
    }
    
    return &market, nil
}

func (r *PostgresMarketRepository) scanMarkets(rows *sql.Rows) ([]*domain.Market, error) {
    var markets []*domain.Market
    
    for rows.Next() {
        var market domain.Market
        var deletedAt sql.NullTime
        
        err := rows.Scan(
            &market.MarketID,
            &market.Name,
            &market.BaseAsset,
            &market.QuoteAsset,
            &market.Enabled,
            &market.CreatedAt,
            &deletedAt,
        )
        
        if err != nil {
            return nil, fmt.Errorf("failed to scan market: %w", err)
        }
        
        if deletedAt.Valid {
            market.DeletedAt = &deletedAt.Time
        }
        
        markets = append(markets, &market)
    }
    
    return markets, nil
}