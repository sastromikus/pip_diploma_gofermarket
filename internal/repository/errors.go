package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrOrderAlreadyExists   = errors.New("order already exists")
	ErrOrderUploadedByUser  = errors.New("order already uploaded by this user")
	ErrOrderUploadedByOther = errors.New("order already uploaded by another user")
	ErrOrderNotFound        = errors.New("order not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
