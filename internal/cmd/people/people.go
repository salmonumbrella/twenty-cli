package people

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "people",
	Short: "Manage people (contacts)",
	Long:  "Create, read, update, and delete people in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(upsertCmd)
	Cmd.AddCommand(batchCreateCmd)
	Cmd.AddCommand(batchDeleteCmd)
	Cmd.AddCommand(exportCmd)
	Cmd.AddCommand(importCmd)
}
