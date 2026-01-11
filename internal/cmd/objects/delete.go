package objects

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <object-id>",
	Short: "Delete a custom object",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	objectID := args[0]

	if !deleteForce {
		fmt.Printf("About to delete object %s. Use --force to confirm.\n", objectID)
		return nil
	}

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	if err := client.DeleteObject(ctx, objectID); err != nil {
		return err
	}

	fmt.Printf("Deleted object %s\n", objectID)
	return nil
}
