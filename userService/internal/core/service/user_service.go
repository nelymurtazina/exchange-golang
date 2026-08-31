package service

import (
	"context"
	"test-project/userService/internal/core/domain"
	"test-project/userService/internal/core/ports"
)

type UserService struct {
	repo   ports.UserRepository
	jwt    JWTManagerInterface
	passwd PasswordManagerInterface
}

type JWTManagerInterface interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(accessToken string) (string, error)
	ValidateRefreshToken(accessToken string) (string, error)
	RefreshToken(refreshToken string) (string, error)
}

type PasswordManagerInterface interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

func NewUserService(repo ports.UserRepository, jwt JWTManagerInterface, passwd PasswordManagerInterface) ports.UserService {
	return &UserService{
		repo:   repo,
		jwt:    jwt,
		passwd: passwd,
	}
}

func (s *UserService) Register(ctx context.Context, input ports.RegisterInput) (*ports.RegisterOutput, error) {
	if err := domain.ValidateUserName(input.UserName); err != nil {
		return nil, domain.ErrInvalidUsername
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, domain.ErrInvalidEmail
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	existingUser, err := s.repo.GetByEmail(ctx, input.Email); 
	if err != nil && err != domain.ErrUserNotFound{
		return nil, domain.ErrInvalidEmail
	}
	if existingUser != nil{
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPaswword, err := s.passwd.HashPassword(input.Password)
	if err != nil{
		return nil,err
	}

	user, err := domain.NewUser(domain.NewUserID(), input.UserName, input.Email, hashedPaswword, domain.RoleUser)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx,user); err != nil{
		return  nil, domain.ErrInvalidUsername
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
	if err != nil {
		return nil, err
	}

	return &ports.RegisterOutput{
		User: user,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if err := domain.ValidateID(userID); err != nil{
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, input ports.LoginInput) (*ports.LoginOutput, error) {
	if err := domain.ValidateEmail(input.Email); err != nil{
		return nil, domain.ErrInvalidEmail
	}
	if input.Password == "" {
		return nil, domain.ErrInvalidPassword
	}

	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && err != domain.ErrUserNotFound{
		return nil, err
	}

	const fakePasswordHash = "$2a$10$eD30kk.F4KTkg3ovmAfcTeFzRykAl.YWrvrWxyK.k8RwswLEXFQAO"
	hashToCompare := fakePasswordHash
	if user != nil {
		hashToCompare = user.Password
	}

	passwordIsValid := s.passwd.CheckPassword((input.Password), hashToCompare)

	if user == nil || !passwordIsValid{
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.UserID)
	if err != nil {
		return nil, err
	}

	return &ports.LoginOutput{
		User: *user,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) ValidateToken(ctx context.Context, input ports.ValidateTokenInput) (*ports.ValidateTokenOutput, error) {
	if input.Token ==""{
		return nil, domain.ErrInvalidToken
	}
	
	userID, err := s.jwt.ValidateAccessToken(input.Token)
	if err != nil{
		return nil, domain.ErrInvalidToken
	}

	if err := domain.ValidateID(userID); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ports.ValidateTokenOutput{
		UserID: userID,
		Role:   user.Role,
	}, nil

}


func (s *UserService) RefreshToken(ctx context.Context, input ports.RefreshTokenInput) (*ports.RefreshTokenOutput, error) {
	if input.RefreshToken == ""{
		return nil, domain.ErrInvalidToken
	}

	userID, err := s.jwt.ValidateRefreshToken(input.RefreshToken) 
	if err != nil{
		return nil, domain.ErrInvalidToken
	}

	if err := domain.ValidateID(input.RefreshToken); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil && err == domain.ErrUserNotFound{
		return nil, domain.ErrInvalidToken
	}

	newAccessToken, err := s.jwt.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &ports.RefreshTokenOutput{
		User: *user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}


func (s *UserService) Logout(ctx context.Context, userID string) error {
	if err := domain.ValidateID(userID); err != nil{
		return domain.ErrInvalidUserID
	}

	return nil
}