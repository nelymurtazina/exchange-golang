package domain

import "errors"

var (
	ErrMarketNotFound      = errors.New("market not found")
	ErrMarketDisabled      = errors.New("market is disabled")
	ErrInvalidMarketID     = errors.New("invalid market_id")
	ErrInvalidMarketName   = errors.New("invalid market name")
	ErrInvalidAsset        = errors.New("invalid asset name")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrMarketAlreadyExists = errors.New("market already exists")
	ErrPermissionDenied    = errors.New("you need to log in.")
)