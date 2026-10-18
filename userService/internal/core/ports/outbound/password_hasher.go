package ports

type PasswordHasher interface {
	// HashPassword превращает открытый пароль в безопасный хеш для сохранения в БД.
	HashPassword(password string) (string, error)
	
	// CheckPassword сравнивает открытый пароль с хешем из БД.
	CheckPassword(password, hash string) bool
}