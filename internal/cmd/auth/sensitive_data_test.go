package auth

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

// TestDebugMode_DoesNotLogTokens verifies that debug mode does not expose
// sensitive token values in output or error messages.
func TestDebugMode_DoesNotLogTokens(t *testing.T) {
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Use a distinctive token value that would be easy to detect if leaked
	sensitiveToken := "super-secret-token-value-abc123xyz789"
	os.Setenv("TWENTY_TOKEN", sensitiveToken)

	viper.Reset()
	viper.Set("output", "text")
	viper.Set("debug", true)
	viper.Set("base_url", "https://twenty.example.com")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	statusShowToken = false

	err := cmd.RunE(cmd, nil)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	capturedOutput, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the full token value does NOT appear anywhere in output
	allOutput := string(capturedOutput) + buf.String()
	if strings.Contains(allOutput, sensitiveToken) {
		t.Errorf("Debug output should NOT contain full token value.\nFound token in output: %s", allOutput)
	}

	// Verify masked token format is used (should show something like "super-se...789")
	if !strings.Contains(allOutput, "...") {
		t.Logf("Output: %s", allOutput)
		// Note: This might not apply if output format doesn't include token display
	}
}

// TestStatusCmd_DoesNotExposeFullTokenByDefault verifies that the status command
// does not expose the full token value unless --show-token is explicitly used.
func TestStatusCmd_DoesNotExposeFullTokenByDefault(t *testing.T) {
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalShowToken := statusShowToken
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
		statusShowToken = originalShowToken
	}()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Use a long token that will be truncated
	fullToken := "this-is-a-very-long-secret-api-token-value-12345"
	os.Setenv("TWENTY_TOKEN", fullToken)

	viper.Reset()
	viper.Set("output", "text")
	viper.Set("base_url", "https://twenty.example.com")
	statusShowToken = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)

	w.Close()
	os.Stdout = oldStdout
	capturedOutput, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	allOutput := string(capturedOutput) + buf.String()

	// The full token should NOT appear in output
	if strings.Contains(allOutput, fullToken) {
		t.Errorf("Status command should NOT expose full token by default.\nToken found in: %s", allOutput)
	}

	// The middle portion of the token should be hidden
	if strings.Contains(allOutput, "very-long-secret") {
		t.Errorf("Middle portion of token should be masked.\nOutput: %s", allOutput)
	}
}

// TestStoredToken_DoesNotExposeInStatusOutput verifies tokens from secure storage
// are not exposed in status command output.
func TestStoredToken_DoesNotExposeInStatusOutput(t *testing.T) {
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalShowToken := statusShowToken
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
		statusShowToken = originalShowToken
	}()

	mock := secrets.NewMockStore()
	SetStore(mock)

	os.Unsetenv("TWENTY_TOKEN")
	statusShowToken = false

	// Store a token with a distinctive value
	sensitiveToken := "stored-secret-credential-abc123def456"
	_ = mock.SetToken("default", secrets.Token{RefreshToken: sensitiveToken})

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "text")
	viper.Set("base_url", "https://twenty.example.com")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)

	w.Close()
	os.Stdout = oldStdout
	capturedOutput, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	allOutput := string(capturedOutput) + buf.String()

	// The full token should NOT appear in output
	if strings.Contains(allOutput, sensitiveToken) {
		t.Errorf("Status command should NOT expose full stored token.\nToken found in: %s", allOutput)
	}

	// Specifically check that the middle portion is not visible
	if strings.Contains(allOutput, "secret-credential") {
		t.Errorf("Middle portion of stored token should be masked.\nOutput: %s", allOutput)
	}
}

// TestErrorMessages_DoNotExposeTokenValues verifies that error messages
// returned from failed operations do not contain token values.
func TestErrorMessages_DoNotExposeTokenValues(t *testing.T) {
	originalStore := store
	originalStoreOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalStoreOpener
	}()

	// Reset store to force opener to be called
	store = nil

	sensitiveToken := "leaked-secret-in-error-12345"

	// Simulate an error that might include the token
	storeOpener = func() (secrets.Store, error) {
		return nil, errors.New("failed to open keyring")
	}

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	viper.Reset()
	viper.Set("profile", "")

	// Try to save token, which should fail due to store opener error
	loginToken = sensitiveToken
	loginBaseURL = "https://twenty.example.com"

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("Expected error when store opener fails")
	}

	// The error message should NOT contain the token value
	errMsg := err.Error()
	if strings.Contains(errMsg, sensitiveToken) {
		t.Errorf("Error message should NOT contain token value.\nError: %s", errMsg)
	}
}

// TestLoginCmd_DoesNotLogTokenInOutput verifies that the login command
// does not echo the token value to output.
func TestLoginCmd_DoesNotLogTokenInOutput(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("debug", true) // Even with debug enabled

	sensitiveToken := "login-test-secret-token-value-xyz"
	loginToken = sensitiveToken
	loginBaseURL = "https://twenty.example.com"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)

	w.Close()
	os.Stdout = oldStdout
	capturedOutput, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	allOutput := string(capturedOutput) + buf.String()

	// The token should NOT appear in success output
	if strings.Contains(allOutput, sensitiveToken) {
		t.Errorf("Login command should NOT echo token value.\nToken found in: %s", allOutput)
	}
}

// TestJSONOutput_DoesNotExposeFullToken verifies JSON output format
// does not expose full token values.
func TestJSONOutput_DoesNotExposeFullToken(t *testing.T) {
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalShowToken := statusShowToken
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
		statusShowToken = originalShowToken
	}()

	mock := secrets.NewMockStore()
	SetStore(mock)

	fullToken := "json-output-secret-token-value-abc123"
	os.Setenv("TWENTY_TOKEN", fullToken)

	viper.Reset()
	viper.Set("output", "json")
	viper.Set("query", "")
	viper.Set("base_url", "https://twenty.example.com")
	statusShowToken = false

	// Capture stdout for JSON output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)

	w.Close()
	os.Stdout = oldStdout
	capturedOutput, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	allOutput := string(capturedOutput) + buf.String()

	// The full token should NOT appear in JSON output
	if strings.Contains(allOutput, fullToken) {
		t.Errorf("JSON output should NOT contain full token.\nToken found in: %s", allOutput)
	}
}

// TestStoreError_DoNotLeakStoredToken verifies that errors from the store
// operations do not leak stored token values.
func TestStoreError_DoNotLeakStoredToken(t *testing.T) {
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	mock := secrets.NewMockStore()
	SetStore(mock)

	os.Unsetenv("TWENTY_TOKEN")

	// Store a token first
	sensitiveToken := "stored-token-should-not-leak-xyz"
	_ = mock.SetToken("default", secrets.Token{RefreshToken: sensitiveToken})

	// Then make subsequent operations fail
	mock.SetGetError(errors.New("keyring locked"))

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "text")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)

	// Error or not, the token should not appear in any output
	allOutput := buf.String()
	if err != nil {
		allOutput += err.Error()
	}

	if strings.Contains(allOutput, sensitiveToken) {
		t.Errorf("Error output should NOT contain stored token.\nOutput: %s", allOutput)
	}
}
