package notes

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "notes",
	Short: "Manage notes",
	Long:  "Create, read, update, and delete notes in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
