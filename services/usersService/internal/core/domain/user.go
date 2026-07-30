package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type User struct {
	UserID    string
	Username  string
	Email     string
	Password  string
	Role      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrInvalidUserID      = errors.New("user id must be valid UUID")
	ErrInvalidUsername    = errors.New("username must be 3-20 characters, only letters and digits")
	ErrInvalidEmail       = errors.New("email must be valid email address")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserDisabled      = errors.New("user is disabled")
	ErrInvalidToken      = errors.New("invalid token")
)

func NewUser(id, name, email, password string) (*User, error){
	if err := ValidateUserName(name); err != nil{
		return nil, err
	}

	if err := ValidateEmail(email); err !=nil{
		return nil, err
	}
	if err := ValidateUserID(id); err != nil{
		return nil, err
	}
	

	return &User{
		UserID: id,
		Username: name,
		Email: email,
		Password: password,
		Role: "user",
		Active: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func ValidateUserName(name string) error {
	if len(name) < 2 || len(name) > 30{
		return ErrInvalidUsername
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", name)
	if !matched{
		return ErrInvalidUsername
	} 
	return nil
}

func ValidateEmail(email string) error {
	if email == ""{
		return ErrInvalidEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2{
		return ErrInvalidEmail
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return ErrInvalidEmail
	}

	if !strings.Contains(parts[1], "."){
		return ErrInvalidEmail
	}
	return nil
}


func ValidateUserID(id string) error{
	if id == ""{
		return ErrInvalidUserID
	}
	return nil
}
//здесь еще будут методы для изменения состояния с валидацией при update

