package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, login string, passwordHash string) (model.User, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, login, password_hash
	`

	var user model.User

	err := r.db.QueryRow(
		ctx,
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
			return model.User{}, ErrUserAlreadyExists
		}

		return model.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	query := `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`

	var user model.User

	err := r.db.QueryRow(
		ctx,
		query,
		login,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}

		return model.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, userID int64) (model.User, error) {
	query := `
		SELECT id, login, password_hash
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}

		return model.User{}, err
	}

	return user, nil
}
