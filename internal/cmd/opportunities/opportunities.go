package opportunities

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "opportunities",
	Short: "Manage opportunities",
	Long:  "Create, read, update, and delete opportunities in Twenty CRM.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
