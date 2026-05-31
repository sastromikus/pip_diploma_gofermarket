package service

import (
	"context"
	"errors"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/luhn"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
)

var (
	ErrInvalidWithdrawOrder = errors.New("invalid withdraw order")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidWithdrawSum   = errors.New("invalid withdraw sum")
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

type BalanceService struct {
	balance BalanceRepository
}

func NewBalanceService(balance BalanceRepository) *BalanceService {
	return &BalanceService{
		balance: balance,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	return s.balance.GetBalance(ctx, userID)
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	if !luhn.Valid(order) {
		return ErrInvalidWithdrawOrder
	}

	if sum <= 0 {
		return ErrInvalidWithdrawSum
	}

	if err := s.balance.Withdraw(ctx, userID, order, sum); err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}

		return err
	}

	return nil
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	return s.balance.GetWithdrawals(ctx, userID)
}
