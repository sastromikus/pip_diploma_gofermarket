package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"log"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type AuthService interface {
	Register(login string, password string) (model.User, error)
}

type Handler struct {
	auth AuthService
}

func NewHandler(auth AuthService) *Handler {
	return &Handler{
		auth: auth,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := h.auth.Register(req.Login, req.Password)
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

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}