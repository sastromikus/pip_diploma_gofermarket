package repository

import (
	_ "errors"
	"sync"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type MemoryUserRepository struct {
	mu     sync.Mutex
	nextID int64
	users  map[string]model.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		nextID: 1,
		users:  make(map[string]model.User),
	}
}

func (r *MemoryUserRepository) CreateUser(login string, passwordHash string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[login]; ok {
		return model.User{}, service.ErrLoginTaken
	}

	user := model.User{
		ID:           r.nextID,
		Login:        login,
		PasswordHash: passwordHash,
	}

	r.users[login] = user
	r.nextID++

	return user, nil
}

func (r *MemoryUserRepository) GetUserByLogin(login string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[login]
	if !ok {
		return model.User{}, service.ErrInvalidLogin
	}

	return user, nil
}