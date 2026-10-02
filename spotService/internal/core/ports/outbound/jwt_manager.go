package outbound

type JWTValidator interface {
	ValidateAccessToken(token string) (userID string, role string, err error)
}