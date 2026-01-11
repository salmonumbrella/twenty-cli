package records

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	batchDeleteData     string
	batchDeleteDataFile string
	batchDeleteIDs      string
	batchDeleteForce    bool
)

var batchDeleteCmd = &cobra.Command{
	Use:   "batch-delete <object>",
	Short: "Delete records in batch",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatchDelete,
}

func init() {
	batchDeleteCmd.Flags().StringVarP(&batchDeleteData, "data", "d", "", "JSON array payload")
	batchDeleteCmd.Flags().StringVarP(&batchDeleteDataFile, "file", "f", "", "JSON file payload (use - for stdin)")
	batchDeleteCmd.Flags().StringVar(&batchDeleteIDs, "ids", "", "comma-separated IDs")
	batchDeleteCmd.Flags().BoolVar(&batchDeleteForce, "force", false, "skip confirmation prompt")
	batchDeleteCmd.Flags().BoolVar(&batchDeleteForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runBatchDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	if !batchDeleteForce {
		fmt.Printf("About to batch delete %s. Use --force to confirm.\n", object)
		return nil
	}

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	ids, err := readBatchIDStrings(batchDeleteData, batchDeleteDataFile, batchDeleteIDs)
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	// Twenty REST API requires filter in query params for batch delete
	// Format: filter=id[in]:[id1,id2,id3]
	path := "/rest/batch/" + plural + "?filter=" + url.QueryEscape("id[in]:["+strings.Join(ids, ",")+"]")
	raw, err := client.DoRaw(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}

func readBatchIDs(data, file, ids string) (interface{}, error) {
	if ids != "" {
		parts := strings.Split(ids, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no valid IDs provided")
		}
		return out, nil
	}

	payload, err := shared.ReadJSONInput(data, file)
	if err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, fmt.Errorf("missing JSON payload; use --data, --file, or --ids")
	}

	if _, ok := payload.([]interface{}); !ok {
		return nil, fmt.Errorf("batch payload must be a JSON array")
	}
	return payload, nil
}

// readBatchIDStrings returns a slice of string IDs for batch operations
func readBatchIDStrings(data, file, ids string) ([]string, error) {
	if ids != "" {
		parts := strings.Split(ids, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no valid IDs provided")
		}
		return out, nil
	}

	payload, err := shared.ReadJSONInput(data, file)
	if err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, fmt.Errorf("missing JSON payload; use --data, --file, or --ids")
	}

	arr, ok := payload.([]interface{})
	if !ok {
		return nil, fmt.Errorf("batch payload must be a JSON array")
	}

	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		} else {
			return nil, fmt.Errorf("batch payload must be a JSON array of strings")
		}
	}
	return out, nil
}
