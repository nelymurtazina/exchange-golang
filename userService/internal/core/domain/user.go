package domain

import (
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

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

const (
    RoleUser  = "user"
    RoleAdmin = "admin"
    RoleGuest = "guest"
)

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
	if len(name) < 3 || len(name) > 50 {
		return ErrInvalidUsername
	}
	if !usernameRegex.MatchString(name) {
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