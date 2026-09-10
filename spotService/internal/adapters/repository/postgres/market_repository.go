package postgres

import (
    "context"
    "database/sql"
    "errors"

    "test-project/spotService/internal/core/domain"
    "test-project/spotService/internal/core/ports"
)

type MarketRepository struct {
    db *sql.DB
}

func NewMarketRepository(db *sql.DB) ports.MarketRepository {
    return &MarketRepository{db: db}
}

func (r *MarketRepository) Create(ctx context.Context, market *domain.Market) error {
    query := `INSERT INTO markets (market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
    
    _, err := r.db.ExecContext(ctx, query,
        market.MarketID,
        market.Name,
        market.BaseAsset,
        market.QuoteAsset,
        market.Enabled,
        market.Price,
        market.CreatedAt,
        market.UpdatedAt,
    )
    return err
}

func (r *MarketRepository) Delete(ctx context.Context, id string) error {
    query := `UPDATE markets SET deleted_at = NOW(), updated_at = NOW() WHERE market_id = $1 AND deleted_at IS NULL`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return domain.ErrMarketNotFound
    }
    return nil
}

func (r *MarketRepository) GetAll(ctx context.Context) ([]*domain.Market, error) {
    query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at, deleted_at FROM markets`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var markets []*domain.Market
    for rows.Next() {
        market, err := r.scanMarketRows(rows)
        if err != nil {
            return nil, err
        }
        markets = append(markets, market)
    }
    return markets, nil
}

func (r *MarketRepository) GetAllActive(ctx context.Context) ([]*domain.Market, error) {
    query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at, deleted_at 
              FROM markets WHERE enabled = true AND deleted_at IS NULL`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var markets []*domain.Market
    for rows.Next() {
        market, err := r.scanMarketRows(rows)
        if err != nil {
            return nil, err
        }
        markets = append(markets, market)
    }
    return markets, nil
}

func (r *MarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
    query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at, deleted_at 
              FROM markets WHERE market_id = $1`
    row := r.db.QueryRowContext(ctx, query, id)
    return r.scanMarket(row)
}

func (r *MarketRepository) GetActiveByID(ctx context.Context, id string) (*domain.Market, error) {
    query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at, deleted_at 
              FROM markets WHERE market_id = $1 AND enabled = true AND deleted_at IS NULL`
    row := r.db.QueryRowContext(ctx, query, id)
    return r.scanMarket(row)
}

func (r *MarketRepository) Update(ctx context.Context, market *domain.Market) error {
    query := `UPDATE markets 
              SET name = $1, base_asset = $2, quote_asset = $3, enabled = $4, price = $5, updated_at = NOW()
              WHERE market_id = $6 AND deleted_at IS NULL`
    
    result, err := r.db.ExecContext(ctx, query,
        market.Name,
        market.BaseAsset,
        market.QuoteAsset,
        market.Enabled,
        market.Price,
        market.MarketID,
    )
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return domain.ErrMarketNotFound
    }
    return nil
}

func (r *MarketRepository) scanMarket(row *sql.Row) (*domain.Market, error) {
    var market domain.Market
    var deletedAt sql.NullTime

    err := row.Scan(
        &market.MarketID,
        &market.Name,
        &market.BaseAsset,
        &market.QuoteAsset,
        &market.Enabled,
        &market.Price,
        &market.CreatedAt,
        &market.UpdatedAt,
        &deletedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domain.ErrMarketNotFound
        }
        return nil, err
    }
    if deletedAt.Valid {
        market.DeletedAt = &deletedAt.Time
    }
    return &market, nil
}

func (r *MarketRepository) scanMarketRows(rows *sql.Rows) (*domain.Market, error) {
    var market domain.Market
    var deletedAt sql.NullTime

    err := rows.Scan(
        &market.MarketID,
        &market.Name,
        &market.BaseAsset,
        &market.QuoteAsset,
        &market.Enabled,
        &market.Price,
        &market.CreatedAt,
        &market.UpdatedAt,
        &deletedAt,
    )
    if err != nil {
        return nil, err
    }
    if deletedAt.Valid {
        market.DeletedAt = &deletedAt.Time
    }
    return &market, nil
}