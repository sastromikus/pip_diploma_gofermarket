package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) CreateUser(login string, passwordHash string) (model.User, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, login, password_hash
	`

	var user model.User

	err := r.db.QueryRowContext(
		context.Background(),
		query,
		login,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, service.ErrLoginTaken
		}

		return model.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByLogin(login string) (model.User, error) {
	query := `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`

	var user model.User

	err := r.db.QueryRowContext(
		context.Background(),
		query,
		login,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, service.ErrInvalidLogin
		}

		return model.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByID(userID int64) (model.User, error) {
	query := `
		SELECT id, login, password_hash
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := r.db.QueryRowContext(
		context.Background(),
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, service.ErrInvalidLogin
		}

		return model.User{}, err
	}

	return user, nil
}
