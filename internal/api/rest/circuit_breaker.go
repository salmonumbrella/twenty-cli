package rest

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open and not allowing requests
var ErrCircuitOpen = errors.New("circuit breaker open: API unavailable")

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// CircuitClosed allows all requests through (normal operation)
	CircuitClosed CircuitState = iota
	// CircuitOpen blocks all requests
	CircuitOpen
	// CircuitHalfOpen allows a single probe request
	CircuitHalfOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

const (
	// DefaultFailureThreshold is the number of consecutive failures before opening
	DefaultFailureThreshold = 5
	// DefaultCooldownPeriod is how long to wait before allowing a probe request
	DefaultCooldownPeriod = 30 * time.Second
)

// CircuitBreaker implements the circuit breaker pattern for API resilience.
// It tracks consecutive failures and temporarily blocks requests when the
// failure threshold is exceeded, allowing the API time to recover.
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold int           // Number of failures before opening
	cooldownPeriod   time.Duration // Time to wait before half-open

	// State
	state               CircuitState
	consecutiveFailures int
	lastFailureTime     time.Time

	// For testing - allows mocking time.Now()
	now func() time.Time
}

// CircuitBreakerOption configures the CircuitBreaker
type CircuitBreakerOption func(*CircuitBreaker)

// WithFailureThreshold sets the number of consecutive failures before opening
func WithFailureThreshold(n int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		if n > 0 {
			cb.failureThreshold = n
		}
	}
}

// WithCooldownPeriod sets the cooldown period before transitioning to half-open
func WithCooldownPeriod(d time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		if d > 0 {
			cb.cooldownPeriod = d
		}
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given options
func NewCircuitBreaker(opts ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		failureThreshold: DefaultFailureThreshold,
		cooldownPeriod:   DefaultCooldownPeriod,
		state:            CircuitClosed,
		now:              time.Now,
	}
	for _, opt := range opts {
		opt(cb)
	}
	return cb
}

// AllowRequest checks if a request should be allowed through.
// Returns true if the request can proceed, false if it should be blocked.
// In half-open state, it atomically transitions to open if a probe is already in progress.
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if cooldown period has elapsed
		if cb.now().Sub(cb.lastFailureTime) >= cb.cooldownPeriod {
			// Transition to half-open, allow probe request
			cb.state = CircuitHalfOpen
			return true
		}
		return false

	case CircuitHalfOpen:
		// Only one probe at a time - block additional requests
		// The first request through half-open will determine the outcome
		return false
	}

	return false
}

// RecordSuccess records a successful request.
// In half-open state, this closes the circuit.
// In closed state, this resets the failure counter.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFailures = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed request.
// In half-open state, this opens the circuit immediately.
// In closed state, this increments the failure counter and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFailures++
	cb.lastFailureTime = cb.now()

	// In half-open state, any failure opens the circuit
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		return
	}

	// In closed state, check if we've hit the threshold
	if cb.state == CircuitClosed && cb.consecutiveFailures >= cb.failureThreshold {
		cb.state = CircuitOpen
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ConsecutiveFailures returns the current consecutive failure count
func (cb *CircuitBreaker) ConsecutiveFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.consecutiveFailures
}

// Reset resets the circuit breaker to its initial closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.consecutiveFailures = 0
	cb.lastFailureTime = time.Time{}
}
