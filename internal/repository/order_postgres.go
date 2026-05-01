package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) CreateOrder(userID int64, number string) (model.Order, error) {
	query := `
		INSERT INTO orders (user_id, number, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, number, status, accrual, uploaded_at
	`

	var order model.Order

	err := r.db.QueryRowContext(
		context.Background(),
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

	existingOrder, err := r.GetOrderByNumber(number)
	if err != nil {
		return model.Order{}, err
	}

	if existingOrder.UserID == userID {
		return model.Order{}, service.ErrOrderUploadedByUser
	}

	return model.Order{}, service.ErrOrderUploadedByOther
}

func (r *PostgresOrderRepository) GetOrderByNumber(number string) (model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order model.Order

	err := r.db.QueryRowContext(
		context.Background(),
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
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, sql.ErrNoRows
		}

		return model.Order{}, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetOrdersByUserID(userID int64) ([]model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(context.Background(), query, userID)
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
