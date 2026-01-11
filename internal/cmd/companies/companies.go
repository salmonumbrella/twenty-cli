package companies

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "companies",
	Short: "Manage companies",
	Long:  "Create, read, update, and delete companies in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
