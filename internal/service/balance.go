package service

import (
	"errors"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/luhn"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

var (
	ErrInvalidWithdrawOrder = errors.New("invalid withdraw order")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidWithdrawSum   = errors.New("invalid withdraw sum")
)

type BalanceRepository interface {
	GetBalance(userID int64) (model.Balance, error)
	Withdraw(userID int64, order string, sum float64) error
	GetWithdrawals(userID int64) ([]model.Withdrawal, error)
}

type BalanceService struct {
	balance BalanceRepository
}

func NewBalanceService(balance BalanceRepository) *BalanceService {
	return &BalanceService{
		balance: balance,
	}
}

func (s *BalanceService) GetBalance(userID int64) (model.Balance, error) {
	return s.balance.GetBalance(userID)
}

func (s *BalanceService) Withdraw(userID int64, order string, sum float64) error {
	if !luhn.Valid(order) {
		return ErrInvalidWithdrawOrder
	}

	if sum <= 0 {
		return ErrInvalidWithdrawSum
	}

	return s.balance.Withdraw(userID, order, sum)
}

func (s *BalanceService) GetWithdrawals(userID int64) ([]model.Withdrawal, error) {
	return s.balance.GetWithdrawals(userID)
}
