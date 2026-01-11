package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultMaxRetries is the default number of retries for rate-limited requests
	DefaultMaxRetries = 3
	// DefaultBaseDelay is the initial backoff delay
	DefaultBaseDelay = 1 * time.Second
	// DefaultMaxDelay is the maximum backoff delay
	DefaultMaxDelay = 30 * time.Second
)

type Client struct {
	baseURL        string
	token          string
	httpClient     *http.Client
	debug          bool
	noRetry        bool
	maxRetries     int
	circuitBreaker *CircuitBreaker
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithNoRetry disables automatic retry on rate limiting
func WithNoRetry() ClientOption {
	return func(c *Client) {
		c.noRetry = true
	}
}

// WithMaxRetries sets the maximum number of retries (default: 3)
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithCircuitBreaker enables the circuit breaker with the given configuration
func WithCircuitBreaker(cb *CircuitBreaker) ClientOption {
	return func(c *Client) {
		c.circuitBreaker = cb
	}
}

func NewClient(baseURL, token string, debug bool, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug:      debug,
		maxRetries: DefaultMaxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}

	// Initialize circuit breaker from environment variable if not already set
	if c.circuitBreaker == nil {
		if enabled := os.Getenv("TWENTY_CIRCUIT_BREAKER"); enabled == "true" || enabled == "1" {
			c.circuitBreaker = NewCircuitBreaker()
		}
	}

	return c
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	// Check circuit breaker before making request
	if c.circuitBreaker != nil {
		if !c.circuitBreaker.AllowRequest() {
			if c.debug {
				fmt.Printf("[DEBUG] Circuit breaker is %s, blocking request\n", c.circuitBreaker.State())
			}
			return ErrCircuitOpen
		}
		if c.debug {
			fmt.Printf("[DEBUG] Circuit breaker state: %s\n", c.circuitBreaker.State())
		}
	}

	err := c.doInternal(ctx, method, path, body, result)

	// Record result in circuit breaker
	if c.circuitBreaker != nil {
		if err != nil && c.isCircuitBreakerFailure(err) {
			c.circuitBreaker.RecordFailure()
			if c.debug {
				fmt.Printf("[DEBUG] Circuit breaker recorded failure (consecutive: %d)\n",
					c.circuitBreaker.ConsecutiveFailures())
			}
		} else if err == nil {
			c.circuitBreaker.RecordSuccess()
			if c.debug && c.circuitBreaker.State() == CircuitClosed {
				fmt.Printf("[DEBUG] Circuit breaker recorded success\n")
			}
		}
	}

	return err
}

// isCircuitBreakerFailure determines if an error should count as a circuit breaker failure.
// Network errors and server errors (5xx) are considered failures.
// Client errors (4xx) are not counted as failures since they indicate the API is reachable.
func (c *Client) isCircuitBreakerFailure(err error) bool {
	if err == nil {
		return false
	}

	// Rate limiting is a transient failure
	var rateLimited *ErrRateLimited
	if errors.As(err, &rateLimited) {
		return true
	}

	// Check for API errors with 5xx status codes
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}

	// Network errors and retries exceeded errors are failures
	errStr := err.Error()
	if strings.Contains(errStr, "request failed:") ||
		strings.Contains(errStr, "max retries") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") {
		return true
	}

	return false
}

// doInternal performs the actual HTTP request with retries
func (c *Client) doInternal(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	// Parse base URL and append path manually to preserve query strings
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	// Combine base path with request path, handling query strings
	u := base.Scheme + "://" + base.Host + path

	// Marshal body once if present (for potential retries)
	var jsonBody []byte
	if body != nil {
		var err error
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	var lastErr error
	var lastResp *http.Response
	maxAttempts := 1
	if !c.noRetry {
		maxAttempts = c.maxRetries + 1 // initial + retries
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Calculate backoff delay (use lastResp for Retry-After header)
			delay := c.calculateBackoff(attempt, lastResp)
			if c.debug {
				fmt.Printf("[DEBUG] Retry %d/%d after %v\n", attempt, c.maxRetries, delay)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		var bodyReader io.Reader
		if jsonBody != nil {
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		if c.debug {
			fmt.Printf("[DEBUG] %s %s\n", method, u)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			lastResp = nil
			continue // Retry on network errors
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			lastResp = resp
			continue
		}

		// Sanitize control characters that break JSON parsing
		respBody = sanitizeJSON(respBody)

		if c.debug {
			fmt.Printf("[DEBUG] Response: %d %s\n", resp.StatusCode, string(respBody))
		}

		// Check if we should retry based on status code
		retryAfterHeader := resp.Header.Get("Retry-After")
		if c.shouldRetry(resp.StatusCode) && !c.noRetry && attempt < maxAttempts-1 {
			if c.debug && retryAfterHeader != "" {
				fmt.Printf("[DEBUG] Retry-After header: %s\n", retryAfterHeader)
			}
			lastErr = parseAPIError(resp.StatusCode, respBody, retryAfterHeader)
			lastResp = resp
			continue
		}

		if resp.StatusCode >= 400 {
			return parseAPIError(resp.StatusCode, respBody, retryAfterHeader)
		}

		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", c.maxRetries, lastErr)
}

// shouldRetry returns true if the status code indicates a retryable error
func (c *Client) shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout,     // 504
		http.StatusBadGateway:         // 502
		return true
	default:
		return false
	}
}

// calculateBackoff returns the delay before the next retry attempt
// It uses exponential backoff with optional Retry-After header support
func (c *Client) calculateBackoff(attempt int, resp *http.Response) time.Duration {
	// Check for Retry-After header first
	if resp != nil {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			// Try parsing as seconds
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				delay := time.Duration(seconds) * time.Second
				if delay > DefaultMaxDelay {
					delay = DefaultMaxDelay
				}
				return delay
			}
			// Try parsing as HTTP date (RFC1123)
			if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
				delay := time.Until(t)
				if delay < 0 {
					delay = DefaultBaseDelay
				}
				if delay > DefaultMaxDelay {
					delay = DefaultMaxDelay
				}
				return delay
			}
		}
	}

	// Exponential backoff: baseDelay * 2^attempt
	delay := DefaultBaseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
	if delay > DefaultMaxDelay {
		delay = DefaultMaxDelay
	}
	return delay
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

func (c *Client) Patch(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPatch, path, body, result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// DoRaw executes a request and returns the raw JSON response.
func (c *Client) DoRaw(ctx context.Context, method, path string, body interface{}) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.do(ctx, method, path, body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// sanitizeJSON escapes control characters (0x00-0x1F) inside JSON strings
// that can appear in API responses and break JSON parsing.
// Control characters outside strings (structural newlines/tabs) are preserved.
func sanitizeJSON(data []byte) []byte {
	result := make([]byte, 0, len(data)+128) // preallocate with some extra space
	inString := false
	escaped := false

	for _, b := range data {
		if escaped {
			// Previous char was backslash, this is an escaped char
			result = append(result, b)
			escaped = false
			continue
		}

		if b == '\\' && inString {
			// Start of escape sequence
			result = append(result, b)
			escaped = true
			continue
		}

		if b == '"' {
			// Toggle string context
			inString = !inString
			result = append(result, b)
			continue
		}

		if inString && b < 0x20 {
			// Control character inside string - escape it
			switch b {
			case '\t':
				result = append(result, '\\', 't')
			case '\n':
				result = append(result, '\\', 'n')
			case '\r':
				result = append(result, '\\', 'r')
			default:
				// Other control chars: use \u00XX format
				result = append(result, []byte(fmt.Sprintf("\\u%04x", b))...)
			}
		} else {
			// Regular character or structural whitespace
			result = append(result, b)
		}
	}
	return result
}
