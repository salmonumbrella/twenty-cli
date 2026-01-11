package people

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var batchCreateFile string

var batchCreateCmd = &cobra.Command{
	Use:   "batch-create",
	Short: "Create multiple people from JSON",
	Long:  "Create multiple people from a JSON array. Reads from file or stdin.",
	RunE:  runBatchCreate,
}

func init() {
	batchCreateCmd.Flags().StringVarP(&batchCreateFile, "file", "f", "", "JSON file with array of people (use - for stdin)")
}

func runBatchCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	// Read input
	var reader io.Reader
	if batchCreateFile == "-" || batchCreateFile == "" {
		reader = os.Stdin
	} else {
		f, err := os.Open(batchCreateFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
		reader = f
	}

	var inputs []rest.CreatePersonInput
	if err := json.NewDecoder(reader).Decode(&inputs); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client := rt.RESTClient()

	created := make([]string, 0, len(inputs))
	var errors []string

	for i, input := range inputs {
		person, err := client.CreatePerson(ctx, &input)
		if err != nil {
			errors = append(errors, fmt.Sprintf("record %d: %v", i, err))
			continue
		}
		created = append(created, person.ID)
	}

	if rt.Output == "json" {
		result := map[string]interface{}{
			"created": created,
			"errors":  errors,
		}
		return outfmt.WriteJSON(os.Stdout, result, rt.Query)
	}

	fmt.Printf("Created %d people\n", len(created))
	if len(errors) > 0 {
		fmt.Printf("Errors (%d):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
	}
	return nil
}
