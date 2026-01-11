package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	duplicatesData     string
	duplicatesDataFile string
)

var findDuplicatesCmd = &cobra.Command{
	Use:   "find-duplicates <object>",
	Short: "Find duplicate records",
	Args:  cobra.ExactArgs(1),
	RunE:  runFindDuplicates,
}

func init() {
	findDuplicatesCmd.Flags().StringVarP(&duplicatesData, "data", "d", "", "JSON payload for duplicate detection")
	findDuplicatesCmd.Flags().StringVarP(&duplicatesDataFile, "file", "f", "", "JSON file payload for duplicate detection (use - for stdin)")
}

func runFindDuplicates(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	payload, err := shared.ReadJSONInput(duplicatesData, duplicatesDataFile)
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

	path := buildPath(plural, "find-duplicates")
	raw, err := client.DoRaw(ctx, "POST", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
