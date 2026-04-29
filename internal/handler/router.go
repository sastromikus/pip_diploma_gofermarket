package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/api/user/register", Register)
	r.Post("/api/user/login", Login)

	r.Post("/api/user/orders", UploadOrder)
	r.Get("/api/user/orders", GetOrders)

	r.Get("/api/user/balance", GetBalance)
	r.Post("/api/user/balance/withdraw", Withdraw)
	r.Get("/api/user/withdrawals", GetWithdrawals)

	return r
}