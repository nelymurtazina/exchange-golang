package interceptor

import "errors"

var (
	ErrMetadataMissing     = errors.New("metadata is not provided")
	ErrTokenMissing        = errors.New("authorization token is missing")
	ErrInvalidHeaderFormat = errors.New("invalid authorization header format")
	ErrTokenValidation     = errors.New("token validation failed") // Ошибка именно от валидатора
)