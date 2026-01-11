package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/salmonumbrella/twenty-cli/internal/config"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cfg.DefaultConfigPath()
		if err != nil {
			return err
		}

		conf, err := cfg.Load(path)
		if err != nil {
			return err
		}

		output := viper.GetString("output")
		query := viper.GetString("query")

		switch output {
		case "json":
			payload := map[string]interface{}{
				"path":   path,
				"config": conf,
			}
			return outfmt.WriteJSON(os.Stdout, payload, query)
		case "csv":
			headers := []string{"key", "value"}
			rows := [][]string{
				{"path", path},
				{"baseUrl", conf.BaseURL},
			}
			if conf.KeyringBackend != "" {
				rows = append(rows, []string{"keyringBackend", conf.KeyringBackend})
			}
			return outfmt.WriteCSV(os.Stdout, headers, rows)
		default:
			fmt.Printf("Path: %s\n", path)
			fmt.Printf("Base URL: %s\n", conf.BaseURL)
			if conf.KeyringBackend != "" {
				fmt.Printf("Keyring Backend: %s\n", conf.KeyringBackend)
			}
		}
		return nil
	},
}
