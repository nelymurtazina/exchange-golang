package auth

import "golang.org/x/crypto/bcrypt"

type PasswordManager struct{
	cost int
}

func NewPasswordManager(cost int) *PasswordManager {
	return &PasswordManager{cost: cost}
}

func (p *PasswordManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost) 
	return string(bytes), err
}

func (p *PasswordManager) CheckPassword(password, hash string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}