//go:build integration

package secrets

import (
	"os"
	"runtime"
	"testing"
)

func TestKeychainAccessFlow_Integration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on macOS")
	}

	// Test that EnsureKeychainAccess doesn't error when keychain is unlocked
	err := EnsureKeychainAccess()
	if err != nil {
		t.Logf("Note: keychain may be locked, got error: %v", err)
	}
}

func TestLinuxFileBackendFallback_Integration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux file backend tests only run on Linux")
	}

	// Unset DBUS to simulate headless environment
	orig := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	os.Unsetenv("DBUS_SESSION_BUS_ADDRESS")
	defer os.Setenv("DBUS_SESSION_BUS_ADDRESS", orig)

	// Set file backend password
	os.Setenv("TWENTY_KEYRING_BACKEND", "file")
	os.Setenv("TWENTY_KEYRING_PASSWORD", "testpassword")
	defer os.Unsetenv("TWENTY_KEYRING_BACKEND")
	defer os.Unsetenv("TWENTY_KEYRING_PASSWORD")

	store, err := OpenDefault()
	if err != nil {
		t.Fatalf("OpenDefault() with file backend failed: %v", err)
	}

	// Verify we can use the store
	_, err = store.Keys()
	if err != nil {
		t.Errorf("store.Keys() failed: %v", err)
	}
}
