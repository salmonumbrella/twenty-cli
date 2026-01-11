package objects

import "github.com/spf13/cobra"

// Cmd is the parent command for object metadata operations
var Cmd = &cobra.Command{
	Use:   "objects",
	Short: "Manage object metadata",
	Long:  "List and inspect object metadata (types, fields, relations) in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
