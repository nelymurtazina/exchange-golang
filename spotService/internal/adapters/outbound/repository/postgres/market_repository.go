package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"test-project/spotService/internal/core/domain"
	ports "test-project/spotService/internal/core/ports/outbound"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type MarketRepository struct {
	db *pgxpool.Pool
}


func NewMarketRepository(db *pgxpool.Pool) ports.MarketRepository {
	return &MarketRepository{db: db}
}

func (r *MarketRepository) Create(ctx context.Context, market *domain.Market) error {
	priceJSON, err := json.Marshal(market.Price)
    if err != nil {
        return fmt.Errorf("failed to marshal price to JSON: %w", err)
    }
	query := `INSERT INTO markets (market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.db.Exec(ctx, query,
		market.MarketID,
		market.Name,
		market.BaseAsset,
		market.QuoteAsset,
		market.Enabled,
		priceJSON,
		market.CreatedAt,
		market.UpdatedAt,
	)
	return err
}

func (r *MarketRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE markets SET updated_at = NOW() WHERE market_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	rows := result.RowsAffected()
	if rows == 0 {
		return domain.ErrMarketNotFound
	}
	return nil
}

func (r *MarketRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
    FROM markets LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *MarketRepository) GetAllActive(ctx context.Context, limit, offset int) ([]*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at
			FROM markets WHERE enabled = true LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *MarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
              FROM markets WHERE market_id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return r.scanMarket(row)
}

func (r *MarketRepository) GetActiveByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at
              FROM markets WHERE market_id = $1 AND enabled = true`
	row := r.db.QueryRow(ctx, query, id)
	return r.scanMarket(row)
}

func (r *MarketRepository) Update(ctx context.Context, market *domain.Market) error {
	priceJSON, err := json.Marshal(market.Price)
	if err != nil {
        return fmt.Errorf("failed to marshal price to JSON: %w", err)
    }
	query := `UPDATE markets 
              SET name = $1, base_asset = $2, quote_asset = $3, enabled = $4, price = $5, updated_at = NOW()
              WHERE market_id = $6 `

	result, err := r.db.Exec(ctx, query,
		market.Name,
		market.BaseAsset,
		market.QuoteAsset,
		market.Enabled,
		priceJSON,
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
	var priceBytes []byte

	err := row.Scan(
		&market.MarketID,
		&market.Name,
		&market.BaseAsset,
		&market.QuoteAsset,
		&market.Enabled,
		&priceBytes,
		&market.CreatedAt,
		&market.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (r *MarketRepository) scanMarketRows(rows pgx.Rows) (*domain.Market, error) {
	var market domain.Market
	var priceBytes []byte

	err := rows.Scan(
		&market.MarketID,
		&market.Name,
		&market.BaseAsset,
		&market.QuoteAsset,
		&market.Enabled,
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


func (r *MarketRepository) GetAllActiveWithCursor(ctx context.Context, limit int, cursor string) ([]*domain.Market, error) {
	var query string
	var args []interface{}

	if cursor == "" {
		// Если курсора нет, берем первые limit записей
		query = `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
                 FROM markets 
                 WHERE enabled = true 
                 ORDER BY market_id ASC 
                 LIMIT $1`
		args = []interface{}{limit}
	} else {
		// Если курсор есть, берем записи, которые идут ПОСЛЕ него в алфавитном  порядке
		query = `SELECT market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at 
                 FROM markets 
                 WHERE enabled = true AND market_id > $1 
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
		market, err := r.scanMarketRows(rows)
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