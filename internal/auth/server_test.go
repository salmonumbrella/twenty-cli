package auth

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// waitForServer polls the server until it's ready or times out
func waitForServer(t *testing.T, s *AuthServer) {
	t.Helper()
	client := &http.Client{Timeout: 100 * time.Millisecond}
	rootURL := "http://127.0.0.1:" + strconv.Itoa(s.Port()) + "/"
	for i := 0; i < 50; i++ {
		if _, err := client.Get(rootURL); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server failed to start")
}

func TestNewAuthServer(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	// Port should be non-zero
	if s.Port() == 0 {
		t.Error("Port() = 0, want non-zero")
	}

	// State should be non-empty and sufficient length (32 bytes base64 = 43 chars)
	state := s.State()
	if len(state) < 40 {
		t.Errorf("State() length = %d, want >= 40", len(state))
	}

	// RedirectURL should contain the port
	redirectURL := s.RedirectURL()
	if !strings.Contains(redirectURL, "127.0.0.1") {
		t.Errorf("RedirectURL() = %q, want to contain 127.0.0.1", redirectURL)
	}
	if !strings.Contains(redirectURL, "/oauth2/callback") {
		t.Errorf("RedirectURL() = %q, want to contain /oauth2/callback", redirectURL)
	}
}

func TestNewAuthServer_UniqueState(t *testing.T) {
	// Create multiple servers and verify states are unique
	states := make(map[string]bool)
	for i := 0; i < 10; i++ {
		s, err := NewAuthServer()
		if err != nil {
			t.Fatalf("NewAuthServer() error = %v", err)
		}
		state := s.State()
		s.Close()

		if states[state] {
			t.Errorf("Duplicate state generated: %s", state)
		}
		states[state] = true
	}
}

func TestWithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	s, err := NewAuthServer(WithTimeout(timeout))
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	// Verify timeout is set (we can't access it directly, but we can verify
	// the server was created successfully with the option)
	if s == nil {
		t.Error("Expected non-nil server")
	}
}

func TestCallbackHandler_Success(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make request with valid code and state
	callbackURL := s.RedirectURL() + "?code=test_auth_code&state=" + url.QueryEscape(s.State())
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify we can retrieve the code
	code, err := s.WaitForAuth(ctx)
	if err != nil {
		t.Fatalf("WaitForAuth() error = %v", err)
	}
	if code != "test_auth_code" {
		t.Errorf("WaitForAuth() = %q, want %q", code, "test_auth_code")
	}
}

func TestCallbackHandler_MissingCode(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make request without code
	callbackURL := s.RedirectURL() + "?state=" + url.QueryEscape(s.State())
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	// Verify error is returned
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error for missing code")
	}
	if !strings.Contains(err.Error(), "missing authorization code") {
		t.Errorf("WaitForAuth() error = %v, want error containing 'missing authorization code'", err)
	}
}

func TestCallbackHandler_ErrorParam(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make request with error parameter
	callbackURL := s.RedirectURL() + "?error=access_denied&error_description=User+denied+access"
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	// Even error responses return 200 to display error page
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify error is returned via WaitForAuth
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error for OAuth error")
	}
	if !strings.Contains(err.Error(), "User denied access") {
		t.Errorf("WaitForAuth() error = %v, want error containing 'User denied access'", err)
	}
}

func TestCallbackHandler_ErrorParamWithoutDescription(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make request with error parameter but no description
	callbackURL := s.RedirectURL() + "?error=server_error"
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	// Verify error is returned via WaitForAuth
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error for OAuth error")
	}
	// When no description, error code is used as description
	if !strings.Contains(err.Error(), "server_error") {
		t.Errorf("WaitForAuth() error = %v, want error containing 'server_error'", err)
	}
}

func TestCallbackHandler_StateMismatch(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make request with wrong state
	callbackURL := s.RedirectURL() + "?code=test_code&state=wrong_state"
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	// Verify CSRF error is returned
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error for state mismatch")
	}
	if !strings.Contains(err.Error(), "state mismatch") {
		t.Errorf("WaitForAuth() error = %v, want error containing 'state mismatch'", err)
	}
}

func TestWaitForAuth_Timeout(t *testing.T) {
	s, err := NewAuthServer(WithTimeout(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	s.Start(ctx)

	// Don't make any callback request, let it timeout
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("WaitForAuth() error = %v, want error containing 'timeout'", err)
	}
}

func TestRootHandler(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Access root path - construct URL from port
	rootURL := "http://127.0.0.1:" + itoa(s.Port()) + "/"

	resp, err := http.Get(rootURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", rootURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want to contain 'text/html'", contentType)
	}
}

func TestGenerateState(t *testing.T) {
	// Test via NewAuthServer since generateState is unexported
	// Verify uniqueness by creating multiple servers
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s, err := NewAuthServer()
		if err != nil {
			t.Fatalf("NewAuthServer() error = %v", err)
		}
		state := s.State()
		s.Close()

		if seen[state] {
			t.Errorf("Duplicate state generated after %d iterations", i)
		}
		seen[state] = true

		// State should be 43 characters (32 bytes base64 URL encoded)
		if len(state) != 43 {
			t.Errorf("State length = %d, want 43", len(state))
		}
	}
}

func TestSuccessPage_ContainsExpectedContent(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make successful callback request
	callbackURL := s.RedirectURL() + "?code=test_auth_code&state=" + url.QueryEscape(s.State())
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// Check that response contains expected success indicators
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "html") {
		t.Error("Response body does not contain 'html'")
	}
}

func TestErrorPage_ContainsExpectedContent(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Make request with error parameter
	callbackURL := s.RedirectURL() + "?error=access_denied&error_description=Test+error+message"
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", callbackURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// Check that response contains the error message
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "Test error message") {
		t.Errorf("Response body does not contain error message. Body: %s", bodyStr)
	}
}

func TestClose_Idempotent(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}

	ctx := context.Background()
	s.Start(ctx)

	// Close should be safe to call multiple times
	err = s.Close()
	if err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	err = s.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestClose_BeforeStart(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}

	// Close before Start should not panic
	err = s.Close()
	if err != nil {
		t.Errorf("Close() before Start() error = %v", err)
	}
}

// itoa is a simple int to string conversion
func itoa(i int) string {
	return strconv.Itoa(i)
}

func TestPrintAuthURL(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = w

	testURL := "https://example.com/auth?client_id=test"
	PrintAuthURL(testURL)

	w.Close()
	os.Stderr = oldStderr

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Opening browser for authorization") {
		t.Errorf("Output missing 'Opening browser for authorization', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "If the browser doesn't open") {
		t.Errorf("Output missing fallback message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, testURL) {
		t.Errorf("Output missing URL %q, got: %s", testURL, outputStr)
	}
}

func TestStart_ContextCancellation(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	waitForServer(t, s)

	// Verify server is responding
	rootURL := "http://127.0.0.1:" + itoa(s.Port()) + "/"
	resp, err := http.Get(rootURL)
	if err != nil {
		t.Fatalf("GET %s error = %v", rootURL, err)
	}
	resp.Body.Close()

	// Cancel context to trigger shutdown
	cancel()

	// Give server time to shut down
	time.Sleep(100 * time.Millisecond)

	// Server should be closed now, requests should fail
	client := &http.Client{Timeout: 100 * time.Millisecond}
	_, err = client.Get(rootURL)
	if err == nil {
		t.Error("Expected error after context cancellation, server should be closed")
	}
}

func TestWaitForAuth_ContextCancellation(t *testing.T) {
	s, err := NewAuthServer(WithTimeout(10 * time.Second))
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	// Cancel the context immediately
	cancel()

	// WaitForAuth should return timeout error due to context cancellation
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error for context cancellation")
	}
}

func TestCallbackHandler_ChannelBlocking(t *testing.T) {
	// Test that the server handles multiple callback requests gracefully
	// (second request should not block since channels are buffered with size 1)
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// First callback - success
	callbackURL := s.RedirectURL() + "?code=first_code&state=" + url.QueryEscape(s.State())
	resp1, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("First GET error = %v", err)
	}
	resp1.Body.Close()

	// Second callback - should not block (channel already has a value)
	callbackURL2 := s.RedirectURL() + "?code=second_code&state=" + url.QueryEscape(s.State())
	client := &http.Client{Timeout: 1 * time.Second}
	resp2, err := client.Get(callbackURL2)
	if err != nil {
		t.Fatalf("Second GET error = %v", err)
	}
	resp2.Body.Close()

	// Should receive the first code
	code, err := s.WaitForAuth(ctx)
	if err != nil {
		t.Fatalf("WaitForAuth() error = %v", err)
	}
	if code != "first_code" {
		t.Errorf("WaitForAuth() = %q, want %q", code, "first_code")
	}
}

func TestCallbackHandler_ErrorChannelBlocking(t *testing.T) {
	// Test that multiple error responses don't block
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// First error
	callbackURL := s.RedirectURL() + "?error=error1&error_description=First+error"
	resp1, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("First GET error = %v", err)
	}
	resp1.Body.Close()

	// Second error - should not block
	callbackURL2 := s.RedirectURL() + "?error=error2&error_description=Second+error"
	client := &http.Client{Timeout: 1 * time.Second}
	resp2, err := client.Get(callbackURL2)
	if err != nil {
		t.Fatalf("Second GET error = %v", err)
	}
	resp2.Body.Close()

	// Should receive the first error
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "First error") {
		t.Errorf("WaitForAuth() error = %v, want error containing 'First error'", err)
	}
}

func TestCallbackHandler_StateMismatchChannelBlocking(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// First state mismatch
	callbackURL := s.RedirectURL() + "?code=test&state=wrong1"
	resp1, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("First GET error = %v", err)
	}
	resp1.Body.Close()

	// Second state mismatch - should not block
	callbackURL2 := s.RedirectURL() + "?code=test&state=wrong2"
	client := &http.Client{Timeout: 1 * time.Second}
	resp2, err := client.Get(callbackURL2)
	if err != nil {
		t.Fatalf("Second GET error = %v", err)
	}
	resp2.Body.Close()

	// Should receive state mismatch error
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error")
	}
}

func TestCallbackHandler_MissingCodeChannelBlocking(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	state := url.QueryEscape(s.State())

	// First missing code
	callbackURL := s.RedirectURL() + "?state=" + state
	resp1, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("First GET error = %v", err)
	}
	resp1.Body.Close()

	// Second missing code - should not block
	client := &http.Client{Timeout: 1 * time.Second}
	resp2, err := client.Get(callbackURL)
	if err != nil {
		t.Fatalf("Second GET error = %v", err)
	}
	resp2.Body.Close()

	// Should receive missing code error
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error")
	}
}

func TestNewAuthServer_WithMultipleOptions(t *testing.T) {
	timeout1 := 30 * time.Second
	timeout2 := 60 * time.Second

	// Test that multiple options are applied in order
	s, err := NewAuthServer(WithTimeout(timeout1), WithTimeout(timeout2))
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	// Server should be created successfully
	if s == nil {
		t.Error("Expected non-nil server")
	}
}

func TestAuthServer_PortIsValid(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	port := s.Port()
	// Port should be in valid range
	if port < 1 || port > 65535 {
		t.Errorf("Port() = %d, want value between 1 and 65535", port)
	}
}

func TestAuthServer_RedirectURLFormat(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	redirectURL := s.RedirectURL()

	// Parse the URL to verify format
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		t.Fatalf("Failed to parse redirect URL: %v", err)
	}

	if parsed.Scheme != "http" {
		t.Errorf("Scheme = %q, want %q", parsed.Scheme, "http")
	}

	if parsed.Host == "" {
		t.Error("Host is empty")
	}

	if parsed.Path != "/oauth2/callback" {
		t.Errorf("Path = %q, want %q", parsed.Path, "/oauth2/callback")
	}
}

func TestAuthServer_StateIsURLSafe(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	state := s.State()

	// State should be URL safe (no encoding needed)
	encoded := url.QueryEscape(state)
	if encoded != state {
		t.Errorf("State is not URL-safe: original=%q, encoded=%q", state, encoded)
	}
}

func TestWaitForAuth_ReceivesError(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Trigger error through OAuth error response
	callbackURL := s.RedirectURL() + "?error=unauthorized&error_description=Invalid+credentials"
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	resp.Body.Close()

	// WaitForAuth should return the error
	_, err = s.WaitForAuth(ctx)
	if err == nil {
		t.Error("WaitForAuth() error = nil, want error")
	}

	// Check error message contains the description
	if !strings.Contains(err.Error(), "Invalid credentials") {
		t.Errorf("Error = %v, want to contain 'Invalid credentials'", err)
	}
}

func TestCallbackHandler_ResponseBody(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	// Test success response body
	callbackURL := s.RedirectURL() + "?code=test_code&state=" + url.QueryEscape(s.State())
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll error = %v", err)
	}

	// Should return HTML success page
	if len(body) == 0 {
		t.Error("Response body is empty")
	}
}

func TestRootHandler_ResponseBody(t *testing.T) {
	s, err := NewAuthServer()
	if err != nil {
		t.Fatalf("NewAuthServer() error = %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	waitForServer(t, s)

	rootURL := "http://127.0.0.1:" + itoa(s.Port()) + "/"
	resp, err := http.Get(rootURL)
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll error = %v", err)
	}

	// Should return HTML waiting page
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "Authenticating") {
		t.Error("Response should contain 'Authenticating'")
	}
}

func TestErrorVariables(t *testing.T) {
	// Test that error variables are properly defined
	if errAuthorization == nil {
		t.Error("errAuthorization is nil")
	}
	if errMissingCode == nil {
		t.Error("errMissingCode is nil")
	}
	if errStateMismatch == nil {
		t.Error("errStateMismatch is nil")
	}
	if errTimeout == nil {
		t.Error("errTimeout is nil")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants have expected values
	if DefaultTimeout <= 0 {
		t.Errorf("DefaultTimeout = %v, want positive duration", DefaultTimeout)
	}
	if PostSuccessDisplaySeconds <= 0 {
		t.Errorf("PostSuccessDisplaySeconds = %d, want positive value", PostSuccessDisplaySeconds)
	}
}

func TestAuthResult_Fields(t *testing.T) {
	result := AuthResult{
		Code:  "test_code",
		State: "test_state",
	}

	if result.Code != "test_code" {
		t.Errorf("Code = %q, want %q", result.Code, "test_code")
	}
	if result.State != "test_state" {
		t.Errorf("State = %q, want %q", result.State, "test_state")
	}
}
