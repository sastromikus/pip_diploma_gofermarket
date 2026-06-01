package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type fakeOrderRepository struct {
	orders map[string]model.Order
	nextID int64
}

func newFakeOrderRepository() *fakeOrderRepository {
	return &fakeOrderRepository{
		orders: make(map[string]model.Order),
		nextID: 1,
	}
}

func (r *fakeOrderRepository) CreateOrder(_ context.Context, userID int64, number string) (model.Order, error) {
	if existing, exists := r.orders[number]; exists {
		if existing.UserID == userID {
			return model.Order{}, repository.ErrOrderUploadedByUser
		}

		return model.Order{}, repository.ErrOrderUploadedByOther
	}

	order := model.Order{
		ID:         r.nextID,
		UserID:     userID,
		Number:     number,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now(),
	}
	r.nextID++
	r.orders[number] = order

	return order, nil
}

func (r *fakeOrderRepository) GetOrdersByUserID(_ context.Context, userID int64) ([]model.Order, error) {
	orders := make([]model.Order, 0)
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (r *fakeOrderRepository) GetPendingOrders(_ context.Context, _ int) ([]model.Order, error) {
	return nil, nil
}

func (r *fakeOrderRepository) UpdateOrderAccrual(_ context.Context, _ string, _ string, _ *float64) error {
	return nil
}

func TestOrderServiceUploadOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	orderService := service.NewOrderService(newFakeOrderRepository())

	order, err := orderService.UploadOrder(ctx, 1, "79927398713")
	if err != nil {
		t.Fatalf("upload order returned error: %v", err)
	}
	if order.UserID != 1 {
		t.Fatalf("expected user id 1, got %d", order.UserID)
	}
	if order.Number != "79927398713" {
		t.Fatalf("expected order number %q, got %q", "79927398713", order.Number)
	}
	if order.Status != model.OrderStatusNew {
		t.Fatalf("expected status %q, got %q", model.OrderStatusNew, order.Status)
	}
}

func TestOrderServiceUploadOrderInvalidNumber(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	orderService := service.NewOrderService(newFakeOrderRepository())

	_, err := orderService.UploadOrder(ctx, 1, "12345")
	if !errors.Is(err, service.ErrInvalidOrderNumber) {
		t.Fatalf("expected ErrInvalidOrderNumber, got %v", err)
	}
}

func TestOrderServiceUploadOrderDuplicateBySameUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	orderService := service.NewOrderService(newFakeOrderRepository())

	if _, err := orderService.UploadOrder(ctx, 1, "79927398713"); err != nil {
		t.Fatalf("upload first order: %v", err)
	}

	_, err := orderService.UploadOrder(ctx, 1, "79927398713")
	if !errors.Is(err, service.ErrOrderUploadedByUser) {
		t.Fatalf("expected ErrOrderUploadedByUser, got %v", err)
	}
}

func TestOrderServiceUploadOrderDuplicateByAnotherUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	orderService := service.NewOrderService(newFakeOrderRepository())

	if _, err := orderService.UploadOrder(ctx, 1, "79927398713"); err != nil {
		t.Fatalf("upload first order: %v", err)
	}

	_, err := orderService.UploadOrder(ctx, 2, "79927398713")
	if !errors.Is(err, service.ErrOrderUploadedByOther) {
		t.Fatalf("expected ErrOrderUploadedByOther, got %v", err)
	}
}
