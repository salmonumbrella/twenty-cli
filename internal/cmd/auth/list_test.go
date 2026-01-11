package auth

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestListCmd_Empty(t *testing.T) {
	setupMockStore(t)

	viper.Reset()
	viper.Set("output", "text")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_WithProfiles(t *testing.T) {
	mock := setupMockStore(t)

	// Add some tokens
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "token1"})
	_ = mock.SetToken("work", secrets.Token{RefreshToken: "token2"})
	_ = mock.SetToken("personal", secrets.Token{RefreshToken: "token3"})

	viper.Reset()
	viper.Set("output", "text")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_WithDefaultProfile(t *testing.T) {
	mock := setupMockStore(t)

	// Add some tokens
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "token1"})
	_ = mock.SetToken("work", secrets.Token{RefreshToken: "token2"})
	_ = mock.SetDefaultAccount("work")

	viper.Reset()
	viper.Set("output", "text")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_JSON(t *testing.T) {
	mock := setupMockStore(t)

	// Add some tokens
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "token1"})
	_ = mock.SetToken("work", secrets.Token{RefreshToken: "token2"})
	_ = mock.SetDefaultAccount("default")

	viper.Reset()
	viper.Set("output", "json")
	viper.Set("query", "")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_YAML(t *testing.T) {
	mock := setupMockStore(t)

	// Add some tokens
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "token1"})

	viper.Reset()
	viper.Set("output", "yaml")
	viper.Set("query", "")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_CSV(t *testing.T) {
	mock := setupMockStore(t)

	// Add some tokens
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "token1"})

	viper.Reset()
	viper.Set("output", "csv")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_NoDefaultSet(t *testing.T) {
	mock := setupMockStore(t)

	// Add profiles but no default
	_ = mock.SetToken("profile1", secrets.Token{RefreshToken: "token1"})
	_ = mock.SetToken("profile2", secrets.Token{RefreshToken: "token2"})
	// Don't set a default

	viper.Reset()
	viper.Set("output", "text")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListCmd_ListProfilesError(t *testing.T) {
	setupCustomStore(t, &listErrorStore{})

	viper.Reset()
	viper.Set("output", "text")

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("Expected error when ListProfiles fails")
	}
}

// listErrorStore is a mock that returns an error on ListTokens
type listErrorStore struct {
	secrets.MockStore
}

func (m *listErrorStore) Keys() ([]string, error) {
	return nil, nil
}

func (m *listErrorStore) SetToken(profile string, tok secrets.Token) error {
	return nil
}

func (m *listErrorStore) GetToken(profile string) (secrets.Token, error) {
	return secrets.Token{}, nil
}

func (m *listErrorStore) DeleteToken(profile string) error {
	return nil
}

func (m *listErrorStore) ListTokens() ([]secrets.Token, error) {
	return nil, errors.New("list tokens error")
}

func (m *listErrorStore) GetDefaultAccount() (string, error) {
	return "", nil
}

func (m *listErrorStore) SetDefaultAccount(profile string) error {
	return nil
}
