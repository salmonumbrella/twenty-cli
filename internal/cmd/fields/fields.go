package fields

import "github.com/spf13/cobra"

// Cmd is the parent command for field metadata operations
var Cmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage field metadata",
	Long:  "List and inspect field metadata across all objects in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
