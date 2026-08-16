package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
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
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidUserID      = errors.New("invalid user_id")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewUser(id, username, email, password, role string) (*User, error) {
	if err := ValidateUserName(username); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if password == "" {
		return nil, ErrInvalidPassword
	}
	if err := ValidateUserID(id); err != nil {
		return nil, err
	}

	return &User{
		UserID:    id,
		Username:  username,
		Email:     email,
		Password:  password,
		Role:      role,
		Active:    true,
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
	//net.mail, uuid проверки добавить.
	_, err := mail.ParseAddress(email) //все решаем только этой библиотекой 
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

// Методы для изменения состояния с валидацией
//здесь еще будут методы для изменения состояния с валидацией при update
func (u *User) UpdateUsername(newUsername string) error {
	if err := ValidateUserName(newUsername); err != nil {
		return err
	}
	u.Username = newUsername
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdateEmail(newEmail string) error {
	if err := ValidateEmail(newEmail); err != nil {
		return err
	}
	u.Email = newEmail
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdatePassword(newPassword string) error {
	if newPassword == "" {
		return ErrInvalidPassword
	}
	u.Password = newPassword
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}

func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}




//реализация кеширования, не будет методов jwtMeneger. ТОлько работа с jwt. Добавить input Output(для решистрации). Более сложна логика для email.Обработка ошибок без nil. а реальные ошибки, если уникальная ошибка, то конвертирую в domain и прокидываю ее(чтобы сервер отличил ошибку)
//refreshToken изменить , добавить в порты. 
//сервис логина, даже если пользователь не найдет, нужно защитить систему от атак. + добавить миграцию бд. (без миграции не валиден). Закрывать после грейсшатдаун. есть готовые паттерны для грейсоушшатдаун, должен найти после, а не перед. 
//добавить бд. 