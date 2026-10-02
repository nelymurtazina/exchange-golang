package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID   string
	UserName string
	Email    string
	PasswordHash string
	Role     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

var ValidUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

const (
    RoleUser  = "user"
    RoleAdmin = "admin"
    RoleGuest = "guest"
)

//подгрузить линтеры 
func NewUser(userID, userName, email, passwordHash, role string, createdAt, updatedAt time.Time) (*User, error){
	if err := ValidateUserName(userName); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if passwordHash == "" {
        return nil, ErrInvalidPassword
    }
	if err := ValidateID(userID); err != nil {
		return nil, err
	}
	if err := ValidateRole(role); err != nil {
        return nil, err
    }
	return &User{
		UserID: userID,
		UserName: userName,
		Email: email,
		PasswordHash : passwordHash,
		Role: role,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func ValidateID(id string) error{
	if id == ""{
		return ErrInvalidUserID
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidUserID
	}
	return nil
}


func ValidateUserName(name string) error{
	if name == ""{
		return ErrInvalidUsername
	}
	if len(name) < 2 || len(name) > 30{
		return ErrInvalidUsername
	}
	if !ValidUsernameRegex.MatchString(name) {
    	return ErrInvalidUsername
	}
	return nil
}

func ValidateEmail(email string) error{
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error{
	if len(password) < 8 || len(password) > 72 {
        return ErrInvalidPassword
    }
	return nil 
}

func NewUserID() string {
	return uuid.New().String()
}

func ValidateRole(role string) error {
    switch role {
    case RoleUser, RoleAdmin, RoleGuest:
        return nil
    default:
        return ErrInvalidRole
    }
}