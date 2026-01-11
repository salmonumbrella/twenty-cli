package auth

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoginCmd_CLIMode_MissingToken(t *testing.T) {
	setupMockStore(t)

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	// Set flags for CLI mode
	loginToken = ""
	loginBaseURL = "https://twenty.example.com"
	loginProfile = ""
	loginNoBrowser = true
	defer func() { loginNoBrowser = false }()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("Expected error when token is missing")
	}

	expected := "--token is required with --no-browser. Get your API token from Twenty Settings -> APIs & Webhooks"
	if err.Error() != expected {
		t.Errorf("Unexpected error message: got %q, want %q", err.Error(), expected)
	}
}

func TestLoginCmd_CLIMode_MissingBaseURL(t *testing.T) {
	setupMockStore(t)

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	// Set token but not base URL in CLI mode
	loginToken = "test-token"
	loginBaseURL = ""
	loginProfile = ""
	loginNoBrowser = true
	defer func() { loginNoBrowser = false }()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("Expected error when base URL is missing")
	}

	expected := "--base-url is required with --no-browser. Example: --base-url https://twenty.example.com"
	if err.Error() != expected {
		t.Errorf("Unexpected error message: got %q, want %q", err.Error(), expected)
	}
}

func TestLoginCmd_CLIMode_Success(t *testing.T) {
	mock := setupMockStore(t)

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	// Set valid inputs in CLI mode
	loginToken = "test-api-token"
	loginBaseURL = "https://twenty.example.com"
	loginProfile = ""
	loginNoBrowser = true
	defer func() { loginNoBrowser = false }()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify token was saved
	tok, err := mock.GetToken("default")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if tok.RefreshToken != "test-api-token" {
		t.Errorf("Token mismatch: got %q, want %q", tok.RefreshToken, "test-api-token")
	}
}

func TestLoginCmd_CLIMode_WithProfile(t *testing.T) {
	mock := setupMockStore(t)

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()

	// Set valid inputs with profile in CLI mode
	loginToken = "work-api-token"
	loginBaseURL = "https://work.twenty.example.com"
	loginProfile = "work"
	loginNoBrowser = true
	defer func() {
		loginNoBrowser = false
		loginProfile = ""
	}()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify token was saved for "work" profile
	tok, err := mock.GetToken("work")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if tok.RefreshToken != "work-api-token" {
		t.Errorf("Token mismatch: got %q, want %q", tok.RefreshToken, "work-api-token")
	}
}

func TestLoginCmd_CLIMode_TokenWithWhitespace(t *testing.T) {
	mock := setupMockStore(t)

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	// Set inputs with whitespace in CLI mode
	loginToken = "   test-token-with-spaces   "
	loginBaseURL = "   https://twenty.example.com   "
	loginProfile = ""
	loginNoBrowser = true
	defer func() { loginNoBrowser = false }()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify token was saved (trimmed)
	tok, err := mock.GetToken("default")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if tok.RefreshToken != "test-token-with-spaces" {
		t.Errorf("Token not trimmed: got %q", tok.RefreshToken)
	}
}

func TestSaveBaseURL(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()

	err := saveBaseURL("https://test.twenty.com")
	if err != nil {
		t.Fatalf("saveBaseURL failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".twenty.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if !bytes.Contains(data, []byte("base_url: https://test.twenty.com")) {
		t.Errorf("Config file doesn't contain expected base_url: %s", string(data))
	}

	// Verify viper was updated
	if viper.GetString("base_url") != "https://test.twenty.com" {
		t.Errorf("Viper not updated: got %q", viper.GetString("base_url"))
	}
}

func TestSaveBaseURL_UpdatesExistingConfig(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create existing config
	configPath := filepath.Join(tmpDir, ".twenty.yaml")
	existingConfig := []byte("existing_key: existing_value\nbase_url: old-url\n")
	if err := os.WriteFile(configPath, existingConfig, 0600); err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	// Reset viper
	viper.Reset()

	err := saveBaseURL("https://new.twenty.com")
	if err != nil {
		t.Fatalf("saveBaseURL failed: %v", err)
	}

	// Verify file was updated
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if !bytes.Contains(data, []byte("base_url: https://new.twenty.com")) {
		t.Errorf("Config file doesn't contain updated base_url: %s", string(data))
	}

	if !bytes.Contains(data, []byte("existing_key: existing_value")) {
		t.Errorf("Config file lost existing keys: %s", string(data))
	}
}

func TestLoginCmd_CLIMode_SaveTokenError(t *testing.T) {
	mock := setupMockStore(t)
	mock.SetSetError(errors.New("save token error"))

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()
	viper.Set("profile", "")

	// Set valid inputs in CLI mode
	loginToken = "test-api-token"
	loginBaseURL = "https://twenty.example.com"
	loginProfile = ""
	loginNoBrowser = true
	defer func() { loginNoBrowser = false }()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("Expected error when SaveToken fails")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("failed to save token")) {
		t.Errorf("Expected 'failed to save token' error, got: %v", err)
	}
}

func TestLoginCmd_CLIMode_SetsPrimaryOnFirstProfile(t *testing.T) {
	mock := setupMockStore(t)

	// Create temp config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()

	// Set valid inputs in CLI mode
	loginToken = "test-api-token"
	loginBaseURL = "https://twenty.example.com"
	loginProfile = "myprofile"
	loginNoBrowser = true
	defer func() {
		loginNoBrowser = false
		loginProfile = ""
	}()

	cmd := loginCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify token was saved
	tok, err := mock.GetToken("myprofile")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if tok.RefreshToken != "test-api-token" {
		t.Errorf("Token mismatch: got %q, want %q", tok.RefreshToken, "test-api-token")
	}

	// Verify it was set as primary (first profile)
	primary, err := mock.GetDefaultAccount()
	if err != nil {
		t.Fatalf("Failed to get default account: %v", err)
	}
	if primary != "myprofile" {
		t.Errorf("Expected myprofile to be primary, got %q", primary)
	}
}
