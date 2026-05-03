package model

import "time"

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawal struct {
	ID          int64
	UserID      int64
	Order       string
	Sum         float64
	ProcessedAt time.Time
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
