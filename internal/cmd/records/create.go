package records

import (
	"github.com/spf13/cobra"
)

var (
	createData     string
	createDataFile string
	createSet      []string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <object>",
		Short: "Create a record",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
	cmd.Flags().StringVarP(&createData, "data", "d", "", "JSON payload for creation")
	cmd.Flags().StringVarP(&createDataFile, "file", "f", "", "JSON file for creation (use - for stdin)")
	cmd.Flags().StringArrayVar(&createSet, "set", nil, "set field value (key=value)")
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	object := args[0]
	return runMutation("POST", object, "", createData, createDataFile, createSet)
}
