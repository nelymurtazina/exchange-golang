package domain

import (
	"time"
)

type Market struct {
	MarketID    string
	Name        string
	BaseAsset   string
	QuoteAsset  string
	Enabled     bool
	CreatedAt   time.Time
	DeletedAt   *time.Time
}