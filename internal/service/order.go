package service

import (
	"errors"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/luhn"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

var (
	ErrInvalidOrderNumber   = errors.New("invalid order number")
	ErrOrderUploadedByUser  = errors.New("order already uploaded by this user")
	ErrOrderUploadedByOther = errors.New("order already uploaded by another user")
)

type OrderRepository interface {
	CreateOrder(userID int64, number string) (model.Order, error)
}

type OrderService struct {
	orders OrderRepository
}

func NewOrderService(orders OrderRepository) *OrderService {
	return &OrderService{
		orders: orders,
	}
}

func (s *OrderService) UploadOrder(userID int64, number string) (model.Order, error) {
	if !luhn.Valid(number) {
		return model.Order{}, ErrInvalidOrderNumber
	}

	return s.orders.CreateOrder(userID, number)
}