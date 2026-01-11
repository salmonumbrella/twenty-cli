package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	cfg "github.com/salmonumbrella/twenty-cli/internal/config"
)

func TestShowCmd_Use(t *testing.T) {
	if showCmd.Use != "show" {
		t.Errorf("Use = %q, want %q", showCmd.Use, "show")
	}
}

func TestShowCmd_Short(t *testing.T) {
	if showCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestShowCmd_TextOutput(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL: "https://api.twenty.com",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Override DefaultConfigPath by using environment
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "text")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_TextOutput_WithKeyringBackend(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL:        "https://api.twenty.com",
		KeyringBackend: "file",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "text")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL: "https://api.twenty.com",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "json")
	viper.Set("query", "")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_JSONOutput_WithQuery(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL: "https://api.twenty.com",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "json")
	viper.Set("query", ".path")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_CSVOutput(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL: "https://api.twenty.com",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "csv")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_CSVOutput_WithKeyringBackend(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	config := &cfg.Config{
		BaseURL:        "https://api.twenty.com",
		KeyringBackend: "file",
	}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "csv")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	// No config file created

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "text")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	// Create empty config
	config := &cfg.Config{}
	if err := config.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "text")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestShowCmd_InvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".twenty.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0600); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	viper.Reset()
	viper.Set("output", "text")

	cmd := showCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid config file")
	}
}
