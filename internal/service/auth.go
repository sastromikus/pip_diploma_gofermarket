package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

var (
	ErrInvalidAuthData = errors.New("invalid auth data")
	ErrLoginTaken      = errors.New("login already taken")
	ErrInvalidLogin    = errors.New("invalid login or password")
)

type UserRepository interface {
	CreateUser(login string, passwordHash string) (model.User, error)
	GetUserByLogin(login string) (model.User, error)
}

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{
		users: users,
	}
}

func (s *AuthService) Register(login string, password string) (model.User, error) {
	if login == "" || password == "" {
		return model.User{}, ErrInvalidAuthData
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	return s.users.CreateUser(login, string(hash))
}

func (s *AuthService) Login(login string, password string) (model.User, error) {
	if login == "" || password == "" {
		return model.User{}, ErrInvalidAuthData
	}

	user, err := s.users.GetUserByLogin(login)
	if err != nil {
		return model.User{}, ErrInvalidLogin
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, ErrInvalidLogin
	}

	return user, nil
}
