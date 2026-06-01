package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
)

var (
	ErrInvalidAuthData = errors.New("invalid auth data")
	ErrLoginTaken      = errors.New("login already taken")
	ErrInvalidLogin    = errors.New("invalid login or password")
)

type UserRepository interface {
	CreateUser(ctx context.Context, login string, passwordHash string) (model.User, error)
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	GetUserByID(ctx context.Context, userID int64) (model.User, error)
}

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{
		users: users,
	}
}

func (s *AuthService) Register(ctx context.Context, login string, password string) (model.User, error) {
	if login == "" || password == "" {
		return model.User{}, ErrInvalidAuthData
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user, err := s.users.CreateUser(ctx, login, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return model.User{}, ErrLoginTaken
		}

		return model.User{}, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, login string, password string) (model.User, error) {
	if login == "" || password == "" {
		return model.User{}, ErrInvalidAuthData
	}

	user, err := s.users.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.User{}, ErrInvalidLogin
		}

		return model.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, ErrInvalidLogin
	}

	return user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (model.User, error) {
	if userID <= 0 {
		return model.User{}, ErrInvalidLogin
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.User{}, ErrInvalidLogin
		}

		return model.User{}, err
	}

	return user, nil
}
