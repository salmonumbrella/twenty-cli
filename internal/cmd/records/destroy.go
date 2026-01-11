package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var destroyForce bool

var destroyCmd = &cobra.Command{
	Use:   "destroy <object> <id>",
	Short: "Hard delete a record",
	Args:  cobra.ExactArgs(2),
	RunE:  runDestroy,
}

func init() {
	destroyCmd.Flags().BoolVar(&destroyForce, "force", false, "skip confirmation prompt")
	destroyCmd.Flags().BoolVar(&destroyForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runDestroy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]
	id := args[1]

	if !destroyForce {
		fmt.Printf("About to destroy %s %s. Use --force to confirm.\n", object, id)
		return nil
	}

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	path := buildPath(plural, id+"/destroy")
	raw, err := client.DoRaw(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	if len(raw) == 0 {
		fmt.Printf("Destroyed %s %s\n", object, id)
		return nil
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
