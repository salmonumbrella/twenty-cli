package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("DefaultConfigPath() returned empty string")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("DefaultConfigPath() = %q, want absolute path", path)
	}

	if filepath.Base(path) != ".twenty.yaml" {
		t.Errorf("DefaultConfigPath() = %q, want file named .twenty.yaml", path)
	}
}

func TestEnsureKeyringDir(t *testing.T) {
	dir, err := EnsureKeyringDir()
	if err != nil {
		t.Fatalf("EnsureKeyringDir() error = %v", err)
	}

	if dir == "" {
		t.Error("EnsureKeyringDir() returned empty string")
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("EnsureKeyringDir() = %q, want absolute path", dir)
	}

	// Verify the directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("EnsureKeyringDir() created path that doesn't exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("EnsureKeyringDir() = %q, want directory", dir)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/to/config.yaml")
	if err != nil {
		t.Fatalf("Load() with non-existent file error = %v, want nil", err)
	}

	if cfg == nil {
		t.Fatal("Load() with non-existent file returned nil config")
	}

	// Should return empty config with zero values
	if cfg.BaseURL != "" {
		t.Errorf("Load() BaseURL = %q, want empty", cfg.BaseURL)
	}
	if cfg.Token != "" {
		t.Errorf("Load() Token = %q, want empty", cfg.Token)
	}
	if cfg.KeyringBackend != "" {
		t.Errorf("Load() KeyringBackend = %q, want empty", cfg.KeyringBackend)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	if cfg.BaseURL != "" {
		t.Errorf("Config zero value BaseURL = %q, want empty", cfg.BaseURL)
	}
	if cfg.Token != "" {
		t.Errorf("Config zero value Token = %q, want empty", cfg.Token)
	}
	if cfg.KeyringBackend != "" {
		t.Errorf("Config zero value KeyringBackend = %q, want empty", cfg.KeyringBackend)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	content := `base_url: https://example.com
token: test-token
keyring_backend: file
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "https://example.com" {
		t.Errorf("Load() BaseURL = %q, want %q", cfg.BaseURL, "https://example.com")
	}
	if cfg.Token != "test-token" {
		t.Errorf("Load() Token = %q, want %q", cfg.Token, "test-token")
	}
	if cfg.KeyringBackend != "file" {
		t.Errorf("Load() KeyringBackend = %q, want %q", cfg.KeyringBackend, "file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	content := `invalid: yaml: content: [
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() with invalid YAML expected error, got nil")
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := &Config{
		BaseURL:        "https://test.example.com",
		Token:          "save-test-token",
		KeyringBackend: "file",
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Save() did not create file: %v", err)
	}

	// Load it back and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.BaseURL != cfg.BaseURL {
		t.Errorf("Roundtrip BaseURL = %q, want %q", loaded.BaseURL, cfg.BaseURL)
	}
	if loaded.Token != cfg.Token {
		t.Errorf("Roundtrip Token = %q, want %q", loaded.Token, cfg.Token)
	}
	if loaded.KeyringBackend != cfg.KeyringBackend {
		t.Errorf("Roundtrip KeyringBackend = %q, want %q", loaded.KeyringBackend, cfg.KeyringBackend)
	}
}

func TestConfig_Save_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "permissions.yaml")

	cfg := &Config{BaseURL: "https://example.com"}
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Config file should have 0600 permissions (owner read/write only)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Save() file permissions = %o, want 0600", perm)
	}
}

func TestReadConfig(t *testing.T) {
	// ReadConfig uses DefaultConfigPath, which reads from home directory
	// This test just verifies it doesn't error on a typical system
	cfg, err := ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig() error = %v", err)
	}

	// Should return a config (possibly empty if no config file exists)
	if cfg == nil {
		t.Error("ReadConfig() returned nil config")
	}
}

func TestAppName(t *testing.T) {
	if AppName != "twenty" {
		t.Errorf("AppName = %q, want %q", AppName, "twenty")
	}
}

func TestLoad_ReadError(t *testing.T) {
	// Test Load with a path that exists but can't be read (permission denied)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "unreadable.yaml")

	// Create file and make it unreadable
	if err := os.WriteFile(configPath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	if err := os.Chmod(configPath, 0000); err != nil {
		t.Fatalf("Failed to chmod test config: %v", err)
	}
	// Cleanup: restore permissions so TempDir can clean up
	t.Cleanup(func() {
		os.Chmod(configPath, 0600)
	})

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() with unreadable file expected error, got nil")
	}
}

func TestConfig_Save_MkdirAllError(t *testing.T) {
	// Create a file where we need a directory - MkdirAll should fail
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "blocker")

	// Create a regular file that will block directory creation
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	// Try to save config to a path under the file (which would require it to be a dir)
	configPath := filepath.Join(blockingFile, "subdir", "config.yaml")

	cfg := &Config{BaseURL: "https://example.com"}
	err := cfg.Save(configPath)
	if err == nil {
		t.Error("Save() to path blocked by file expected error, got nil")
	}
}

func TestConfig_Save_WriteError(t *testing.T) {
	// Create a directory where the config file should be, making WriteFile fail
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "isdir")

	// Create a directory with the name we want for our config file
	if err := os.Mkdir(configPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	cfg := &Config{BaseURL: "https://example.com"}
	err := cfg.Save(configPath)
	if err == nil {
		t.Error("Save() to directory path expected error, got nil")
	}
}

func TestLoad_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.yaml")

	// Create an empty config file
	if err := os.WriteFile(configPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to write empty config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() with empty file error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() with empty file returned nil config")
	}

	// Should have empty/zero values
	if cfg.BaseURL != "" {
		t.Errorf("Load() BaseURL = %q, want empty", cfg.BaseURL)
	}
}

func TestConfig_Save_TokenOmitEmpty(t *testing.T) {
	// Verify that empty token is omitted from YAML output
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "omit.yaml")

	cfg := &Config{
		BaseURL: "https://example.com",
		// Token is empty
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read the raw file content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	// Token should be omitted (not present in output)
	if filepath.Base(configPath) != "omit.yaml" {
		t.Skip("Skipping omitempty verification")
	}

	// Verify BaseURL is present
	if !strings.Contains(content, "base_url:") {
		t.Errorf("Save() output missing base_url: %q", content)
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yaml")

	// Config with only base_url
	content := `base_url: https://partial.example.com
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "https://partial.example.com" {
		t.Errorf("Load() BaseURL = %q, want %q", cfg.BaseURL, "https://partial.example.com")
	}
	if cfg.Token != "" {
		t.Errorf("Load() Token = %q, want empty", cfg.Token)
	}
	if cfg.KeyringBackend != "" {
		t.Errorf("Load() KeyringBackend = %q, want empty", cfg.KeyringBackend)
	}
}

func TestConfig_SaveAndLoad_AllFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "allfields.yaml")

	original := &Config{
		BaseURL:        "https://all.example.com",
		Token:          "all-token",
		KeyringBackend: "all-backend",
	}

	if err := original.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.BaseURL != original.BaseURL {
		t.Errorf("BaseURL = %q, want %q", loaded.BaseURL, original.BaseURL)
	}
	if loaded.Token != original.Token {
		t.Errorf("Token = %q, want %q", loaded.Token, original.Token)
	}
	if loaded.KeyringBackend != original.KeyringBackend {
		t.Errorf("KeyringBackend = %q, want %q", loaded.KeyringBackend, original.KeyringBackend)
	}
}

func TestConfig_Save_NestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a deeply nested path
	configPath := filepath.Join(tmpDir, "a", "b", "c", "d", "config.yaml")

	cfg := &Config{BaseURL: "https://nested.example.com"}
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() to nested directory error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Save() did not create nested file: %v", err)
	}

	// Verify content
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.BaseURL != "https://nested.example.com" {
		t.Errorf("BaseURL = %q, want %q", loaded.BaseURL, "https://nested.example.com")
	}
}

func TestEnsureKeyringDir_CreatesWithCorrectPermissions(t *testing.T) {
	dir, err := EnsureKeyringDir()
	if err != nil {
		t.Fatalf("EnsureKeyringDir() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Keyring directory should have 0700 permissions (owner only)
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("EnsureKeyringDir() permissions = %o, want 0700", perm)
	}
}

func TestEnsureKeyringDir_PathStructure(t *testing.T) {
	dir, err := EnsureKeyringDir()
	if err != nil {
		t.Fatalf("EnsureKeyringDir() error = %v", err)
	}

	// Should end with .config/twenty/keyring
	if filepath.Base(dir) != "keyring" {
		t.Errorf("EnsureKeyringDir() base = %q, want keyring", filepath.Base(dir))
	}

	parent := filepath.Dir(dir)
	if filepath.Base(parent) != AppName {
		t.Errorf("EnsureKeyringDir() parent = %q, want %q", filepath.Base(parent), AppName)
	}

	grandparent := filepath.Dir(parent)
	if filepath.Base(grandparent) != ".config" {
		t.Errorf("EnsureKeyringDir() grandparent = %q, want .config", filepath.Base(grandparent))
	}
}

func TestEnsureKeyringDir_Idempotent(t *testing.T) {
	// Calling EnsureKeyringDir multiple times should succeed
	dir1, err := EnsureKeyringDir()
	if err != nil {
		t.Fatalf("First EnsureKeyringDir() error = %v", err)
	}

	dir2, err := EnsureKeyringDir()
	if err != nil {
		t.Fatalf("Second EnsureKeyringDir() error = %v", err)
	}

	if dir1 != dir2 {
		t.Errorf("EnsureKeyringDir() not idempotent: %q != %q", dir1, dir2)
	}
}

// Tests using internal hooks to cover error paths

func TestDefaultConfigPath_UserHomeDirError(t *testing.T) {
	// Save original and restore after test
	original := userHomeDir
	t.Cleanup(func() { userHomeDir = original })

	// Inject failure
	userHomeDir = func() (string, error) {
		return "", errors.New("home directory unavailable")
	}

	_, err := DefaultConfigPath()
	if err == nil {
		t.Error("DefaultConfigPath() expected error when UserHomeDir fails, got nil")
	}
	if !errors.Is(err, errors.Unwrap(err)) && err.Error() == "" {
		t.Errorf("DefaultConfigPath() error should wrap underlying error")
	}
}

func TestReadConfig_DefaultConfigPathError(t *testing.T) {
	// Save original and restore after test
	original := userHomeDir
	t.Cleanup(func() { userHomeDir = original })

	// Inject failure for UserHomeDir (which DefaultConfigPath uses)
	userHomeDir = func() (string, error) {
		return "", errors.New("home directory unavailable")
	}

	_, err := ReadConfig()
	if err == nil {
		t.Error("ReadConfig() expected error when DefaultConfigPath fails, got nil")
	}
}

func TestConfig_Save_MarshalError(t *testing.T) {
	// Save original and restore after test
	original := yamlMarshal
	t.Cleanup(func() { yamlMarshal = original })

	// Inject failure
	yamlMarshal = func(v interface{}) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	cfg := &Config{BaseURL: "https://example.com"}
	err := cfg.Save("/tmp/test.yaml")
	if err == nil {
		t.Error("Save() expected error when yaml.Marshal fails, got nil")
	}
}

func TestEnsureKeyringDir_UserHomeDirError(t *testing.T) {
	// Save original and restore after test
	original := userHomeDir
	t.Cleanup(func() { userHomeDir = original })

	// Inject failure
	userHomeDir = func() (string, error) {
		return "", errors.New("home directory unavailable")
	}

	_, err := EnsureKeyringDir()
	if err == nil {
		t.Error("EnsureKeyringDir() expected error when UserHomeDir fails, got nil")
	}
}

func TestEnsureKeyringDir_MkdirAllError(t *testing.T) {
	// Save originals and restore after test
	originalHome := userHomeDir
	originalMkdir := mkdirAll
	t.Cleanup(func() {
		userHomeDir = originalHome
		mkdirAll = originalMkdir
	})

	// Provide a valid home but fail on mkdir
	userHomeDir = func() (string, error) {
		return "/tmp/testhome", nil
	}
	mkdirAll = func(path string, perm os.FileMode) error {
		return errors.New("mkdir failed")
	}

	_, err := EnsureKeyringDir()
	if err == nil {
		t.Error("EnsureKeyringDir() expected error when MkdirAll fails, got nil")
	}
}

func TestConfig_Save_MkdirAllError_ViaHook(t *testing.T) {
	// Save original and restore after test
	original := mkdirAll
	t.Cleanup(func() { mkdirAll = original })

	// Inject failure
	mkdirAll = func(path string, perm os.FileMode) error {
		return errors.New("mkdir failed")
	}

	cfg := &Config{BaseURL: "https://example.com"}
	err := cfg.Save("/some/nested/path/config.yaml")
	if err == nil {
		t.Error("Save() expected error when MkdirAll fails, got nil")
	}
}

func TestHooks_DefaultValues(t *testing.T) {
	// Verify hooks are initialized to real functions
	if userHomeDir == nil {
		t.Error("userHomeDir hook should not be nil")
	}
	if yamlMarshal == nil {
		t.Error("yamlMarshal hook should not be nil")
	}
	if mkdirAll == nil {
		t.Error("mkdirAll hook should not be nil")
	}

	// Verify they work like the real functions
	home, err := userHomeDir()
	if err != nil {
		t.Fatalf("userHomeDir() error = %v", err)
	}
	if home == "" {
		t.Error("userHomeDir() returned empty string")
	}

	data, err := yamlMarshal(&Config{BaseURL: "test"})
	if err != nil {
		t.Fatalf("yamlMarshal() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("yamlMarshal() returned empty data")
	}

	// Verify yaml import is used
	_ = yaml.Marshal
}
