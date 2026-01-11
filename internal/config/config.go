package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppName is the application name used for keyring and config
const AppName = "twenty"

// Internal hooks for testing - these allow tests to inject failures
var (
	userHomeDir = os.UserHomeDir
	yamlMarshal = yaml.Marshal
	mkdirAll    = os.MkdirAll
)

// Config holds CLI configuration
type Config struct {
	BaseURL        string `yaml:"base_url"`
	Token          string `yaml:"token,omitempty"`
	KeyringBackend string `yaml:"keyring_backend,omitempty"`
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".twenty.yaml"), nil
}

// Load loads config from the given path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Save saves config to the given path
func (c *Config) Save(path string) error {
	data, err := yamlMarshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := mkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// ReadConfig reads the config from the default path
func ReadConfig() (*Config, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	return Load(path)
}

// EnsureKeyringDir creates and returns the keyring directory path
func EnsureKeyringDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	keyringDir := filepath.Join(home, ".config", AppName, "keyring")
	if err := mkdirAll(keyringDir, 0700); err != nil {
		return "", fmt.Errorf("creating keyring directory: %w", err)
	}

	return keyringDir, nil
}
