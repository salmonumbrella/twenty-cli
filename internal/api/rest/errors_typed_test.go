package rest

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrUnauthorized(t *testing.T) {
	err := &ErrUnauthorized{Message: "token expired"}

	if err.Error() != "unauthorized: token expired" {
		t.Errorf("got %q, want %q", err.Error(), "unauthorized: token expired")
	}

	// Test errors.Is
	var target *ErrUnauthorized
	if !errors.As(err, &target) {
		t.Error("errors.As should match ErrUnauthorized")
	}
}

func TestErrNotFound(t *testing.T) {
	err := &ErrNotFound{Resource: "person", ID: "abc123"}

	want := "not found: person with id abc123"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestErrNotFoundEmptyID(t *testing.T) {
	err := &ErrNotFound{Resource: "person", ID: ""}

	want := "not found: person"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestErrRateLimited(t *testing.T) {
	err := &ErrRateLimited{RetryAfter: 5}

	want := "rate limited: retry after 5s"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestErrValidation(t *testing.T) {
	err := &ErrValidation{Field: "email", Message: "invalid format"}

	want := "validation error: email - invalid format"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestParseAPIErrorTyped(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		retryAfter string
		wantType   string
	}{
		{
			name:       "unauthorized",
			statusCode: 401,
			body:       `{"error":{"code":"UNAUTHENTICATED","message":"invalid token"}}`,
			wantType:   "*rest.ErrUnauthorized",
		},
		{
			name:       "not found",
			statusCode: 404,
			body:       `{"error":{"code":"NOT_FOUND","message":"person not found"}}`,
			wantType:   "*rest.ErrNotFound",
		},
		{
			name:       "rate limited",
			statusCode: 429,
			body:       `{"error":{"code":"RATE_LIMITED","message":"too many requests"}}`,
			wantType:   "*rest.ErrRateLimited",
		},
		{
			name:       "validation",
			statusCode: 400,
			body:       `{"error":{"code":"VALIDATION_ERROR","message":"email is invalid"}}`,
			wantType:   "*rest.ErrValidation",
		},
		{
			name:       "generic error",
			statusCode: 500,
			body:       `{"error":{"code":"INTERNAL","message":"server error"}}`,
			wantType:   "*rest.APIError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.statusCode, []byte(tt.body), tt.retryAfter)
			gotType := fmt.Sprintf("%T", err)
			if gotType != tt.wantType {
				t.Errorf("parseAPIError() type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestParseAPIErrorRateLimitedWithRetryAfter(t *testing.T) {
	tests := []struct {
		name           string
		retryAfter     string
		wantRetryAfter int
	}{
		{
			name:           "numeric seconds",
			retryAfter:     "30",
			wantRetryAfter: 30,
		},
		{
			name:           "empty header",
			retryAfter:     "",
			wantRetryAfter: 0,
		},
		{
			name:           "invalid value",
			retryAfter:     "invalid",
			wantRetryAfter: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(429, []byte(`{}`), tt.retryAfter)
			rateLimitErr, ok := err.(*ErrRateLimited)
			if !ok {
				t.Fatalf("expected *ErrRateLimited, got %T", err)
			}
			if rateLimitErr.RetryAfter != tt.wantRetryAfter {
				t.Errorf("RetryAfter = %d, want %d", rateLimitErr.RetryAfter, tt.wantRetryAfter)
			}
		})
	}
}
