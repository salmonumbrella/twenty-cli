package webhooks

import "github.com/spf13/cobra"

// Cmd is the parent command for webhook operations
var Cmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhooks",
	Long:  "Create, list, and delete webhooks in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(deleteCmd)
}
