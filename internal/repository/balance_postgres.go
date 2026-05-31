package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

type PostgresBalanceRepository struct {
	db *pgxpool.Pool
}

func NewPostgresBalanceRepository(db *pgxpool.Pool) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{
		db: db,
	}
}

func (r *PostgresBalanceRepository) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
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

	err := r.db.QueryRow(
		ctx,
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

func (r *PostgresBalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)

	for rows.Next() {
		var withdrawal model.Withdrawal

		if err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.Order,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		); err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (r *PostgresBalanceRepository) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedUserID int64

	if err := tx.QueryRow(
		ctx,
		`SELECT id FROM users WHERE id = $1 FOR UPDATE`,
		userID,
	).Scan(&lockedUserID); err != nil {
		return err
	}

	var current float64

	queryBalance := `
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
			), 0) AS current
	`

	if err := tx.QueryRow(ctx, queryBalance, userID).Scan(&current); err != nil {
		return err
	}

	if current < sum {
		return ErrInsufficientFunds
	}

	queryWithdraw := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`

	if _, err := tx.Exec(ctx, queryWithdraw, userID, order, sum); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
