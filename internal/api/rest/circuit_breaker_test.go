package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCircuitBreaker_States(t *testing.T) {
	t.Run("initial state is closed", func(t *testing.T) {
		cb := NewCircuitBreaker()
		if cb.State() != CircuitClosed {
			t.Errorf("expected initial state Closed, got %v", cb.State())
		}
	})

	t.Run("opens after threshold failures", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(3))

		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		if cb.State() != CircuitOpen {
			t.Errorf("expected state Open after 3 failures, got %v", cb.State())
		}
	})

	t.Run("stays closed under threshold", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(5))

		for i := 0; i < 4; i++ {
			cb.RecordFailure()
		}

		if cb.State() != CircuitClosed {
			t.Errorf("expected state Closed with 4 failures (threshold 5), got %v", cb.State())
		}
	})

	t.Run("success resets failure count", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(3))

		cb.RecordFailure()
		cb.RecordFailure()
		cb.RecordSuccess()

		if cb.ConsecutiveFailures() != 0 {
			t.Errorf("expected failure count 0 after success, got %d", cb.ConsecutiveFailures())
		}

		// Should need 3 more failures to open
		cb.RecordFailure()
		cb.RecordFailure()
		if cb.State() != CircuitClosed {
			t.Errorf("expected state Closed, got %v", cb.State())
		}

		cb.RecordFailure()
		if cb.State() != CircuitOpen {
			t.Errorf("expected state Open, got %v", cb.State())
		}
	})
}

func TestCircuitBreaker_AllowRequest(t *testing.T) {
	t.Run("allows requests when closed", func(t *testing.T) {
		cb := NewCircuitBreaker()
		if !cb.AllowRequest() {
			t.Error("expected AllowRequest=true when closed")
		}
	})

	t.Run("blocks requests when open", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(1))
		cb.RecordFailure()

		if cb.State() != CircuitOpen {
			t.Fatalf("expected Open state, got %v", cb.State())
		}

		if cb.AllowRequest() {
			t.Error("expected AllowRequest=false when open")
		}
	})

	t.Run("transitions to half-open after cooldown", func(t *testing.T) {
		now := time.Now()
		cb := NewCircuitBreaker(
			WithFailureThreshold(1),
			WithCooldownPeriod(100*time.Millisecond),
		)
		cb.now = func() time.Time { return now }

		cb.RecordFailure()

		if cb.State() != CircuitOpen {
			t.Fatalf("expected Open state, got %v", cb.State())
		}

		// Still within cooldown
		cb.now = func() time.Time { return now.Add(50 * time.Millisecond) }
		if cb.AllowRequest() {
			t.Error("expected AllowRequest=false during cooldown")
		}

		// After cooldown
		cb.now = func() time.Time { return now.Add(150 * time.Millisecond) }
		if !cb.AllowRequest() {
			t.Error("expected AllowRequest=true after cooldown (half-open probe)")
		}

		if cb.State() != CircuitHalfOpen {
			t.Errorf("expected HalfOpen state, got %v", cb.State())
		}
	})

	t.Run("blocks additional requests in half-open", func(t *testing.T) {
		now := time.Now()
		cb := NewCircuitBreaker(
			WithFailureThreshold(1),
			WithCooldownPeriod(10*time.Millisecond),
		)
		cb.now = func() time.Time { return now }

		cb.RecordFailure()

		// Wait for cooldown
		cb.now = func() time.Time { return now.Add(20 * time.Millisecond) }

		// First request should be allowed (probe)
		if !cb.AllowRequest() {
			t.Error("expected first AllowRequest=true in half-open")
		}

		// Second request should be blocked
		if cb.AllowRequest() {
			t.Error("expected second AllowRequest=false in half-open")
		}
	})
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	t.Run("success in half-open closes circuit", func(t *testing.T) {
		now := time.Now()
		cb := NewCircuitBreaker(
			WithFailureThreshold(1),
			WithCooldownPeriod(10*time.Millisecond),
		)
		cb.now = func() time.Time { return now }

		cb.RecordFailure()

		// Wait for cooldown and trigger half-open
		cb.now = func() time.Time { return now.Add(20 * time.Millisecond) }
		cb.AllowRequest() // Transition to half-open

		cb.RecordSuccess()

		if cb.State() != CircuitClosed {
			t.Errorf("expected Closed state after success in half-open, got %v", cb.State())
		}
		if cb.ConsecutiveFailures() != 0 {
			t.Errorf("expected 0 failures after success, got %d", cb.ConsecutiveFailures())
		}
	})

	t.Run("failure in half-open opens circuit", func(t *testing.T) {
		now := time.Now()
		cb := NewCircuitBreaker(
			WithFailureThreshold(1),
			WithCooldownPeriod(10*time.Millisecond),
		)
		cb.now = func() time.Time { return now }

		cb.RecordFailure()

		// Wait for cooldown and trigger half-open
		cb.now = func() time.Time { return now.Add(20 * time.Millisecond) }
		cb.AllowRequest()

		cb.RecordFailure()

		if cb.State() != CircuitOpen {
			t.Errorf("expected Open state after failure in half-open, got %v", cb.State())
		}
	})
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(WithFailureThreshold(1))

	// Open the circuit
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open state, got %v", cb.State())
	}

	// Reset
	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Errorf("expected Closed state after reset, got %v", cb.State())
	}
	if cb.ConsecutiveFailures() != 0 {
		t.Errorf("expected 0 failures after reset, got %d", cb.ConsecutiveFailures())
	}
}

func TestCircuitBreaker_Concurrency(t *testing.T) {
	cb := NewCircuitBreaker(WithFailureThreshold(100))

	var wg sync.WaitGroup
	failures := 50
	successes := 50

	// Concurrent failures
	for i := 0; i < failures; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.RecordFailure()
		}()
	}

	// Concurrent successes (interspersed)
	for i := 0; i < successes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.RecordSuccess()
		}()
	}

	// Concurrent AllowRequest calls
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.AllowRequest()
		}()
	}

	wg.Wait()

	// Just verify no panic occurred and state is valid
	state := cb.State()
	if state != CircuitClosed && state != CircuitOpen && state != CircuitHalfOpen {
		t.Errorf("invalid state after concurrent operations: %v", state)
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("CircuitState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestCircuitBreaker_Options(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cb := NewCircuitBreaker()
		// Verify defaults by behavior
		for i := 0; i < DefaultFailureThreshold-1; i++ {
			cb.RecordFailure()
		}
		if cb.State() != CircuitClosed {
			t.Error("expected closed before reaching default threshold")
		}
		cb.RecordFailure()
		if cb.State() != CircuitOpen {
			t.Error("expected open at default threshold")
		}
	})

	t.Run("custom failure threshold", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(2))
		cb.RecordFailure()
		if cb.State() != CircuitClosed {
			t.Error("expected closed after 1 failure")
		}
		cb.RecordFailure()
		if cb.State() != CircuitOpen {
			t.Error("expected open after 2 failures")
		}
	})

	t.Run("ignores invalid threshold", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(0))
		// Should use default
		for i := 0; i < DefaultFailureThreshold; i++ {
			cb.RecordFailure()
		}
		if cb.State() != CircuitOpen {
			t.Error("expected open at default threshold when 0 provided")
		}
	})

	t.Run("ignores invalid cooldown", func(t *testing.T) {
		cb := NewCircuitBreaker(WithCooldownPeriod(0))
		// Should use default - hard to test directly but verify no panic
		if cb.State() != CircuitClosed {
			t.Error("expected initial closed state")
		}
	})
}

// ==============================================================================
// Client Integration Tests for Circuit Breaker
// ==============================================================================

func TestClient_CircuitBreakerIntegration(t *testing.T) {
	t.Run("opens after consecutive failures", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"server error"}}`))
		}))
		defer server.Close()

		cb := NewCircuitBreaker(WithFailureThreshold(3))
		client := NewClient(server.URL, "token", false, WithNoRetry(), WithCircuitBreaker(cb))

		// Make 3 failing requests
		for i := 0; i < 3; i++ {
			var result map[string]interface{}
			err := client.Get(context.Background(), "/test", &result)
			if err == nil {
				t.Errorf("request %d: expected error, got nil", i+1)
			}
		}

		// Circuit should be open now
		if cb.State() != CircuitOpen {
			t.Errorf("expected Open state after 3 failures, got %v", cb.State())
		}

		// Next request should be blocked by circuit breaker
		var result map[string]interface{}
		err := client.Get(context.Background(), "/test", &result)
		if err != ErrCircuitOpen {
			t.Errorf("expected ErrCircuitOpen, got %v", err)
		}

		// Server should only have received 3 requests
		if atomic.LoadInt32(&attempts) != 3 {
			t.Errorf("expected 3 server requests, got %d", attempts)
		}
	})

	t.Run("recovers after cooldown with successful probe", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attempts, 1)
			if count <= 3 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":{"message":"server error"}}`))
				return
			}
			// After 3 failures, server recovers
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		now := time.Now()
		cb := NewCircuitBreaker(
			WithFailureThreshold(3),
			WithCooldownPeriod(50*time.Millisecond),
		)
		cb.now = func() time.Time { return now }

		client := NewClient(server.URL, "token", false, WithNoRetry(), WithCircuitBreaker(cb))

		// Make 3 failing requests
		for i := 0; i < 3; i++ {
			var result map[string]interface{}
			client.Get(context.Background(), "/test", &result)
		}

		// Advance time past cooldown
		cb.now = func() time.Time { return now.Add(100 * time.Millisecond) }

		// Probe request should succeed
		var result map[string]interface{}
		err := client.Get(context.Background(), "/test", &result)
		if err != nil {
			t.Errorf("expected success on probe, got %v", err)
		}

		// Circuit should be closed now
		if cb.State() != CircuitClosed {
			t.Errorf("expected Closed state after successful probe, got %v", cb.State())
		}
	})

	t.Run("does not count 4xx as circuit breaker failures", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"not found"}}`))
		}))
		defer server.Close()

		cb := NewCircuitBreaker(WithFailureThreshold(2))
		client := NewClient(server.URL, "token", false, WithNoRetry(), WithCircuitBreaker(cb))

		// Make multiple 404 requests
		for i := 0; i < 5; i++ {
			var result map[string]interface{}
			client.Get(context.Background(), "/test", &result)
		}

		// Circuit should still be closed (4xx are not failures)
		if cb.State() != CircuitClosed {
			t.Errorf("expected Closed state (4xx should not count as failures), got %v", cb.State())
		}
		if cb.ConsecutiveFailures() != 0 {
			t.Errorf("expected 0 consecutive failures, got %d", cb.ConsecutiveFailures())
		}
	})

	t.Run("counts 429 rate limiting as failures", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"rate limited"}}`))
		}))
		defer server.Close()

		cb := NewCircuitBreaker(WithFailureThreshold(2))
		client := NewClient(server.URL, "token", false, WithNoRetry(), WithCircuitBreaker(cb))

		// Make 2 rate-limited requests
		for i := 0; i < 2; i++ {
			var result map[string]interface{}
			client.Get(context.Background(), "/test", &result)
		}

		// Circuit should be open (rate limiting counts as failure)
		if cb.State() != CircuitOpen {
			t.Errorf("expected Open state after rate limiting, got %v", cb.State())
		}
	})

	t.Run("success resets consecutive failures", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attempts, 1)
			if count%3 == 0 {
				// Every 3rd request succeeds
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"error"}}`))
		}))
		defer server.Close()

		cb := NewCircuitBreaker(WithFailureThreshold(3))
		client := NewClient(server.URL, "token", false, WithNoRetry(), WithCircuitBreaker(cb))

		// Fail, fail, success pattern - should never reach threshold
		for i := 0; i < 9; i++ {
			var result map[string]interface{}
			client.Get(context.Background(), "/test", &result)
		}

		// Circuit should still be closed
		if cb.State() != CircuitClosed {
			t.Errorf("expected Closed state with intermittent successes, got %v", cb.State())
		}
	})
}

func TestClient_CircuitBreakerDisabled(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"server error"}}`))
	}))
	defer server.Close()

	// Client without circuit breaker
	client := NewClient(server.URL, "token", false, WithNoRetry())

	// Make many failing requests
	for i := 0; i < 10; i++ {
		var result map[string]interface{}
		client.Get(context.Background(), "/test", &result)
	}

	// All requests should reach server
	if atomic.LoadInt32(&attempts) != 10 {
		t.Errorf("expected 10 server requests without circuit breaker, got %d", attempts)
	}
}

func TestClient_CircuitBreakerEnvironmentVariable(t *testing.T) {
	t.Run("enabled via env var true", func(t *testing.T) {
		os.Setenv("TWENTY_CIRCUIT_BREAKER", "true")
		defer os.Unsetenv("TWENTY_CIRCUIT_BREAKER")

		client := NewClient("http://example.com", "token", false)
		if client.circuitBreaker == nil {
			t.Error("expected circuit breaker to be enabled when TWENTY_CIRCUIT_BREAKER=true")
		}
	})

	t.Run("enabled via env var 1", func(t *testing.T) {
		os.Setenv("TWENTY_CIRCUIT_BREAKER", "1")
		defer os.Unsetenv("TWENTY_CIRCUIT_BREAKER")

		client := NewClient("http://example.com", "token", false)
		if client.circuitBreaker == nil {
			t.Error("expected circuit breaker to be enabled when TWENTY_CIRCUIT_BREAKER=1")
		}
	})

	t.Run("disabled by default", func(t *testing.T) {
		os.Unsetenv("TWENTY_CIRCUIT_BREAKER")

		client := NewClient("http://example.com", "token", false)
		if client.circuitBreaker != nil {
			t.Error("expected circuit breaker to be nil when env var not set")
		}
	})

	t.Run("disabled with false", func(t *testing.T) {
		os.Setenv("TWENTY_CIRCUIT_BREAKER", "false")
		defer os.Unsetenv("TWENTY_CIRCUIT_BREAKER")

		client := NewClient("http://example.com", "token", false)
		if client.circuitBreaker != nil {
			t.Error("expected circuit breaker to be nil when TWENTY_CIRCUIT_BREAKER=false")
		}
	})
}

func TestClient_CircuitBreakerWithRetries(t *testing.T) {
	t.Run("retries count as single failure when all fail", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":{"message":"service unavailable"}}`))
		}))
		defer server.Close()

		cb := NewCircuitBreaker(WithFailureThreshold(2))
		// Enable retries
		client := NewClient(server.URL, "token", false, WithMaxRetries(2), WithCircuitBreaker(cb))

		// First request with retries (3 attempts)
		var result map[string]interface{}
		err := client.Get(context.Background(), "/test", &result)
		if err == nil {
			t.Error("expected error, got nil")
		}

		// Should count as 1 failure (the whole retry sequence)
		if cb.ConsecutiveFailures() != 1 {
			t.Errorf("expected 1 consecutive failure after retry sequence, got %d", cb.ConsecutiveFailures())
		}

		// Second request with retries (3 more attempts)
		client.Get(context.Background(), "/test", &result)

		// Now circuit should be open
		if cb.State() != CircuitOpen {
			t.Errorf("expected Open state after 2 failed retry sequences, got %v", cb.State())
		}

		// Total server attempts: 3 (first) + 3 (second) = 6
		if atomic.LoadInt32(&attempts) != 6 {
			t.Errorf("expected 6 total server attempts, got %d", attempts)
		}
	})
}

func TestErrCircuitOpen(t *testing.T) {
	err := ErrCircuitOpen
	expected := "circuit breaker open: API unavailable"
	if err.Error() != expected {
		t.Errorf("ErrCircuitOpen.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestClient_IsCircuitBreakerFailure(t *testing.T) {
	client := NewClient("http://example.com", "token", false)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "rate limited error",
			err:      &ErrRateLimited{RetryAfter: 10},
			expected: true,
		},
		{
			name:     "500 API error",
			err:      &APIError{StatusCode: 500, Message: "server error"},
			expected: true,
		},
		{
			name:     "502 API error",
			err:      &APIError{StatusCode: 502, Message: "bad gateway"},
			expected: true,
		},
		{
			name:     "400 API error",
			err:      &APIError{StatusCode: 400, Message: "bad request"},
			expected: false,
		},
		{
			name:     "404 API error",
			err:      &APIError{StatusCode: 404, Message: "not found"},
			expected: false,
		},
		{
			name:     "401 unauthorized",
			err:      &ErrUnauthorized{Message: "invalid token"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isCircuitBreakerFailure(tt.err)
			if result != tt.expected {
				t.Errorf("isCircuitBreakerFailure(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
