package graphql

import "github.com/spf13/cobra"

// Cmd is the parent command for raw GraphQL operations.
var Cmd = &cobra.Command{
	Use:   "graphql",
	Short: "Call GraphQL API endpoints",
}

func init() {
	Cmd.AddCommand(queryCmd)
	Cmd.AddCommand(mutateCmd)
	Cmd.AddCommand(schemaCmd)
}
