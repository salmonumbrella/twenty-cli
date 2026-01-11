package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/salmonumbrella/twenty-cli/internal/auth"
)

var (
	loginToken     string
	loginBaseURL   string
	loginProfile   string
	loginNoBrowser bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Twenty API",
	Long: `Login to Twenty API using browser-based setup or CLI flags.

By default, opens a browser for interactive credential setup.
Use --no-browser for non-interactive login.

Examples:
  # Interactive browser-based login
  twenty auth login

  # CLI-based login
  twenty auth login --no-browser --base-url https://app.twenty.com --token YOUR_TOKEN

  # Login to a named profile
  twenty auth login --profile staging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve profile name
		profile := strings.TrimSpace(loginProfile)
		if profile == "" {
			profile = "default"
		}

		if loginNoBrowser {
			return runCLILogin(profile)
		}
		return runBrowserLogin(cmd.Context())
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginToken, "token", "", "API token (from Twenty settings)")
	loginCmd.Flags().StringVar(&loginBaseURL, "base-url", "", "Twenty API base URL (e.g., https://twenty.example.com)")
	loginCmd.Flags().StringVarP(&loginProfile, "profile", "p", "", "Profile name (default: default)")
	loginCmd.Flags().BoolVar(&loginNoBrowser, "no-browser", false, "Use CLI mode instead of browser-based setup")
}

// runBrowserLogin performs browser-based login using SetupServer.
// The profile is collected via the web form, not from CLI flags.
func runBrowserLogin(ctx context.Context) error {
	// Create setup server
	server, err := auth.NewSetupServer()
	if err != nil {
		return fmt.Errorf("failed to start setup server: %w", err)
	}
	defer server.Close()

	// Start server
	server.Start(ctx)

	// Open browser
	setupURL := server.URL()
	fmt.Printf("Opening browser for credential setup...\n")
	fmt.Printf("If the browser doesn't open, visit: %s\n\n", setupURL)

	if err := auth.OpenBrowser(setupURL); err != nil {
		// Browser failed to open, but URL is already printed
		fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", err)
	}

	// Wait for setup completion
	result, err := server.WaitForSetup(ctx)
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Save base_url to config file
	if err := saveBaseURL(result.BaseURL); err != nil {
		return fmt.Errorf("failed to save base URL: %w", err)
	}

	fmt.Printf("\nLogged in successfully (profile: %s)\n", result.Profile)
	fmt.Printf("Base URL: %s\n", result.BaseURL)
	return nil
}

// runCLILogin performs CLI-based login with --token and --base-url flags
func runCLILogin(profile string) error {
	loginToken = strings.TrimSpace(loginToken)
	if loginToken == "" {
		return fmt.Errorf("--token is required with --no-browser. Get your API token from Twenty Settings -> APIs & Webhooks")
	}

	loginBaseURL = strings.TrimSpace(loginBaseURL)
	if loginBaseURL == "" {
		return fmt.Errorf("--base-url is required with --no-browser. Example: --base-url https://twenty.example.com")
	}

	// Save token
	if err := SaveToken(profile, loginToken); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Set as primary if first profile or no primary is set
	if err := setPrimaryIfNeeded(profile); err != nil {
		// Non-fatal, just log warning
		fmt.Fprintf(os.Stderr, "Warning: could not set primary profile: %v\n", err)
	}

	// Save base_url to config file
	if err := saveBaseURL(loginBaseURL); err != nil {
		return fmt.Errorf("failed to save base URL: %w", err)
	}

	fmt.Printf("Logged in successfully (profile: %s)\n", profile)
	fmt.Printf("Base URL: %s\n", loginBaseURL)
	return nil
}

// setPrimaryIfNeeded sets the profile as primary if it's the first profile or no primary is set
func setPrimaryIfNeeded(profile string) error {
	s, err := getStore()
	if err != nil {
		return err
	}

	// Check if this is the first profile
	tokens, err := s.ListTokens()
	if err == nil && len(tokens) == 1 {
		return s.SetDefaultAccount(profile)
	}

	// Check if no primary is set
	primary, err := s.GetDefaultAccount()
	if err == nil && primary == "" {
		return s.SetDefaultAccount(profile)
	}

	return nil
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
