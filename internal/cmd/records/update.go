package records

import (
	"github.com/spf13/cobra"
)

var (
	updateData     string
	updateDataFile string
	updateSet      []string
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <object> <id>",
		Short: "Update a record",
		Args:  cobra.ExactArgs(2),
		RunE:  runUpdate,
	}
	cmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON payload for update")
	cmd.Flags().StringVarP(&updateDataFile, "file", "f", "", "JSON file for update (use - for stdin)")
	cmd.Flags().StringArrayVar(&updateSet, "set", nil, "set field value (key=value)")
	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	object := args[0]
	id := args[1]
	return runMutation("PATCH", object, id, updateData, updateDataFile, updateSet)
}
