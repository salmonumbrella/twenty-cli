package records

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	mergeIDs      string
	mergePriority int
	mergeDryRun   bool
	mergeData     string
	mergeDataFile string
)

var mergeCmd = &cobra.Command{
	Use:   "merge <object>",
	Short: "Merge records",
	Long: `Merge multiple records into one.

The first record in the --ids list (or the record at conflictPriorityIndex) takes priority
when there are conflicting field values.

Examples:
  # Merge three people records, first one takes priority
  twenty records merge people --ids "uuid1,uuid2,uuid3"

  # Merge with second record taking priority on conflicts
  twenty records merge people --ids "uuid1,uuid2,uuid3" --priority 1

  # Preview merge without executing
  twenty records merge people --ids "uuid1,uuid2" --dry-run

  # Use raw JSON payload
  twenty records merge people --data '{"ids":["uuid1","uuid2"],"conflictPriorityIndex":0}'

Payload format (for --data/--file):
  {
    "ids": ["uuid1", "uuid2", "uuid3"],
    "conflictPriorityIndex": 0,
    "dryRun": false
  }`,
	Args: cobra.ExactArgs(1),
	RunE: runMerge,
}

func init() {
	mergeCmd.Flags().StringVar(&mergeIDs, "ids", "", "Comma-separated list of record IDs to merge")
	mergeCmd.Flags().IntVar(&mergePriority, "priority", 0, "Index of record to use for conflict resolution (0-based)")
	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "Preview merge without executing")
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

	var payload json.RawMessage

	// If --ids is provided, build payload from flags
	if mergeIDs != "" {
		ids := strings.Split(mergeIDs, ",")
		for i := range ids {
			ids[i] = strings.TrimSpace(ids[i])
		}

		mergePayload := map[string]interface{}{
			"ids":                   ids,
			"conflictPriorityIndex": mergePriority,
		}
		if mergeDryRun {
			mergePayload["dryRun"] = true
		}

		payload, err = json.Marshal(mergePayload)
		if err != nil {
			return fmt.Errorf("failed to build payload: %w", err)
		}
	} else {
		// Fall back to --data or --file
		rawInput, err := shared.ReadJSONInput(mergeData, mergeDataFile)
		if err != nil {
			return err
		}
		if rawInput == nil {
			return fmt.Errorf("missing payload; use --ids or --data/--file")
		}
		payload, err = json.Marshal(rawInput)
		if err != nil {
			return fmt.Errorf("failed to serialize payload: %w", err)
		}
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	path := buildPath(plural, "merge")
	raw, err := client.DoRaw(ctx, "PATCH", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
