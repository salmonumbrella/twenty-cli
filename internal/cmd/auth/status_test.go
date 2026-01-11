package auth

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestStatusCmd_WithEnvToken(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Setenv("TWENTY_TOKEN", "env-token-12345678")

	viper.Reset()
	viper.Set("output", "text")
	viper.Set("base_url", "https://twenty.example.com")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithEnvToken_JSON(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Setenv("TWENTY_TOKEN", "env-token-12345678")

	viper.Reset()
	viper.Set("output", "json")
	viper.Set("query", "")
	viper.Set("base_url", "https://twenty.example.com")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithEnvToken_ShortToken(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	// Short token (less than 12 chars)
	os.Setenv("TWENTY_TOKEN", "short")

	viper.Reset()
	viper.Set("output", "text")
	viper.Set("base_url", "")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithEnvToken_ShowToken(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalShowToken := statusShowToken
	t.Cleanup(func() {
		os.Setenv("TWENTY_TOKEN", originalEnv)
		statusShowToken = originalShowToken
	})

	os.Setenv("TWENTY_TOKEN", "full-token-value-here")
	statusShowToken = true

	viper.Reset()
	viper.Set("output", "text")
	viper.Set("base_url", "https://twenty.example.com")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithStoredToken(t *testing.T) {
	mock := setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Unsetenv("TWENTY_TOKEN")

	// Store a token
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "stored-token-12345678"})

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "text")
	viper.Set("base_url", "https://twenty.example.com")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithStoredToken_JSON(t *testing.T) {
	mock := setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Unsetenv("TWENTY_TOKEN")

	// Store a token
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "stored-token-12345678"})

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "json")
	viper.Set("query", "")
	viper.Set("base_url", "https://twenty.example.com")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithStoredToken_ShowToken(t *testing.T) {
	mock := setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalShowToken := statusShowToken
	t.Cleanup(func() {
		os.Setenv("TWENTY_TOKEN", originalEnv)
		statusShowToken = originalShowToken
	})

	os.Unsetenv("TWENTY_TOKEN")
	statusShowToken = true

	// Store a token
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "stored-token-12345678"})

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "text")
	viper.Set("base_url", "")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_WithStoredToken_ShortToken(t *testing.T) {
	mock := setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Unsetenv("TWENTY_TOKEN")

	// Store a short token
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "short"})

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "text")
	viper.Set("base_url", "")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_NotLoggedIn(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Unsetenv("TWENTY_TOKEN")

	// No token stored

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "text")
	viper.Set("base_url", "")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStatusCmd_NotLoggedIn_JSON(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	os.Unsetenv("TWENTY_TOKEN")

	// No token stored

	viper.Reset()
	viper.Set("profile", "")
	viper.Set("output", "json")
	viper.Set("query", "")
	viper.Set("base_url", "https://twenty.example.com")

	cmd := statusCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
