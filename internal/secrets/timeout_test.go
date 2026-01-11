package secrets

import (
	"errors"
	"testing"
	"time"

	"github.com/99designs/keyring"
)

func TestOpenKeyringWithTimeout_Success(t *testing.T) {
	// Mock a fast keyring open
	originalOpen := keyringOpenFunc
	defer func() { keyringOpenFunc = originalOpen }()

	keyringOpenFunc = func(cfg keyring.Config) (keyring.Keyring, error) {
		return newFakeKeyring(), nil
	}

	ring, err := openKeyringWithTimeout(keyring.Config{}, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("openKeyringWithTimeout() error = %v", err)
	}
	if ring == nil {
		t.Error("openKeyringWithTimeout() returned nil ring")
	}
}

func TestOpenKeyringWithTimeout_Timeout(t *testing.T) {
	originalOpen := keyringOpenFunc
	defer func() { keyringOpenFunc = originalOpen }()

	// Mock a slow keyring open that blocks forever
	keyringOpenFunc = func(cfg keyring.Config) (keyring.Keyring, error) {
		time.Sleep(10 * time.Second)
		return newFakeKeyring(), nil
	}

	_, err := openKeyringWithTimeout(keyring.Config{}, 50*time.Millisecond)
	if err == nil {
		t.Fatal("openKeyringWithTimeout() expected error, got nil")
	}
	if !errors.Is(err, errKeyringTimeout) {
		t.Errorf("openKeyringWithTimeout() error = %v, want errKeyringTimeout", err)
	}
}

func TestOpenKeyringWithTimeout_Error(t *testing.T) {
	originalOpen := keyringOpenFunc
	defer func() { keyringOpenFunc = originalOpen }()

	expectedErr := errors.New("keyring open failed")
	keyringOpenFunc = func(cfg keyring.Config) (keyring.Keyring, error) {
		return nil, expectedErr
	}

	_, err := openKeyringWithTimeout(keyring.Config{}, 100*time.Millisecond)
	if err == nil {
		t.Fatal("openKeyringWithTimeout() expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("openKeyringWithTimeout() error = %v, want %v", err, expectedErr)
	}
}

func TestShouldForceFileBackend(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		backend  string
		dbusAddr string
		expected bool
	}{
		{"linux auto no dbus", "linux", "auto", "", true},
		{"linux auto with dbus", "linux", "auto", "/run/user/1000/bus", false},
		{"linux explicit keychain", "linux", "keychain", "", false},
		{"linux explicit file", "linux", "file", "", false},
		{"darwin auto", "darwin", "auto", "", false},
		{"darwin auto with dbus", "darwin", "auto", "/some/path", false},
		{"windows auto no dbus", "windows", "auto", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := KeyringBackendInfo{Value: tt.backend}
			if got := shouldForceFileBackend(tt.goos, info, tt.dbusAddr); got != tt.expected {
				t.Errorf("shouldForceFileBackend() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldUseKeyringTimeout(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		backend  string
		dbusAddr string
		expected bool
	}{
		{"linux auto with dbus", "linux", "auto", "/run/user/1000/bus", true},
		{"linux auto no dbus", "linux", "auto", "", false},
		{"linux explicit keychain with dbus", "linux", "keychain", "/run/user/1000/bus", false},
		{"linux explicit file with dbus", "linux", "file", "/run/user/1000/bus", false},
		{"darwin auto with dbus", "darwin", "auto", "/some/path", false},
		{"windows auto with dbus", "windows", "auto", "/some/path", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := KeyringBackendInfo{Value: tt.backend}
			if got := shouldUseKeyringTimeout(tt.goos, info, tt.dbusAddr); got != tt.expected {
				t.Errorf("shouldUseKeyringTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKeyringOpenTimeout_Constant(t *testing.T) {
	// Verify the timeout constant is reasonable (not too short, not too long)
	if keyringOpenTimeout < 1*time.Second {
		t.Errorf("keyringOpenTimeout = %v, too short (< 1s)", keyringOpenTimeout)
	}
	if keyringOpenTimeout > 30*time.Second {
		t.Errorf("keyringOpenTimeout = %v, too long (> 30s)", keyringOpenTimeout)
	}
}

func TestErrKeyringTimeout(t *testing.T) {
	if errKeyringTimeout == nil {
		t.Error("errKeyringTimeout should not be nil")
	}
	if errKeyringTimeout.Error() == "" {
		t.Error("errKeyringTimeout should have a message")
	}
}
