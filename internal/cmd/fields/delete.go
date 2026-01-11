package fields

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <field-id>",
	Short: "Delete a field",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	fieldID := args[0]

	if !deleteForce {
		fmt.Printf("About to delete field %s. Use --force to confirm.\n", fieldID)
		return nil
	}

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	if err := client.DeleteField(ctx, fieldID); err != nil {
		return err
	}

	fmt.Printf("Deleted field %s\n", fieldID)
	return nil
}
