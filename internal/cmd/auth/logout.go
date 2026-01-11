package auth

import (
	"errors"
	"fmt"

	"github.com/99designs/keyring"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := CurrentProfile()
		if err != nil {
			return err
		}

		token, err := GetToken(profile)
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				fmt.Printf("Already logged out (profile: %s)\n", profile)
				return nil
			}
			return err
		}
		if token == "" {
			fmt.Printf("Already logged out (profile: %s)\n", profile)
			return nil
		}

		if err := DeleteToken(profile); err != nil {
			return fmt.Errorf("failed to delete token: %w", err)
		}

		fmt.Printf("Logged out (profile: %s)\n", profile)
		return nil
	},
}
