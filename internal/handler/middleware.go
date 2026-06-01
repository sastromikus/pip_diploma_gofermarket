package handler

import (
	"context"
	"net/http"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/auth"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

type UserChecker interface {
	GetUserByID(ctx context.Context, userID int64) (model.User, error)
}

func AuthMiddleware(authManager *auth.Manager, users UserChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authCookieName)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, err := authManager.ParseToken(cookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if _, err := users.GetUserByID(r.Context(), userID); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := contextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
