package people

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:          "delete <id>",
	Short:        "Delete a person",
	Long:         "Delete a person from Twenty CRM. This action is irreversible. Use --force to skip confirmation.",
	Args:         cobra.ExactArgs(1),
	RunE:         runDelete,
	SilenceUsage: true,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&forceDelete, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	if !forceDelete {
		return fmt.Errorf("delete aborted: use --force to confirm deletion of %s", id)
	}

	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	if err := client.DeletePerson(ctx, id); err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	fmt.Printf("Deleted person: %s\n", id)
	return nil
}
