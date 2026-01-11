package records

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <object> <id>",
	Short: "Restore a soft-deleted record",
	Args:  cobra.ExactArgs(2),
	RunE:  runRestore,
}

func runRestore(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]
	id := args[1]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	path := buildPath(plural, id+"/restore")
	raw, err := client.DoRaw(ctx, "POST", path, nil)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
