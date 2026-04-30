package handler

import (
	"context"
	"errors"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

var ErrNoUserID = errors.New("no user id in context")

func contextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func userIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	if !ok || userID <= 0 {
		return 0, ErrNoUserID
	}

	return userID, nil
}