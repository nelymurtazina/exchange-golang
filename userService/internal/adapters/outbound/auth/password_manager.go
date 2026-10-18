package auth

import "golang.org/x/crypto/bcrypt"

type PasswordManager struct{
	bcryptCost int
}

func NewPasswordManager(cost int) *PasswordManager {
	return &PasswordManager{bcryptCost: cost}
}

func (p *PasswordManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.bcryptCost)
	if err != nil {
		return "", err
	} 
	return string(bytes), err
}

func (p *PasswordManager) CheckPassword(password, hash string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}