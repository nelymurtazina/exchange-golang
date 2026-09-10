package domain

import (
	"errors"
	"regexp"
	commonv1 "test-project/api/gen/common"
	"time"

)

var (
	ErrMarketNotFound      = errors.New("market not found")
    ErrMarketDisabled      = errors.New("market is disabled")
    ErrMarketDeleted       = errors.New("market is deleted")
    ErrInvalidMarketID     = errors.New("invalid market_id")
    ErrInvalidMarketName   = errors.New("invalid market name")
    ErrInvalidAsset        = errors.New("invalid asset name")
    ErrInvalidPrice        = errors.New("invalid price")
	ValidUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

type Market struct {
	MarketID   string
	Name       string
	BaseAsset  string
	QuoteAsset string
	Enabled    bool
	Price *commonv1.Money
	CreatedAt  time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewMarket(marketID, name, baseAsset, quoteAsset string, price *commonv1.Money) (*Market, error){
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

	return &Market{
		MarketID: marketID,
		Name: name,
		BaseAsset: baseAsset,
		QuoteAsset: quoteAsset,
		Enabled: true,
		Price: price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}, nil
}

func ValidateMarketID(marketID string) error{
	if marketID == ""{
		return ErrInvalidMarketID
	}
	if !ValidUsernameRegex.MatchString(marketID){
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
	if !ValidUsernameRegex.MatchString(name){
		return ErrInvalidMarketName
	}
	return nil
}

func ValidateAsset(asset string) error {
    if asset == "" {
        return errors.New("asset cannot be empty")
    }
    matched, _ := regexp.MatchString(`^[A-Z]+$`, asset)
    if !matched {
        return errors.New("asset must be uppercase letters")
    }
    return nil
}

func ValidatePrice(price *commonv1.Money) error {
	if price == nil {
        return ErrInvalidPrice
    }
    if price.Amount == nil {
        return ErrInvalidPrice
    }
    if price.Amount.Units < 0 {
        return ErrInvalidPrice
    }
    return nil
}