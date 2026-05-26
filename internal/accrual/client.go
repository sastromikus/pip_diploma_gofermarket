package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func (c *Client) GetOrder(number string) (model.AccrualResponse, time.Duration, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)

	resp, err := c.httpClient.Get(url)
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
	if value == "" {
		return time.Second
	}

	seconds, err := time.ParseDuration(value + "s")
	if err == nil {
		return seconds
	}

	return time.Second
}
