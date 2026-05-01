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
