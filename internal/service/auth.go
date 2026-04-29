package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

var (
	ErrInvalidAuthData = errors.New("invalid auth data")
	ErrLoginTaken      = errors.New("login already taken")
)

type UserRepository interface {
	CreateUser(login string, passwordHash string) (model.User, error)
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