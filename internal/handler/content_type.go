package handler

import (
	"net/http"
	"strings"
)

func hasContentType(r *http.Request, expected string) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return false
	}

	contentType = strings.ToLower(contentType)
	expected = strings.ToLower(expected)

	return strings.HasPrefix(contentType, expected)
}
