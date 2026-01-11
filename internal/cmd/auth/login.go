package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	loginToken   string
	loginBaseURL string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Twenty API",
	Long:  "Login using API token. Get your token from Settings -> APIs & Webhooks in Twenty.",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := CurrentProfile()
		if err != nil {
			return err
		}

		loginToken = strings.TrimSpace(loginToken)
		if loginToken == "" {
			return fmt.Errorf("--token is required. Get your API token from Twenty Settings -> APIs & Webhooks")
		}

		loginBaseURL = strings.TrimSpace(loginBaseURL)
		if loginBaseURL == "" {
			return fmt.Errorf("--base-url is required. Example: --base-url https://twenty.example.com")
		}

		if err := SaveToken(profile, loginToken); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		// Save base_url to config file
		if err := saveBaseURL(loginBaseURL); err != nil {
			return fmt.Errorf("failed to save base URL: %w", err)
		}

		fmt.Printf("Logged in successfully (profile: %s)\n", profile)
		fmt.Printf("Base URL: %s\n", loginBaseURL)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginToken, "token", "", "API token (from Twenty settings)")
	loginCmd.Flags().StringVar(&loginBaseURL, "base-url", "", "Twenty API base URL (e.g., https://twenty.example.com)")
}

func saveBaseURL(baseURL string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".twenty.yaml")

	// Read existing config or create new
	config := make(map[string]interface{})
	if data, err := os.ReadFile(configPath); err == nil {
		_ = yaml.Unmarshal(data, &config) // Ignore error; start fresh if malformed
	}

	config["base_url"] = baseURL

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	// Update viper so it's available immediately
	viper.Set("base_url", baseURL)

	return nil
}
