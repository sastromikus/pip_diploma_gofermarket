package repository

import (
	"sync"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type MemoryOrderRepository struct {
	mu     sync.Mutex
	nextID int64
	orders map[string]model.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		nextID: 1,
		orders: make(map[string]model.Order),
	}
}

func (r *MemoryOrderRepository) CreateOrder(userID int64, number string) (model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order, ok := r.orders[number]; ok {
		if order.UserID == userID {
			return model.Order{}, service.ErrOrderUploadedByUser
		}

		return model.Order{}, service.ErrOrderUploadedByOther
	}

	order := model.Order{
		ID:         r.nextID,
		UserID:     userID,
		Number:     number,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	r.orders[number] = order
	r.nextID++

	return order, nil
}

func (r *MemoryOrderRepository) GetOrdersByUserID(userID int64) ([]model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders := make([]model.Order, 0)

	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (r *MemoryOrderRepository) GetPendingOrders(limit int) ([]model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders := make([]model.Order, 0)

	for _, order := range r.orders {
		if order.Status == model.OrderStatusNew || order.Status == model.OrderStatusProcessing {
			orders = append(orders, order)

			if len(orders) >= limit {
				break
			}
		}
	}

	return orders, nil
}

func (r *MemoryOrderRepository) UpdateOrderAccrual(number string, status string, accrual *float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[number]
	if !ok {
		return nil
	}

	order.Status = status
	order.Accrual = accrual
	r.orders[number] = order

	return nil
}
