package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	batchUpdateData     string
	batchUpdateDataFile string
)

var batchUpdateCmd = &cobra.Command{
	Use:   "batch-update <object>",
	Short: "Update records in batch",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatchUpdate,
}

func init() {
	batchUpdateCmd.Flags().StringVarP(&batchUpdateData, "data", "d", "", "JSON array payload")
	batchUpdateCmd.Flags().StringVarP(&batchUpdateDataFile, "file", "f", "", "JSON file payload (use - for stdin)")
}

func runBatchUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	payload, err := shared.ReadJSONInput(batchUpdateData, batchUpdateDataFile)
	if err != nil {
		return err
	}
	if payload == nil {
		return fmt.Errorf("missing JSON payload; use --data or --file")
	}

	if _, ok := payload.([]interface{}); !ok {
		return fmt.Errorf("batch payload must be a JSON array")
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	path := "/rest/batch/" + plural
	raw, err := client.DoRaw(ctx, "PATCH", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
