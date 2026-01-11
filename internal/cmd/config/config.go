package config

import "github.com/spf13/cobra"

// Cmd is the parent command for config operations.
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

func init() {
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(setCmd)
}
