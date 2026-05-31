package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/accrual"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

type AccrualClient interface {
	GetOrder(ctx context.Context, number string) (model.AccrualResponse, time.Duration, error)
}

type AccrualWorker struct {
	orders OrderRepository
	client AccrualClient

	pollInterval time.Duration
	batchSize    int

	wg sync.WaitGroup
}

func NewAccrualWorker(orders OrderRepository, client AccrualClient) *AccrualWorker {
	return &AccrualWorker{
		orders:       orders,
		client:       client,
		pollInterval: time.Second,
		batchSize:    10,
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.run(ctx)
	}()
}

func (w *AccrualWorker) Wait() {
	w.wg.Wait()
}

func (w *AccrualWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *AccrualWorker) processBatch(ctx context.Context) {
	orders, err := w.orders.GetPendingOrders(ctx, w.batchSize)
	if err != nil {
		log.Printf("accrual worker: get pending orders: %v", err)
		return
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, retryAfter, err := w.client.GetOrder(ctx, order.Number)

		if err != nil {
			switch {
			case errors.Is(err, accrual.ErrOrderNotRegistered):
				continue

			case errors.Is(err, accrual.ErrTooManyRequests):
				if retryAfter <= 0 {
					retryAfter = time.Second
				}

				log.Printf("accrual worker: too many requests, retry after %s", retryAfter)

				timer := time.NewTimer(retryAfter)
				select {
				case <-ctx.Done():
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					return
				case <-timer.C:
					return
				}

			default:
				log.Printf("accrual worker: get order %s: %v", order.Number, err)
				continue
			}
		}

		status := mapAccrualStatus(result.Status)
		if status == "" {
			continue
		}

		if err := w.orders.UpdateOrderAccrual(ctx, order.Number, status, result.Accrual); err != nil {
			log.Printf("accrual worker: update order %s: %v", order.Number, err)
		}
	}
}

func mapAccrualStatus(status string) string {
	switch status {
	case "REGISTERED":
		return model.OrderStatusNew
	case model.OrderStatusProcessing:
		return model.OrderStatusProcessing
	case model.OrderStatusInvalid:
		return model.OrderStatusInvalid
	case model.OrderStatusProcessed:
		return model.OrderStatusProcessed
	default:
		return ""
	}
}
