package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"test-project/spotService/internal/core/domain"
	ports "test-project/spotService/internal/core/ports/outbound"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const marketSelectColumns = `market_id, name, base_asset, quote_asset, enabled, price_units, price_nanos, price_currency, min_quantity, created_at, updated_at, deleted_at`

type MarketRepository struct {
	db *pgxpool.Pool
}

func NewMarketRepository(db *pgxpool.Pool) ports.MarketRepository {
	return &MarketRepository{db: db}
}

func (r *MarketRepository) Create(ctx context.Context, market *domain.Market) error {
	query := `INSERT INTO markets (market_id, name, base_asset, quote_asset, enabled, price_units, price_nanos, price_currency, min_quantity, created_at, updated_at, deleted_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(ctx, query,
		market.MarketID,
		market.Name,
		market.BaseAsset,
		market.QuoteAsset,
		market.Enabled,
		market.Price.Units,
		market.Price.Nanos,
		market.Price.CurrencyCode,
		market.MinQuantity,
		market.CreatedAt,
		market.UpdatedAt,
		market.DeletedAt,
	)
	return err
}

func (r *MarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT ` + marketSelectColumns + `
              FROM markets WHERE market_id = $1 AND deleted_at IS NULL`
	row := r.db.QueryRow(ctx, query, id)
	return r.scanMarket(row)
}

func (r *MarketRepository) Update(ctx context.Context, market *domain.Market) error {
	query := `UPDATE markets
              SET name = $1, base_asset = $2, quote_asset = $3, enabled = $4,
                  price_units = $5, price_nanos = $6, price_currency = $7,
                  min_quantity = $8, updated_at = NOW(), deleted_at = $9
              WHERE market_id = $10 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		market.Name,
		market.BaseAsset,
		market.QuoteAsset,
		market.Enabled,
		market.Price.Units,
		market.Price.Nanos,
		market.Price.CurrencyCode,
		market.MinQuantity,
		market.DeletedAt,
		market.MarketID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMarketNotFound
	}
	return nil
}

func (r *MarketRepository) scanMarket(row pgx.Row) (*domain.Market, error) {
	var market domain.Market
	var deletedAt *time.Time 

	err := row.Scan(
		&market.MarketID,
		&market.Name,
		&market.BaseAsset,
		&market.QuoteAsset,
		&market.Enabled,
		&market.Price.Units,
		&market.Price.Nanos,
		&market.Price.CurrencyCode,
		&market.MinQuantity,
		&market.CreatedAt,
		&market.UpdatedAt,
		&deletedAt, 
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMarketNotFound
		}
		return nil, err
	}

	market.DeletedAt = deletedAt

	return &market, nil
}

func (r *MarketRepository) GetAllActiveWithCursor(ctx context.Context, limit int, cursor string) ([]*domain.Market, error) {
	var query string
	var args []interface{}

	if cursor == "" {
		query = `SELECT ` + marketSelectColumns + `
                 FROM markets
                 WHERE enabled = true AND deleted_at IS NULL
                 ORDER BY market_id ASC
                 LIMIT $1`
		args = []interface{}{limit}
	} else {
		query = `SELECT ` + marketSelectColumns + `
                 FROM markets
                 WHERE enabled = true AND deleted_at IS NULL AND market_id > $1
                 ORDER BY market_id ASC
                 LIMIT $2`
		args = []interface{}{cursor, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query markets: %w", err)
	}
	defer rows.Close()

	var markets []*domain.Market
	for rows.Next() {
		market, err := r.scanMarket(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return markets, nil
}

func (r *MarketRepository) SoftDelete(ctx context.Context, id string) error {
	query := `
		UPDATE markets
		SET deleted_at = NOW(), updated_at = NOW(), enabled = false
		WHERE market_id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete market: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMarketNotFound
	}
	return nil
}