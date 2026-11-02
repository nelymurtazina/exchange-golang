package domain

import (
	"regexp"
	"test-project/shared/money"
	"time"
)

var (
	marketIDRegex = regexp.MustCompile(`^[A-Z]+-[A-Z]+$`)
	marketNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
	assetRegex    = regexp.MustCompile(`^[A-Z]+$`)
	RoleUser  = "user"
	RoleAdmin = "admin"
	RoleGuest = "guest"
)


type Market struct {
	MarketID    string
	Name        string
	BaseAsset   string
	QuoteAsset  string
	Enabled     bool
	Price       money.Money
	MinQuantity int64 
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time 
}

func NewMarket(marketID, name, baseAsset, quoteAsset string, price money.Money, minQuantity int64, createdAt, updatedAt time.Time) (*Market, error){
	if err := ValidateMarketID(marketID); err != nil{
		return nil, err
	}
	if err := ValidateMarketName(name); err != nil{
		return nil, err
	}
	if err := ValidateAsset(baseAsset); err != nil {
        return nil, err
    }
    if err := ValidateAsset(quoteAsset); err != nil {
        return nil, err
    }
    if err := ValidatePrice(price); err != nil {
        return nil, err
    }
	if minQuantity <= 0 {
		return nil, ErrInvalidPrice 
	}

	return &Market{
		MarketID:    marketID,
		Name:        name,
		BaseAsset:   baseAsset,
		QuoteAsset:  quoteAsset,
		Enabled:     true,
		Price:       price,
		MinQuantity: minQuantity,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   nil,
	}, nil
}

func (m *Market) IsDeleted() bool {
	return m.DeletedAt != nil
}

func (m *Market) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
	m.Enabled = false
}

func ValidateMarketID(marketID string) error{
	if marketID == ""{
		return ErrInvalidMarketID
	}
	if !marketIDRegex.MatchString(marketID) {
		return ErrInvalidMarketID
	}
	return nil
}

func ValidateMarketName(name string) error{
	if name == ""{
		return ErrInvalidMarketName
	}
	if len(name) < 1 || len(name) > 40{
		return ErrInvalidMarketName
	}
	if !marketNameRegex.MatchString(name) {
		return ErrInvalidMarketName
	}
	return nil
}

func ValidateAsset(asset string) error {
    if asset == "" {
        return ErrInvalidAsset
    }
	if !assetRegex.MatchString(asset) {
        return ErrInvalidAsset 
    }
    return nil
}

func ValidatePrice(price money.Money) error {
    if price.CurrencyCode == "" {
        return ErrInvalidPrice
    }
    if price.Nanos < 0 || price.Nanos > 999999999 {
        return ErrInvalidPrice
    }
    if price.Units < 0 || (price.Units == 0 && price.Nanos == 0) {
        return ErrInvalidPrice
    }
    return nil
}