package service

import "github.com/sastromikus/pip_diploma_gofermarket/internal/model"

type BalanceRepository interface {
	GetBalance(userID int64) (model.Balance, error)
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
