package webhooks

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&forceDelete, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	if !forceDelete {
		fmt.Printf("Are you sure you want to delete webhook %s? Use --force to confirm.\n", id)
		return nil
	}

	client := rt.RESTClient()

	if err := client.DeleteWebhook(ctx, id); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("Deleted webhook: %s\n", id)
	return nil
}
