package records

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	batchRestoreData     string
	batchRestoreDataFile string
	batchRestoreIDs      string
)

var batchRestoreCmd = &cobra.Command{
	Use:   "batch-restore <object>",
	Short: "Restore records in batch",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatchRestore,
}

func init() {
	batchRestoreCmd.Flags().StringVarP(&batchRestoreData, "data", "d", "", "JSON array payload")
	batchRestoreCmd.Flags().StringVarP(&batchRestoreDataFile, "file", "f", "", "JSON file payload (use - for stdin)")
	batchRestoreCmd.Flags().StringVar(&batchRestoreIDs, "ids", "", "comma-separated IDs")
}

func runBatchRestore(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	payload, err := readBatchIDs(batchRestoreData, batchRestoreDataFile, batchRestoreIDs)
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	// Twenty REST API batch restore endpoint
	path := "/rest/batch/" + plural + "/restore"
	raw, err := client.DoRaw(ctx, "POST", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
