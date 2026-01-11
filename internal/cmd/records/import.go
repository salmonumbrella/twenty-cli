package records

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var (
	importData          string
	importBatchSize     int
	importDryRun        bool
	importContinueOnErr bool
)

var importCmd = &cobra.Command{
	Use:   "import <object> [file]",
	Short: "Import records from JSON",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runImport,
}

func init() {
	importCmd.Flags().StringVarP(&importData, "data", "d", "", "JSON array payload")
	importCmd.Flags().IntVar(&importBatchSize, "batch-size", 60, "batch size (max 60)")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "validate without importing")
	importCmd.Flags().BoolVar(&importContinueOnErr, "continue-on-error", false, "continue on batch errors")
}

func runImport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]
	file := ""
	if len(args) > 1 {
		file = args[1]
	}

	if importBatchSize <= 0 {
		importBatchSize = 60
	}
	if importBatchSize > 60 {
		importBatchSize = 60
	}

	payload, err := shared.ReadJSONInput(importData, file)
	if err != nil {
		return err
	}
	if payload == nil {
		return fmt.Errorf("missing JSON payload; provide --data or a file")
	}

	records, ok := payload.([]interface{})
	if !ok {
		return fmt.Errorf("import payload must be a JSON array")
	}

	if importDryRun {
		fmt.Printf("Would import %d records into %s\n", len(records), object)
		return nil
	}

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	path := "/rest/batch/" + plural
	var (
		created int
		errors  []string
	)

	for i := 0; i < len(records); i += importBatchSize {
		end := i + importBatchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		_, err := client.DoRaw(ctx, "POST", path, batch)
		if err != nil {
			errors = append(errors, fmt.Sprintf("batch %d-%d: %v", i+1, end, err))
			if !importContinueOnErr {
				break
			}
			continue
		}
		created += len(batch)
	}

	if rt.Output == "json" {
		result := map[string]interface{}{
			"imported": created,
			"errors":   errors,
		}
		return outfmt.WriteJSON(os.Stdout, result, rt.Query)
	}

	fmt.Printf("Import complete: %d imported", created)
	if len(errors) > 0 {
		fmt.Printf(", %d errors\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return nil
	}
	fmt.Println()
	return nil
}
