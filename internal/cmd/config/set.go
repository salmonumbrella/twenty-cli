package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/salmonumbrella/twenty-cli/internal/config"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(strings.TrimSpace(args[0]))
		value := strings.TrimSpace(args[1])

		path, err := cfg.DefaultConfigPath()
		if err != nil {
			return err
		}

		conf, err := cfg.Load(path)
		if err != nil {
			return err
		}

		switch key {
		case "base_url", "base-url":
			conf.BaseURL = value
		case "keyring_backend", "keyring-backend":
			conf.KeyringBackend = value
		default:
			return fmt.Errorf("unknown config key %q (supported: base_url, keyring_backend)", key)
		}

		if err := conf.Save(path); err != nil {
			return err
		}

		fmt.Printf("Updated %s\n", key)
		return nil
	},
}
