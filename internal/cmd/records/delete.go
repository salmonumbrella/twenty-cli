package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <object> <id>",
	Short: "Delete a record (soft delete)",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]
	id := args[1]

	if !deleteForce {
		fmt.Printf("About to delete %s %s. Use --force to confirm.\n", object, id)
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

	raw, err := client.DoRaw(ctx, "DELETE", buildPath(plural, id), nil)
	if err != nil {
		return err
	}

	if len(raw) == 0 {
		fmt.Printf("Deleted %s %s\n", object, id)
		return nil
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
