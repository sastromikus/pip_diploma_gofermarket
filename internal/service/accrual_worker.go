package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/accrual"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

const (
	defaultAccrualPollInterval = time.Second
	defaultAccrualBatchSize    = 10
	defaultAccrualWorkerCount  = 4
)

type AccrualClient interface {
	GetOrder(ctx context.Context, number string) (model.AccrualResponse, time.Duration, error)
}

type AccrualWorker struct {
	orders OrderRepository
	client AccrualClient
	logger *slog.Logger

	pollInterval time.Duration
	batchSize    int
	workerCount  int

	rateLimitMu    sync.Mutex
	rateLimitUntil time.Time

	wg sync.WaitGroup
}

func NewAccrualWorker(orders OrderRepository, client AccrualClient, logger *slog.Logger) *AccrualWorker {
	if logger == nil {
		logger = slog.Default()
	}

	return &AccrualWorker{
		orders:       orders,
		client:       client,
		logger:       logger,
		pollInterval: defaultAccrualPollInterval,
		batchSize:    defaultAccrualBatchSize,
		workerCount:  defaultAccrualWorkerCount,
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
		w.logger.Error("get pending orders", "error", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	jobs := make(chan model.Order)

	workerCount := w.workerCount
	if workerCount <= 0 {
		workerCount = 1
	}
	if workerCount > len(orders) {
		workerCount = len(orders)
	}

	var batchWG sync.WaitGroup
	batchWG.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			defer batchWG.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case order, ok := <-jobs:
					if !ok {
						return
					}

					w.processOrder(ctx, workerID, order)
				}
			}
		}(i + 1)
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			close(jobs)
			batchWG.Wait()
			return
		case jobs <- order:
		}
	}

	close(jobs)
	batchWG.Wait()
}

func (w *AccrualWorker) processOrder(ctx context.Context, workerID int, order model.Order) {
	if err := w.waitRateLimit(ctx); err != nil {
		return
	}

	result, retryAfter, err := w.client.GetOrder(ctx, order.Number)
	if err != nil {
		switch {
		case errors.Is(err, accrual.ErrOrderNotRegistered):
			return

		case errors.Is(err, accrual.ErrTooManyRequests):
			w.setRateLimit(retryAfter)
			w.logger.Warn(
				"accrual rate limit reached",
				"worker", workerID,
				"retry_after", retryAfter,
			)
			return

		default:
			w.logger.Error(
				"get accrual order",
				"worker", workerID,
				"order", order.Number,
				"error", err,
			)
			return
		}
	}

	status := mapAccrualStatus(result.Status)
	if status == "" {
		w.logger.Warn(
			"unknown accrual status",
			"worker", workerID,
			"order", order.Number,
			"status", result.Status,
		)
		return
	}

	if err := w.orders.UpdateOrderAccrual(ctx, order.Number, status, result.Accrual); err != nil {
		w.logger.Error(
			"update order accrual",
			"worker", workerID,
			"order", order.Number,
			"error", err,
		)
	}
}

func (w *AccrualWorker) waitRateLimit(ctx context.Context) error {
	for {
		w.rateLimitMu.Lock()
		wait := time.Until(w.rateLimitUntil)
		w.rateLimitMu.Unlock()

		if wait <= 0 {
			return nil
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (w *AccrualWorker) setRateLimit(retryAfter time.Duration) {
	if retryAfter <= 0 {
		retryAfter = time.Second
	}

	until := time.Now().Add(retryAfter)

	w.rateLimitMu.Lock()
	defer w.rateLimitMu.Unlock()

	if until.After(w.rateLimitUntil) {
		w.rateLimitUntil = until
	}
}

func mapAccrualStatus(status string) string {
	switch status {
	case "REGISTERED", model.OrderStatusProcessing:
		return model.OrderStatusProcessing
	case model.OrderStatusInvalid:
		return model.OrderStatusInvalid
	case model.OrderStatusProcessed:
		return model.OrderStatusProcessed
	default:
		return ""
	}
}
