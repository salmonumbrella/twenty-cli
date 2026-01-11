package records

import "github.com/spf13/cobra"

// Cmd is the parent command for generic record operations.
var Cmd = &cobra.Command{
	Use:   "records",
	Short: "Operate on any object records",
	Long:  "Generic CRUD and batch operations for any standard or custom object.",
}

var noResolve bool

func init() {
	Cmd.PersistentFlags().BoolVar(&noResolve, "no-resolve", false, "skip object name resolution")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(destroyCmd)
	Cmd.AddCommand(restoreCmd)
	Cmd.AddCommand(batchCreateCmd)
	Cmd.AddCommand(batchUpdateCmd)
	Cmd.AddCommand(batchDeleteCmd)
	Cmd.AddCommand(batchDestroyCmd)
	Cmd.AddCommand(batchRestoreCmd)
	Cmd.AddCommand(mergeCmd)
	Cmd.AddCommand(findDuplicatesCmd)
	Cmd.AddCommand(groupByCmd)
	Cmd.AddCommand(exportCmd)
	Cmd.AddCommand(importCmd)
}
