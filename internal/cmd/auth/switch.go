package auth

import (
	"fmt"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <profile>",
	Short: "Set the default profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := args[0]

		// Ensure the profile exists before setting as default.
		profiles, err := ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}
		found := false
		for _, p := range profiles {
			if p == profile {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("profile %q not found; run: twenty auth login --profile %s --token <token>", profile, profile)
		}

		store, err := getStore()
		if err != nil {
			return fmt.Errorf("failed to open store: %w", err)
		}
		if err := store.SetDefaultAccount(profile); err != nil {
			return fmt.Errorf("failed to set default profile: %w", err)
		}

		fmt.Printf("Default profile set to %s\n", profile)
		return nil
	},
}
