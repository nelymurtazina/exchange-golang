package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"test-project/spotService/internal/core/domain"
	ports "test-project/spotService/internal/core/ports/outbound"
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
	query := `UPDATE markets SET updated_at = NOW() WHERE market_id = $1`
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

func (r *MarketRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
    FROM markets LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close rows: %w", closeErr)
			} else {
				log.Printf("failed to close rows after error: %v", closeErr)
			}
		}
	}()

	var markets []*domain.Market
	for rows.Next() {
		market, err := r.scanMarketRows(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *MarketRepository) GetAllActive(ctx context.Context, limit, offset int) ([]*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at
              FROM markets LIMIT $1 OFFSET $2 WHERE enabled = true`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var markets []*domain.Market
	for rows.Next() {
		market, err := r.scanMarketRows(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *MarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
              FROM markets WHERE market_id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanMarket(row)
}

func (r *MarketRepository) GetActiveByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at
              FROM markets WHERE market_id = $1 AND enabled = true`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanMarket(row)
}

func (r *MarketRepository) Update(ctx context.Context, market *domain.Market) error {
	query := `UPDATE markets 
              SET name = $1, base_asset = $2, quote_asset = $3, enabled = $4, price = $5, updated_at = NOW()
              WHERE market_id = $6 `

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
	var priceBytes []byte

	err := row.Scan(
		&market.MarketID,
		&market.Name,
		&market.BaseAsset,
		&market.QuoteAsset,
		&market.Enabled,
		&priceBytes,
		&market.Price,
		&market.CreatedAt,
		&market.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMarketNotFound
		}
		return nil, err
	}
	if len(priceBytes) > 0 {
		if err := json.Unmarshal(priceBytes, &market.Price); err != nil {
			return nil, fmt.Errorf("failed to unmarshal price JSONB in scanMarket: %w", err)
		}
	}
	return &market, nil
}

func (r *MarketRepository) scanMarketRows(rows *sql.Rows) (*domain.Market, error) {
	var market domain.Market
	var priceBytes []byte

	err := rows.Scan(
		&market.MarketID,
		&market.Name,
		&market.BaseAsset,
		&market.QuoteAsset,
		&market.Enabled,
		&market.Price,
		&priceBytes,
		&market.CreatedAt,
		&market.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(priceBytes) > 0 {
		err := json.Unmarshal(priceBytes, &market.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSONB metadata: %w", err)
		}
	}
	return &market, nil
}


func (r *MarketRepository) GetAllActiveWithCursor(ctx context.Context, limit int, cursor string) (markets []*domain.Market, err error) {
	var query string
	var rows *sql.Rows

	if cursor == "" {
		query = `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
                 FROM markets 
                 WHERE enabled = true 
                 ORDER BY market_id ASC 
                 LIMIT $1`
		rows, err = r.db.QueryContext(ctx, query, limit)
	} else {
		query = `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
                 FROM markets 
                 WHERE enabled = true AND market_id > $1 
                 ORDER BY market_id ASC 
                 LIMIT $2`
		rows, err = r.db.QueryContext(ctx, query, cursor, limit)
	}

	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close rows: %w", closeErr)
			} else {
				log.Printf("failed to close rows after error: %v", closeErr)
			}
		}
	}()

	for rows.Next() {
		market, scanErr := r.scanMarketRows(rows) 
		if scanErr != nil {
			return nil, scanErr
		}
		markets = append(markets, market)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return markets, nil
}
