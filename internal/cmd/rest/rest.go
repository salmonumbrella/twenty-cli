package rest

import "github.com/spf13/cobra"

// Cmd is the parent command for raw REST API calls.
var Cmd = &cobra.Command{
	Use:   "rest",
	Short: "Call REST API endpoints",
}

func init() {
	Cmd.AddCommand(requestCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(postCmd)
	Cmd.AddCommand(patchCmd)
	Cmd.AddCommand(deleteCmd)
}
