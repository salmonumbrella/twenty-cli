package records

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	batchCreateData     string
	batchCreateDataFile string
)

var batchCreateCmd = &cobra.Command{
	Use:   "batch-create <object>",
	Short: "Create records in batch",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatchCreate,
}

func init() {
	batchCreateCmd.Flags().StringVarP(&batchCreateData, "data", "d", "", "JSON array payload")
	batchCreateCmd.Flags().StringVarP(&batchCreateDataFile, "file", "f", "", "JSON file payload (use - for stdin)")
}

func runBatchCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	payload, err := shared.ReadJSONInput(batchCreateData, batchCreateDataFile)
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
	raw, err := client.DoRaw(ctx, "POST", path, payload)
	if err != nil {
		return err
	}

	return outputRecords(raw, rt.Output, rt.Query)
}
