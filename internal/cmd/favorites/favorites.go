package favorites

import "github.com/spf13/cobra"

// Cmd is the parent command for favorite operations.
var Cmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage favorites",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
