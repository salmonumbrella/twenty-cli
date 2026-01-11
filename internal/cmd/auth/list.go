package auth

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}
		sort.Strings(profiles)

		store, err := getStore()
		if err != nil {
			return fmt.Errorf("failed to open store: %w", err)
		}
		defaultProfile, _ := store.GetDefaultAccount()

		output := viper.GetString("output")
		query := viper.GetString("query")

		payload := map[string]interface{}{
			"default":  defaultProfile,
			"profiles": profiles,
		}
		switch output {
		case "json":
			return outfmt.WriteJSON(os.Stdout, payload, query)
		case "yaml":
			return outfmt.WriteYAML(os.Stdout, payload, query)
		case "csv":
			return outfmt.WriteCSVFromJSON(os.Stdout, payload)
		}

		fmt.Println("Profiles:")
		for _, p := range profiles {
			marker := " "
			if p == defaultProfile {
				marker = "*"
			}
			fmt.Printf(" %s %s\n", marker, p)
		}
		if defaultProfile == "" {
			fmt.Println("No default profile set. Use: twenty auth switch <profile>")
		}
		return nil
	},
}
