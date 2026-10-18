package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidUsername       = errors.New("invalid username")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrUserDisabled          = errors.New("user is disabled")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrInvalidRole           = errors.New("invalid role")
)
