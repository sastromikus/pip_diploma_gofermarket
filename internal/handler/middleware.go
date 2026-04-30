package handler

import (
	"net/http"
	"strconv"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authCookieName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(cookie.Value, 10, 64)
		if err != nil || userID <= 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := contextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
