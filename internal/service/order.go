package service

import (
	"context"
	"errors"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/luhn"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
)

var (
	ErrInvalidOrderNumber   = errors.New("invalid order number")
	ErrOrderUploadedByUser  = errors.New("order already uploaded by this user")
	ErrOrderUploadedByOther = errors.New("order already uploaded by another user")
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID int64, number string) (model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]model.Order, error)

	GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error)
	UpdateOrderAccrual(ctx context.Context, number string, status string, accrual *float64) error
}

type OrderService struct {
	orders OrderRepository
}

func NewOrderService(orders OrderRepository) *OrderService {
	return &OrderService{
		orders: orders,
	}
}

func (s *OrderService) UploadOrder(ctx context.Context, userID int64, number string) (model.Order, error) {
	if !luhn.Valid(number) {
		return model.Order{}, ErrInvalidOrderNumber
	}

	order, err := s.orders.CreateOrder(ctx, userID, number)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOrderUploadedByUser):
			return model.Order{}, ErrOrderUploadedByUser
		case errors.Is(err, repository.ErrOrderUploadedByOther):
			return model.Order{}, ErrOrderUploadedByOther
		default:
			return model.Order{}, err
		}
	}

	return order, nil
}

func (s *OrderService) GetOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.orders.GetOrdersByUserID(ctx, userID)
}
