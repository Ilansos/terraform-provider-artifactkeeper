package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    any
	RetryAfter *time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	details := formatErrorDetails(e.Details)
	if e.Code != "" && e.Message != "" {
		if details != "" {
			return fmt.Sprintf("artifact keeper API error (%d %s): %s: %s", e.StatusCode, e.Code, e.Message, details)
		}
		return fmt.Sprintf("artifact keeper API error (%d %s): %s", e.StatusCode, e.Code, e.Message)
	}
	if e.Message != "" {
		if details != "" {
			return fmt.Sprintf("artifact keeper API error (%d): %s: %s", e.StatusCode, e.Message, details)
		}
		return fmt.Sprintf("artifact keeper API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("artifact keeper API error (%d)", e.StatusCode)
}

func (e *APIError) IsNotFound() bool {
	return e != nil && e.StatusCode == http.StatusNotFound
}

func decodeAPIError(res *http.Response) error {
	apiErr := &APIError{
		StatusCode: res.StatusCode,
		Message:    http.StatusText(res.StatusCode),
	}

	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details any    `json:"details"`
		} `json:"error"`
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err == nil {
		if payload.Error != nil {
			apiErr.Code = payload.Error.Code
			apiErr.Message = payload.Error.Message
			apiErr.Details = payload.Error.Details
		} else {
			apiErr.Code = payload.Code
			if payload.Message != "" {
				apiErr.Message = payload.Message
			}
			apiErr.Details = payload.Details
		}
	} else if text := strings.TrimSpace(string(body)); text != "" {
		apiErr.Details = text
	}

	if res.StatusCode == http.StatusTooManyRequests {
		apiErr.RetryAfter = parseRetryAfter(res.Header.Get("Retry-After"))
	}

	return apiErr
}

func formatErrorDetails(details any) string {
	if details == nil {
		return ""
	}
	switch v := details.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(encoded)
	}
}

func parseRetryAfter(value string) *time.Duration {
	if value == "" {
		return nil
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		duration := time.Duration(seconds) * time.Second
		return &duration
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		duration := time.Until(retryAt)
		if duration < 0 {
			duration = 0
		}
		return &duration
	}
	return nil
}
