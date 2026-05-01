package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"log"
	"strconv"
	"io"
	"strings"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

type AuthService interface {
	Register(login string, password string) (model.User, error)
	Login(login string, password string) (model.User, error)
}

type Handler struct {
	auth   AuthService
	orders OrderService
}

type OrderService interface {
	UploadOrder(userID int64, number string) (model.Order, error)
}

const authCookieName = "user_id"

func NewHandler(auth AuthService, orders OrderService) *Handler {
	return &Handler{
		auth:   auth,
		orders: orders,
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

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatInt(userID, 10)))
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

func setAuthCookie(w http.ResponseWriter, userID int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    strconv.FormatInt(userID, 10),
		Path:     "/",
		HttpOnly: true,
	})
}
