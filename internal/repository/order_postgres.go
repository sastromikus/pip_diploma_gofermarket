package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, userID int64, number string) (model.Order, error) {
	query := `
		INSERT INTO orders (user_id, number, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, number, status, accrual, uploaded_at
	`

	var order model.Order

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		number,
		model.OrderStatusNew,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err == nil {
		return order, nil
	}

	if !isUniqueViolation(err) {
		return model.Order{}, err
	}

	existingOrder, err := r.GetOrderByNumber(ctx, number)
	if err != nil {
		return model.Order{}, err
	}

	if existingOrder.UserID == userID {
		return model.Order{}, ErrOrderUploadedByUser
	}

	return model.Order{}, ErrOrderUploadedByOther
}

func (r *PostgresOrderRepository) GetOrderByNumber(ctx context.Context, number string) (model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order model.Order

	err := r.db.QueryRow(
		ctx,
		query,
		number,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}

		return model.Order{}, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)

	for rows.Next() {
		var order model.Order

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresOrderRepository) GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY uploaded_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		model.OrderStatusNew,
		model.OrderStatusProcessing,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)

	for rows.Next() {
		var order model.Order

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresOrderRepository) UpdateOrderAccrual(ctx context.Context, number string, status string, accrual *float64) error {
	var accrualValue any
	if accrual != nil {
		accrualValue = *accrual
	}

	query := `
		UPDATE orders
		SET status = $1::text,
		    accrual = CASE
		        WHEN $1::text = $2::text THEN $3::numeric
		        ELSE NULL::numeric
		    END,
		    updated_at = NOW()
		WHERE number = $4
		  AND status IN ($5::text, $6::text)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		status,
		model.OrderStatusProcessed,
		accrualValue,
		number,
		model.OrderStatusNew,
		model.OrderStatusProcessing,
	)

	return err
}
