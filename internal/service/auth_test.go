package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type fakeUserRepository struct {
	users  map[string]model.User
	nextID int64
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		users:  make(map[string]model.User),
		nextID: 1,
	}
}

func (r *fakeUserRepository) CreateUser(_ context.Context, login string, passwordHash string) (model.User, error) {
	if _, exists := r.users[login]; exists {
		return model.User{}, repository.ErrUserAlreadyExists
	}

	user := model.User{
		ID:           r.nextID,
		Login:        login,
		PasswordHash: passwordHash,
	}
	r.nextID++
	r.users[login] = user

	return user, nil
}

func (r *fakeUserRepository) GetUserByLogin(_ context.Context, login string) (model.User, error) {
	user, exists := r.users[login]
	if !exists {
		return model.User{}, repository.ErrUserNotFound
	}

	return user, nil
}

func (r *fakeUserRepository) GetUserByID(_ context.Context, userID int64) (model.User, error) {
	for _, user := range r.users {
		if user.ID == userID {
			return user, nil
		}
	}

	return model.User{}, repository.ErrUserNotFound
}

func TestAuthServiceRegister(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authService := service.NewAuthService(newFakeUserRepository())

	user, err := authService.Register(ctx, "user", "password")
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if user.ID == 0 {
		t.Fatalf("expected user id to be set")
	}
	if user.Login != "user" {
		t.Fatalf("expected login %q, got %q", "user", user.Login)
	}
	if user.PasswordHash == "" || user.PasswordHash == "password" {
		t.Fatalf("expected password to be hashed")
	}
}

func TestAuthServiceRegisterValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authService := service.NewAuthService(newFakeUserRepository())

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{name: "empty login", login: "", password: "password"},
		{name: "empty password", login: "user", password: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := authService.Register(ctx, tt.login, tt.password)
			if !errors.Is(err, service.ErrInvalidAuthData) {
				t.Fatalf("expected ErrInvalidAuthData, got %v", err)
			}
		})
	}
}

func TestAuthServiceRegisterDuplicateLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authService := service.NewAuthService(newFakeUserRepository())

	if _, err := authService.Register(ctx, "user", "password"); err != nil {
		t.Fatalf("register first user: %v", err)
	}

	_, err := authService.Register(ctx, "user", "another-password")
	if !errors.Is(err, service.ErrLoginTaken) {
		t.Fatalf("expected ErrLoginTaken, got %v", err)
	}
}

func TestAuthServiceLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authService := service.NewAuthService(newFakeUserRepository())

	registered, err := authService.Register(ctx, "user", "password")
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	loggedIn, err := authService.Login(ctx, "user", "password")
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if loggedIn.ID != registered.ID {
		t.Fatalf("expected user id %d, got %d", registered.ID, loggedIn.ID)
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	authService := service.NewAuthService(newFakeUserRepository())

	if _, err := authService.Register(ctx, "user", "password"); err != nil {
		t.Fatalf("register user: %v", err)
	}

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{name: "unknown login", login: "missing", password: "password"},
		{name: "wrong password", login: "user", password: "wrong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := authService.Login(ctx, tt.login, tt.password)
			if !errors.Is(err, service.ErrInvalidLogin) {
				t.Fatalf("expected ErrInvalidLogin, got %v", err)
			}
		})
	}
}
