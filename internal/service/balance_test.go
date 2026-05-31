package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type fakeBalanceRepository struct {
	balance     model.Balance
	withdrawals []model.Withdrawal
}

func (r *fakeBalanceRepository) GetBalance(_ context.Context, _ int64) (model.Balance, error) {
	return r.balance, nil
}

func (r *fakeBalanceRepository) Withdraw(_ context.Context, userID int64, order string, sum float64) error {
	if r.balance.Current < sum {
		return repository.ErrInsufficientFunds
	}

	r.balance.Current -= sum
	r.balance.Withdrawn += sum
	r.withdrawals = append(r.withdrawals, model.Withdrawal{
		UserID: userID,
		Order:  order,
		Sum:    sum,
	})

	return nil
}

func (r *fakeBalanceRepository) GetWithdrawals(_ context.Context, _ int64) ([]model.Withdrawal, error) {
	return r.withdrawals, nil
}

func TestBalanceServiceGetBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	balanceService := service.NewBalanceService(&fakeBalanceRepository{
		balance: model.Balance{Current: 100, Withdrawn: 50},
	})

	balance, err := balanceService.GetBalance(ctx, 1)
	if err != nil {
		t.Fatalf("get balance returned error: %v", err)
	}
	if balance.Current != 100 || balance.Withdrawn != 50 {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestBalanceServiceWithdraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &fakeBalanceRepository{
		balance: model.Balance{Current: 100},
	}
	balanceService := service.NewBalanceService(repo)

	if err := balanceService.Withdraw(ctx, 1, "79927398713", 30); err != nil {
		t.Fatalf("withdraw returned error: %v", err)
	}

	balance, err := balanceService.GetBalance(ctx, 1)
	if err != nil {
		t.Fatalf("get balance returned error: %v", err)
	}
	if balance.Current != 70 || balance.Withdrawn != 30 {
		t.Fatalf("unexpected balance after withdraw: %+v", balance)
	}
}

func TestBalanceServiceWithdrawValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	balanceService := service.NewBalanceService(&fakeBalanceRepository{
		balance: model.Balance{Current: 100},
	})

	tests := []struct {
		name    string
		order   string
		sum     float64
		expects error
	}{
		{name: "invalid order", order: "12345", sum: 10, expects: service.ErrInvalidWithdrawOrder},
		{name: "zero sum", order: "79927398713", sum: 0, expects: service.ErrInvalidWithdrawSum},
		{name: "negative sum", order: "79927398713", sum: -1, expects: service.ErrInvalidWithdrawSum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := balanceService.Withdraw(ctx, 1, tt.order, tt.sum)
			if !errors.Is(err, tt.expects) {
				t.Fatalf("expected %v, got %v", tt.expects, err)
			}
		})
	}
}

func TestBalanceServiceWithdrawInsufficientFunds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	balanceService := service.NewBalanceService(&fakeBalanceRepository{
		balance: model.Balance{Current: 10},
	})

	err := balanceService.Withdraw(ctx, 1, "79927398713", 30)
	if !errors.Is(err, service.ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
}
