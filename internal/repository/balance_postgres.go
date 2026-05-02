package repository

import (
	"context"
	"database/sql"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

type PostgresBalanceRepository struct {
	db *sql.DB
}

func NewPostgresBalanceRepository(db *sql.DB) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{
		db: db,
	}
}

func (r *PostgresBalanceRepository) GetBalance(userID int64) (model.Balance, error) {
	query := `
		SELECT
			COALESCE((
				SELECT SUM(accrual)
				FROM orders
				WHERE user_id = $1
				  AND status = 'PROCESSED'
				  AND accrual IS NOT NULL
			), 0)
			-
			COALESCE((
				SELECT SUM(sum)
				FROM withdrawals
				WHERE user_id = $1
			), 0) AS current,

			COALESCE((
				SELECT SUM(sum)
				FROM withdrawals
				WHERE user_id = $1
			), 0) AS withdrawn
	`

	var balance model.Balance

	err := r.db.QueryRowContext(
		context.Background(),
		query,
		userID,
	).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)

	if err != nil {
		return model.Balance{}, err
	}

	return balance, nil
}
