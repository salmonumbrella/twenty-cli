package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	mergeData     string
	mergeDataFile string
)

var mergeCmd = &cobra.Command{
	Use:   "merge <object>",
	Short: "Merge records",
	Args:  cobra.ExactArgs(1),
	RunE:  runMerge,
}

func init() {
	mergeCmd.Flags().StringVarP(&mergeData, "data", "d", "", "JSON payload for merge")
	mergeCmd.Flags().StringVarP(&mergeDataFile, "file", "f", "", "JSON file payload for merge (use - for stdin)")
}

func runMerge(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	payload, err := shared.ReadJSONInput(mergeData, mergeDataFile)
	if err != nil {
		return err
	}
	if payload == nil {
		return fmt.Errorf("missing JSON payload; use --data or --file")
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	path := buildPath(plural, "merge")
	raw, err := client.DoRaw(ctx, "POST", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
