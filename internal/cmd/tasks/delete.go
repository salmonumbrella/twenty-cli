package tasks

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
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

	// Note: In a full implementation, you would prompt for confirmation
	// unless --force is provided. For now, we just delete.
	if !forceDelete {
		fmt.Printf("Are you sure you want to delete task %s? Use --force to confirm.\n", id)
		return nil
	}

	client := rt.RESTClient()

	if err := client.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Deleted task: %s\n", id)
	return nil
}
