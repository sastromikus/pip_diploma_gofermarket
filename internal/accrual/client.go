package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

var (
	ErrOrderNotRegistered = errors.New("order is not registered in accrual system")
	ErrTooManyRequests    = errors.New("too many requests")
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	baseURL = strings.Replace(baseURL, "http://localhost:", "http://127.0.0.1:", 1)
	baseURL = strings.Replace(baseURL, "https://localhost:", "https://127.0.0.1:", 1)

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) GetOrder(ctx context.Context, number string) (model.AccrualResponse, time.Duration, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.AccrualResponse{}, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.AccrualResponse{}, 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result model.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return model.AccrualResponse{}, 0, err
		}

		return result, 0, nil

	case http.StatusNoContent:
		return model.AccrualResponse{}, 0, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return model.AccrualResponse{}, retryAfter, ErrTooManyRequests

	default:
		return model.AccrualResponse{}, 0, fmt.Errorf("unexpected accrual status: %d", resp.StatusCode)
	}
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Second
	}

	seconds, err := strconv.Atoi(value)
	if err == nil {
		if seconds <= 0 {
			return time.Second
		}

		return time.Duration(seconds) * time.Second
	}

	when, err := http.ParseTime(value)
	if err == nil {
		duration := time.Until(when)
		if duration > 0 {
			return duration
		}
	}

	return time.Second
}
