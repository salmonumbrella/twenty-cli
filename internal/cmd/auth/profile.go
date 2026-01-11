package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// tokenEnvData represents the JSON structure of TWENTY_TOKEN environment variable
type tokenEnvData struct {
	RefreshToken string `json:"refresh_token"`
	CreatedAt    string `json:"created_at"`
}

const (
	// tokenEnv is the environment variable for providing a token directly
	tokenEnv = "TWENTY_TOKEN" //nolint:gosec // env var name, not a credential
)

// ResolveProfile returns the profile to use, honoring the active default profile if set.
func ResolveProfile(explicit string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		return explicit, nil
	}

	s, err := getStore()
	if err != nil {
		return "default", nil
	}

	defaultProfile, err := s.GetDefaultAccount()
	if err != nil {
		return "", err
	}
	if defaultProfile != "" {
		return defaultProfile, nil
	}

	return "default", nil
}

// CurrentProfile resolves the current profile based on flags and stored default.
func CurrentProfile() (string, error) {
	return ResolveProfile(viper.GetString("profile"))
}

// RequireToken returns the resolved profile and token or an error if missing.
// It first checks for TWENTY_TOKEN environment variable, then falls back to keychain.
// TWENTY_TOKEN can be either a raw token string or a JSON object with "refresh_token" field.
func RequireToken() (string, string, error) {
	// Check for environment variable first (avoids keychain prompts)
	if envToken := strings.TrimSpace(os.Getenv(tokenEnv)); envToken != "" {
		// Try to parse as JSON first (for {"refresh_token": "...", "created_at": "..."} format)
		if strings.HasPrefix(envToken, "{") {
			var data tokenEnvData
			if err := json.Unmarshal([]byte(envToken), &data); err == nil && data.RefreshToken != "" {
				return "env", data.RefreshToken, nil
			}
		}
		// Otherwise treat as raw token
		return "env", envToken, nil
	}

	profile, err := CurrentProfile()
	if err != nil {
		return "", "", err
	}

	token, err := GetToken(profile)
	if err != nil || token == "" {
		// Check if profile was explicitly specified via -p flag
		explicitProfile := strings.TrimSpace(viper.GetString("profile"))
		if explicitProfile != "" {
			// Check if this profile exists (normalize for comparison)
			profiles, listErr := ListProfiles()
			if listErr == nil {
				normalizedProfile := strings.ToLower(strings.TrimSpace(profile))
				found := false
				for _, p := range profiles {
					if p == normalizedProfile {
						found = true
						break
					}
				}
				if !found {
					return profile, "", fmt.Errorf("profile %q not found. Run: twenty auth login -p %s --token <your-token> --base-url <url>", profile, profile)
				}
			}
		}
		return profile, "", fmt.Errorf("not logged in. Run: twenty auth login --token <your-token>")
	}

	return profile, token, nil
}
