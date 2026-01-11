package auth

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestLogoutCmd_Success(t *testing.T) {
	mock := setupMockStore(t)

	// Add a token to delete
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "test-token"})

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	cmd := logoutCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify token was deleted
	_, err = mock.GetToken("default")
	if err == nil {
		t.Error("Token should have been deleted")
	}
}

func TestLogoutCmd_AlreadyLoggedOut(t *testing.T) {
	setupMockStore(t)

	// No token stored

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	cmd := logoutCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Should succeed without error (already logged out)
}

func TestLogoutCmd_WithProfile(t *testing.T) {
	mock := setupMockStore(t)

	// Add tokens for multiple profiles
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "default-token"})
	_ = mock.SetToken("work", secrets.Token{RefreshToken: "work-token"})

	// Reset viper and set specific profile
	viper.Reset()
	viper.Set("profile", "work")

	cmd := logoutCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify work token was deleted
	_, err = mock.GetToken("work")
	if err == nil {
		t.Error("Work token should have been deleted")
	}

	// Verify default token still exists
	tok, err := mock.GetToken("default")
	if err != nil {
		t.Fatalf("Default token should still exist: %v", err)
	}
	if tok.RefreshToken != "default-token" {
		t.Errorf("Default token mismatch: got %q", tok.RefreshToken)
	}
}

func TestLogoutCmd_EmptyToken(t *testing.T) {
	setupMockStore(t)

	// No token stored - GetToken returns ErrKeyNotFound

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	cmd := logoutCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestLogoutCmd_DeleteError(t *testing.T) {
	setupCustomStore(t, &deleteErrorStore{})

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	cmd := logoutCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("Expected error when DeleteToken fails")
	}
}

// deleteErrorStore is a mock that has a token but fails to delete
type deleteErrorStore struct {
	secrets.MockStore
}

func (m *deleteErrorStore) Keys() ([]string, error) {
	return nil, nil
}

func (m *deleteErrorStore) SetToken(profile string, tok secrets.Token) error {
	return nil
}

func (m *deleteErrorStore) GetToken(profile string) (secrets.Token, error) {
	return secrets.Token{RefreshToken: "token"}, nil
}

func (m *deleteErrorStore) DeleteToken(profile string) error {
	return errors.New("delete error")
}

func (m *deleteErrorStore) ListTokens() ([]secrets.Token, error) {
	return nil, nil
}

func (m *deleteErrorStore) GetDefaultAccount() (string, error) {
	return "", nil
}

func (m *deleteErrorStore) SetDefaultAccount(profile string) error {
	return nil
}
