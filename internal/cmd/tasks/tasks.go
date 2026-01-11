package tasks

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
	Long:  "Create, read, update, and delete tasks in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
