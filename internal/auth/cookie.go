package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidToken = errors.New("invalid auth token")

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{
		secret: []byte(secret),
	}
}

func (m *Manager) BuildToken(userID int64) string {
	id := strconv.FormatInt(userID, 10)
	signature := m.sign(id)

	return id + ":" + signature
}

func (m *Manager) ParseToken(token string) (int64, error) {
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return 0, ErrInvalidToken
	}

	id := parts[0]
	signature := parts[1]

	expectedSignature := m.sign(id)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}

	return userID, nil
}

func (m *Manager) sign(value string) string {
	hash := hmac.New(sha256.New, m.secret)
	hash.Write([]byte(value))

	return hex.EncodeToString(hash.Sum(nil))
}
