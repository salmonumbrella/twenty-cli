package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	cfg "github.com/salmonumbrella/twenty-cli/internal/config"
)

func TestSetCmd_Use(t *testing.T) {
	if setCmd.Use != "set <key> <value>" {
		t.Errorf("Use = %q, want %q", setCmd.Use, "set <key> <value>")
	}
}

func TestSetCmd_Short(t *testing.T) {
	if setCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestSetCmd_Args(t *testing.T) {
	// Test that command requires exactly 2 args
	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// With no args, should fail
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}

	// With one arg, should fail
	err = cmd.Args(cmd, []string{"key"})
	if err == nil {
		t.Error("expected error for one arg")
	}

	// With two args, should pass
	err = cmd.Args(cmd, []string{"key", "value"})
	if err != nil {
		t.Errorf("unexpected error for two args: %v", err)
	}

	// With three args, should fail
	err = cmd.Args(cmd, []string{"key", "value", "extra"})
	if err == nil {
		t.Error("expected error for three args")
	}
}

func TestSetCmd_SetBaseURL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	// Create initial config
	config := &cfg.Config{
		BaseURL: "https://old.example.com",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"base_url", "https://new.example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify config was updated
	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updatedConfig.BaseURL != "https://new.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://new.example.com")
	}
}

func TestSetCmd_SetBaseURL_Hyphenated(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"base-url", "https://hyphen.example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updatedConfig.BaseURL != "https://hyphen.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://hyphen.example.com")
	}
}

func TestSetCmd_SetKeyringBackend(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL: "https://api.example.com",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"keyring_backend", "file"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updatedConfig.KeyringBackend != "file" {
		t.Errorf("KeyringBackend = %q, want %q", updatedConfig.KeyringBackend, "file")
	}
}

func TestSetCmd_SetKeyringBackend_Hyphenated(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"keyring-backend", "pass"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updatedConfig.KeyringBackend != "pass" {
		t.Errorf("KeyringBackend = %q, want %q", updatedConfig.KeyringBackend, "pass")
	}
}

func TestSetCmd_UnknownKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"unknown_key", "value"})
	if err == nil {
		t.Fatal("expected error for unknown key")
	}

	expectedMsg := "unknown config key"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("error message should start with %q, got %q", expectedMsg, err.Error())
	}
}

func TestSetCmd_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	// No config file created - should work as Load returns empty config

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"base_url", "https://new.example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".twenty.yaml")
	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if updatedConfig.BaseURL != "https://new.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://new.example.com")
	}
}

func TestSetCmd_KeyCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test uppercase
	err := cmd.RunE(cmd, []string{"BASE_URL", "https://upper.example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updatedConfig.BaseURL != "https://upper.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://upper.example.com")
	}
}

func TestSetCmd_KeyWithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test with leading/trailing whitespace
	err := cmd.RunE(cmd, []string{"  base_url  ", "https://whitespace.example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updatedConfig.BaseURL != "https://whitespace.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://whitespace.example.com")
	}
}

func TestSetCmd_ValueWithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test with leading/trailing whitespace in value
	err := cmd.RunE(cmd, []string{"base_url", "  https://trimmed.example.com  "})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	// Value should be trimmed
	if updatedConfig.BaseURL != "https://trimmed.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://trimmed.example.com")
	}
}

func TestSetCmd_PreservesOtherConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	// Create initial config with multiple values
	config := &cfg.Config{
		BaseURL:        "https://original.example.com",
		KeyringBackend: "file",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Update only base_url
	err := cmd.RunE(cmd, []string{"base_url", "https://updated.example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updatedConfig, err := cfg.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	// Check base_url was updated
	if updatedConfig.BaseURL != "https://updated.example.com" {
		t.Errorf("BaseURL = %q, want %q", updatedConfig.BaseURL, "https://updated.example.com")
	}

	// Check keyring_backend was preserved
	if updatedConfig.KeyringBackend != "file" {
		t.Errorf("KeyringBackend = %q, want %q", updatedConfig.KeyringBackend, "file")
	}
}

func TestSetCmd_InvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0600); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cmd := setCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"base_url", "https://example.com"})
	if err == nil {
		t.Fatal("expected error for invalid config file")
	}
}
