package rest

import "fmt"

// ErrUnauthorized indicates authentication failure (401)
type ErrUnauthorized struct {
	Message string
}

func (e *ErrUnauthorized) Error() string {
	return fmt.Sprintf("unauthorized: %s", e.Message)
}

// ErrNotFound indicates resource not found (404)
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e *ErrNotFound) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("not found: %s", e.Resource)
	}
	return fmt.Sprintf("not found: %s with id %s", e.Resource, e.ID)
}

// ErrRateLimited indicates rate limiting (429)
type ErrRateLimited struct {
	RetryAfter int // seconds
}

func (e *ErrRateLimited) Error() string {
	return fmt.Sprintf("rate limited: retry after %ds", e.RetryAfter)
}

// ErrValidation indicates validation error (400)
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}
