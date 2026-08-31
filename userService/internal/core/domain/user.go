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
	Password string
	Role     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var(
	ErrUserNotFound = errors.New("Пользователь не найден в БД")
	ErrUserAlreadyExists = errors.New("Email уже зарегистрирован")
	ErrInvalidUsername = errors.New("Невалидный username")
	ErrInvalidEmail = errors.New("Невалидный email")
	ErrInvalidPassword = errors.New("Пароль пустой или слишком короткий")
	ErrInvalidUserID = errors.New("Невалидный UUID")
	ErrUserDisabled = errors.New("Пользователь заблокирован")
	ErrInvalidCredentials = errors.New("Неверные учётные данные")
	ErrInvalidToken = errors.New("Невалидный JWT токен")
	
	ValidUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	RoleUser = "ROLE_USER"
)

func NewUser(userID, userName, email, password, role string) (*User, error){
	if err := ValidateUserName(userName); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if password == "" {
		return nil, ErrInvalidPassword
	}
	if err := ValidateID(userID); err != nil {
		return nil, err
	}
	return &User{
		UserID: userID,
		UserName: userName,
		Email: email,
		Password: password,
		Role: RoleUser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
	if password == "" && len(password)>8{
		return ErrInvalidPassword
	}
	return nil
}

func NewUserID() string {
	return uuid.New().String()
}
