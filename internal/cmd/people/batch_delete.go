package people

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var (
	batchDeleteFile  string
	batchDeleteForce bool
)

var batchDeleteCmd = &cobra.Command{
	Use:   "batch-delete",
	Short: "Delete multiple people by ID",
	Long:  "Delete multiple people from a JSON array of IDs. Reads from file or stdin.",
	RunE:  runBatchDelete,
}

func init() {
	batchDeleteCmd.Flags().StringVarP(&batchDeleteFile, "file", "f", "", "JSON file with array of IDs (use - for stdin)")
	batchDeleteCmd.Flags().BoolVar(&batchDeleteForce, "force", false, "skip confirmation prompt")
	batchDeleteCmd.Flags().BoolVar(&batchDeleteForce, "yes", false, "skip confirmation prompt (alias for --force)")
}

func runBatchDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	// Read input
	var reader io.Reader
	if batchDeleteFile == "-" || batchDeleteFile == "" {
		reader = os.Stdin
	} else {
		f, err := os.Open(batchDeleteFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
		reader = f
	}

	var ids []string
	if err := json.NewDecoder(reader).Decode(&ids); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if !batchDeleteForce {
		fmt.Printf("About to delete %d people. Use --force to confirm.\n", len(ids))
		return nil
	}

	client := rt.RESTClient()

	deleted := make([]string, 0, len(ids))
	var errors []string

	for _, id := range ids {
		if err := client.DeletePerson(ctx, id); err != nil {
			errors = append(errors, fmt.Sprintf("ID %s: %v", id, err))
			continue
		}
		deleted = append(deleted, id)
	}

	if rt.Output == "json" {
		result := map[string]interface{}{
			"deleted": deleted,
			"errors":  errors,
		}
		return outfmt.WriteJSON(os.Stdout, result, rt.Query)
	}

	fmt.Printf("Deleted %d people\n", len(deleted))
	if len(errors) > 0 {
		fmt.Printf("Errors (%d):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
	}
	return nil
}
