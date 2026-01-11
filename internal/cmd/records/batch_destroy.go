package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	batchDestroyData     string
	batchDestroyDataFile string
	batchDestroyIDs      string
	batchDestroyForce    bool
)

var batchDestroyCmd = &cobra.Command{
	Use:   "batch-destroy <object>",
	Short: "Hard delete records in batch",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatchDestroy,
}

func init() {
	batchDestroyCmd.Flags().StringVarP(&batchDestroyData, "data", "d", "", "JSON array payload")
	batchDestroyCmd.Flags().StringVarP(&batchDestroyDataFile, "file", "f", "", "JSON file payload (use - for stdin)")
	batchDestroyCmd.Flags().StringVar(&batchDestroyIDs, "ids", "", "comma-separated IDs")
	batchDestroyCmd.Flags().BoolVar(&batchDestroyForce, "force", false, "skip confirmation prompt")
	batchDestroyCmd.Flags().BoolVar(&batchDestroyForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runBatchDestroy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	if !batchDestroyForce {
		fmt.Printf("About to batch destroy %s. Use --force to confirm.\n", object)
		return nil
	}

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	payload, err := readBatchIDs(batchDestroyData, batchDestroyDataFile, batchDestroyIDs)
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	// Twenty REST API batch destroy endpoint
	path := "/rest/batch/" + plural + "/destroy"
	raw, err := client.DoRaw(ctx, "DELETE", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
