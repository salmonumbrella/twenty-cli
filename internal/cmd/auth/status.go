package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var statusShowToken bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		output := viper.GetString("output")
		query := viper.GetString("query")

		// Check for env var token first
		envToken := os.Getenv(tokenEnv)
		if envToken != "" {
			displayToken := envToken
			if !statusShowToken {
				displayToken = envToken
				if len(envToken) > 12 {
					displayToken = envToken[:8] + "..." + envToken[len(envToken)-4:]
				}
			}
			baseURL := viper.GetString("base_url")

			if output == "json" {
				payload := map[string]interface{}{
					"profile":   "env",
					"logged_in": true,
					"token":     displayToken,
					"base_url":  baseURL,
					"source":    "TWENTY_TOKEN",
				}
				return outfmt.WriteJSON(os.Stdout, payload, query)
			}
			fmt.Println("Profile: env (from TWENTY_TOKEN)")
			fmt.Println("Status: Logged in")
			fmt.Printf("Token: %s\n", displayToken)
			if baseURL != "" {
				fmt.Printf("Base URL: %s\n", baseURL)
			}
			return nil
		}

		profile, err := CurrentProfile()
		if err != nil {
			return err
		}

		token, err := GetToken(profile)
		if err != nil || token == "" {
			if output == "json" {
				payload := map[string]interface{}{
					"profile":   profile,
					"logged_in": false,
					"base_url":  viper.GetString("base_url"),
				}
				return outfmt.WriteJSON(os.Stdout, payload, query)
			}
			fmt.Printf("Profile: %s\n", profile)
			fmt.Println("Status: Not logged in")
			return nil
		}

		displayToken := token
		if !statusShowToken {
			if len(token) > 12 {
				displayToken = token[:8] + "..." + token[len(token)-4:]
			}
		}

		baseURL := viper.GetString("base_url")

		if output == "json" {
			payload := map[string]interface{}{
				"profile":   profile,
				"logged_in": true,
				"token":     displayToken,
				"base_url":  baseURL,
			}
			return outfmt.WriteJSON(os.Stdout, payload, query)
		}

		fmt.Printf("Profile: %s\n", profile)
		fmt.Println("Status: Logged in")
		fmt.Printf("Token: %s\n", displayToken)

		if baseURL != "" {
			fmt.Printf("Base URL: %s\n", baseURL)
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&statusShowToken, "show-token", false, "show full token value")
}
