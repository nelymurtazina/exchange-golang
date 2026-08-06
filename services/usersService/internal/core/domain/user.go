package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidUsername   = errors.New("invalid username")
	ErrUserDisabled      = errors.New("user is disabled")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidUserID     = errors.New("invalid user_id")
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
	if name == "" {
		return ErrInvalidUsername
	}
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
	
	//net.mail, uuid проверки добавить.
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	return nil
}


func ValidateUserID(id string) error{
	if id == ""{
		return ErrInvalidUserID
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidUserID
	}
	return nil
}

func NewUserID() string {
	return uuid.New().String()
}
//здесь еще будут методы для изменения состояния с валидацией при update



//реализация кеширования, не будет методов jwtMeneger. ТОлько работа с jwt. Добавить input Output(для решистрации). Более сложна логика для email.Обработка ошибок без nil. а реальные ошибки, если уникальная ошибка, то конвертирую в domain и прокидываю ее(чтобы сервер отличил ошибку)
//refreshToken изменить , добавить в порты. 
//сервис логина, даже если пользователь не найдет, нужно защитить систему от атак. + добавить миграцию бд. (без миграции не валиден). Закрывать после грейсшатдаун. есть готовые паттерны для грейсоушшатдаун, должен найти после, а не перед. 
//добавить бд. 