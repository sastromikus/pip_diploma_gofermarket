package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type AuthService interface {
	Register(login string, password string) (model.User, error)
	Login(login string, password string) (model.User, error)
}

type Handler struct {
	auth    AuthService
	orders  OrderService
	balance BalanceService
}

type OrderService interface {
	UploadOrder(userID int64, number string) (model.Order, error)
	GetOrders(userID int64) ([]model.Order, error)
}

type BalanceService interface {
	GetBalance(userID int64) (model.Balance, error)
	Withdraw(userID int64, order string, sum float64) error
	GetWithdrawals(userID int64) ([]model.Withdrawal, error)
}

const authCookieName = "user_id"

func NewHandler(auth AuthService, orders OrderService, balance BalanceService) *Handler {
	return &Handler{
		auth:    auth,
		orders:  orders,
		balance: balance,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.auth.Register(req.Login, req.Password)
	if err != nil {
		log.Printf("register error: %v", err)

		switch {
		case errors.Is(err, service.ErrInvalidAuthData):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, service.ErrLoginTaken):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	setAuthCookie(w, user.ID)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.auth.Login(req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAuthData):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidLogin):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	setAuthCookie(w, user.ID)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.orders.UploadOrder(userID, number)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrOrderUploadedByUser):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, service.ErrOrderUploadedByOther):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := h.orders.GetOrders(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]model.OrderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, model.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := h.balance.GetBalance(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		return
	}
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req model.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.balance.Withdraw(userID, req.Order, req.Sum); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidWithdrawOrder):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrInvalidWithdrawSum):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, service.ErrInsufficientFunds):
			w.WriteHeader(http.StatusPaymentRequired)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balance.GetWithdrawals(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]model.WithdrawalResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		response = append(response, model.WithdrawalResponse{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}

func setAuthCookie(w http.ResponseWriter, userID int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    strconv.FormatInt(userID, 10),
		Path:     "/",
		HttpOnly: true,
	})
}
