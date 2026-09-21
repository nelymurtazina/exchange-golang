package domain

import (
	"errors"
	"regexp"
	"time"

)

var (
	ErrMarketNotFound      = errors.New("market not found")
    ErrMarketDisabled      = errors.New("market is disabled")
    ErrInvalidMarketID     = errors.New("invalid market_id")
    ErrInvalidMarketName   = errors.New("invalid market name")
    ErrInvalidAsset        = errors.New("invalid asset name")
    ErrInvalidPrice        = errors.New("invalid price")
	ErrMarketAlreadyExists = errors.New("market already exists")
	ValidUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	assetRegex = regexp.MustCompile(`^[A-Z]+$`)
	RoleUser = "ROLE_USER"
	RoleGuest = "ROLE_GUEST"
)

type Money struct {
    Units        int64
    Nanos        int32
    CurrencyCode string
}

type Market struct {
	MarketID   string
	Name       string
	BaseAsset  string
	QuoteAsset string
	Enabled    bool
	Price Money  
	CreatedAt  time.Time
	UpdatedAt time.Time
}

func NewMarket(marketID, name, baseAsset, quoteAsset string, price Money, createdAt, updatedAt time.Time) (*Market, error){
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
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
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
        return ErrInvalidAsset
    }
    matched, _ := regexp.MatchString(`^[A-Z]+$`, asset)
    if !matched {
        return errors.New("asset must be uppercase letters")
    }
	if !assetRegex.MatchString(asset) {
        return ErrInvalidAsset 
    }
    return nil
}

func ValidatePrice(price Money) error {
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