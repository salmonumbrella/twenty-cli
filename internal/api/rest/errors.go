package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// APIError is the generic API error for untyped errors
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func parseAPIError(statusCode int, body []byte, retryAfter string) error {
	// Try to parse error response
	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &errResp)

	code := errResp.Error.Code
	message := errResp.Error.Message
	if code == "" {
		code = http.StatusText(statusCode)
	}

	// Return typed errors based on status code
	switch statusCode {
	case http.StatusUnauthorized: // 401
		return &ErrUnauthorized{Message: message}
	case http.StatusNotFound: // 404
		// Try to extract resource type from message
		resource := "resource"
		if strings.Contains(strings.ToLower(message), "person") {
			resource = "person"
		} else if strings.Contains(strings.ToLower(message), "company") {
			resource = "company"
		}
		return &ErrNotFound{Resource: resource, ID: ""}
	case http.StatusTooManyRequests: // 429
		return &ErrRateLimited{RetryAfter: parseRetryAfter(retryAfter)}
	case http.StatusBadRequest: // 400
		if strings.Contains(strings.ToLower(code), "validation") ||
			strings.Contains(strings.ToLower(message), "invalid") {
			return &ErrValidation{Field: "", Message: message}
		}
	}

	// Default to generic APIError
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// parseRetryAfter extracts retry delay from header value
func parseRetryAfter(value string) int {
	if seconds, err := strconv.Atoi(value); err == nil {
		return seconds
	}
	return 0
}
