package rest

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "with message",
			err: &APIError{
				StatusCode: 400,
				Code:       "validation_error",
				Message:    "Invalid input",
			},
			expected: "validation_error: Invalid input",
		},
		{
			name: "without message",
			err: &APIError{
				StatusCode: 500,
				Code:       "",
				Message:    "",
			},
			expected: "HTTP 500",
		},
		{
			name: "with code only",
			err: &APIError{
				StatusCode: 403,
				Code:       "forbidden",
				Message:    "",
			},
			expected: "HTTP 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		retryAfter string
		errType    string
	}{
		{
			name:       "401 unauthorized",
			statusCode: 401,
			body:       []byte(`{"error":{"message":"Invalid token"}}`),
			retryAfter: "",
			errType:    "*rest.ErrUnauthorized",
		},
		{
			name:       "404 not found - person",
			statusCode: 404,
			body:       []byte(`{"error":{"message":"Person not found"}}`),
			retryAfter: "",
			errType:    "*rest.ErrNotFound",
		},
		{
			name:       "404 not found - company",
			statusCode: 404,
			body:       []byte(`{"error":{"message":"Company not found"}}`),
			retryAfter: "",
			errType:    "*rest.ErrNotFound",
		},
		{
			name:       "429 rate limited",
			statusCode: 429,
			body:       []byte(`{"error":{"message":"Too many requests"}}`),
			retryAfter: "5",
			errType:    "*rest.ErrRateLimited",
		},
		{
			name:       "400 validation error",
			statusCode: 400,
			body:       []byte(`{"error":{"code":"validation_error","message":"Invalid field value"}}`),
			retryAfter: "",
			errType:    "*rest.ErrValidation",
		},
		{
			name:       "400 generic error",
			statusCode: 400,
			body:       []byte(`{"error":{"message":"Bad request"}}`),
			retryAfter: "",
			errType:    "*rest.APIError",
		},
		{
			name:       "500 internal server error",
			statusCode: 500,
			body:       []byte(`{"error":{"message":"Internal server error"}}`),
			retryAfter: "",
			errType:    "*rest.APIError",
		},
		{
			name:       "empty body",
			statusCode: 503,
			body:       []byte(``),
			retryAfter: "",
			errType:    "*rest.APIError",
		},
		{
			name:       "invalid json",
			statusCode: 500,
			body:       []byte(`not json`),
			retryAfter: "",
			errType:    "*rest.APIError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.statusCode, tt.body, tt.retryAfter)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Verify error type
			switch tt.errType {
			case "*rest.ErrUnauthorized":
				if _, ok := err.(*ErrUnauthorized); !ok {
					t.Errorf("expected *ErrUnauthorized, got %T", err)
				}
			case "*rest.ErrNotFound":
				if _, ok := err.(*ErrNotFound); !ok {
					t.Errorf("expected *ErrNotFound, got %T", err)
				}
			case "*rest.ErrRateLimited":
				if rl, ok := err.(*ErrRateLimited); !ok {
					t.Errorf("expected *ErrRateLimited, got %T", err)
				} else if rl.RetryAfter != 5 {
					t.Errorf("expected RetryAfter 5, got %d", rl.RetryAfter)
				}
			case "*rest.ErrValidation":
				if _, ok := err.(*ErrValidation); !ok {
					t.Errorf("expected *ErrValidation, got %T", err)
				}
			case "*rest.APIError":
				if _, ok := err.(*APIError); !ok {
					t.Errorf("expected *APIError, got %T", err)
				}
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected int
	}{
		{
			name:     "valid seconds",
			value:    "5",
			expected: 5,
		},
		{
			name:     "zero",
			value:    "0",
			expected: 0,
		},
		{
			name:     "large number",
			value:    "120",
			expected: 120,
		},
		{
			name:     "invalid string",
			value:    "not-a-number",
			expected: 0,
		},
		{
			name:     "empty string",
			value:    "",
			expected: 0,
		},
		{
			name:     "negative number",
			value:    "-5",
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.value)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
