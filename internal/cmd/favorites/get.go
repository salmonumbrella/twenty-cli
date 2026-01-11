package favorites

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a favorite by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	raw, err := client.DoRaw(ctx, "GET", "/rest/favorites/"+id, nil)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
