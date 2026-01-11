package shared

import (
	"os"
	"testing"
)

func TestRequireAuthRuntime_CallsKeychainCheck(t *testing.T) {
	// Skip keychain check when using file backend (expected behavior)
	if os.Getenv("TWENTY_KEYRING_BACKEND") == "file" {
		t.Skip("Keychain check is intentionally skipped for file backend")
	}

	// Save original function
	originalEnsure := ensureKeychainAccessFunc
	defer func() { ensureKeychainAccessFunc = originalEnsure }()

	called := false
	ensureKeychainAccessFunc = func() error {
		called = true
		return nil
	}

	// This will fail because no token is set, but that's after keychain check
	_, _ = RequireAuthRuntime()

	if !called {
		t.Error("RequireAuthRuntime() did not call ensureKeychainAccessFunc")
	}
}

func TestRequireAuthRuntime_SkipsKeychainCheckForFileBackend(t *testing.T) {
	// Save original env
	origBackend := os.Getenv("TWENTY_KEYRING_BACKEND")
	origPassword := os.Getenv("TWENTY_KEYRING_PASSWORD")
	defer func() {
		os.Setenv("TWENTY_KEYRING_BACKEND", origBackend)
		os.Setenv("TWENTY_KEYRING_PASSWORD", origPassword)
	}()

	// Set file backend
	os.Setenv("TWENTY_KEYRING_BACKEND", "file")
	os.Setenv("TWENTY_KEYRING_PASSWORD", "test")

	// Save original function
	originalEnsure := ensureKeychainAccessFunc
	defer func() { ensureKeychainAccessFunc = originalEnsure }()

	called := false
	ensureKeychainAccessFunc = func() error {
		called = true
		return nil
	}

	// This will fail because no token is set, but keychain check should be skipped
	_, _ = RequireAuthRuntime()

	if called {
		t.Error("RequireAuthRuntime() should NOT call ensureKeychainAccessFunc for file backend")
	}
}
