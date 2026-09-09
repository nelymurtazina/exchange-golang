package postgres

import (
	"context"
	"database/sql"
	"test-project/spotService/internal/core/domain"
	"test-project/spotService/internal/core/ports"
)

type MarketRepository struct {
	db *sql.DB
}

func NewSpotRepository(db *sql.DB) ports.MarketRepository {
	return &MarketRepository{db: db}
}


func (m *MarketRepository) Create(ctx context.Context, market *domain.Market) error {
	query := `INSERT INTO markets (market_id, name, base_asset, quote_asset, enabled, price, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
}

func (m *MarketRepository) Delete(ctx context.Context, id string) error {
    query := `UPDATE markets SET deleted_at = NOW(), updated_at = NOW() WHERE market_id = $1`
    _, err := m.db.ExecContext(ctx, query, id)
    return err
}
func (m *MarketRepository) GetAll(ctx context.Context) ([]*domain.Market, error) {
	query := `SELECT * FROM markets`
}

func (m *MarketRepository) GetAllActive(ctx context.Context) ([]*domain.Market, error) {
	query := `SELECT * FROM markets WHERE enabled = true AND deleted_at IS NULL`
}

func (m *MarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT * FROM markets WHERE market_id = $1`
}

func (m *MarketRepository) Update(ctx context.Context, market *domain.Market) error {
	query := `UPDATE markets`
}

func (m *MarketRepository) scanMarket(row *sql.Row) (*domain.Market, error){
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
		&market.DeletedAt,
	)
	if err != nil {

	}
}