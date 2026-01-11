package attachments

import "github.com/spf13/cobra"

// Cmd is the parent command for attachment operations.
var Cmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage attachments",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
