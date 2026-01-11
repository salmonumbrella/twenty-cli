package people

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var (
	importFormat string
	importDryRun bool
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import people from file",
	Long:  "Import people from a JSON or CSV file into Twenty CRM.",
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func init() {
	importCmd.Flags().StringVar(&importFormat, "format", "", "input format: json or csv (auto-detected from extension if not specified)")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "show what would be imported without actually importing")
}

func runImport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	filename := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	// Open input file
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Detect format
	format := detectFormat(filename, importFormat)
	if format == "" {
		return fmt.Errorf("cannot detect format. Use --format json or --format csv")
	}

	// Parse input
	var inputs []rest.CreatePersonInput
	switch format {
	case "json":
		inputs, err = parseJSONInput(f)
	case "csv":
		inputs, err = parseCSVInput(f)
	default:
		return fmt.Errorf("unsupported format: %s (use 'json' or 'csv')", format)
	}
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if len(inputs) == 0 {
		fmt.Println("No records to import.")
		return nil
	}

	// Dry run mode
	if importDryRun {
		fmt.Print(buildImportSummary(inputs))
		return nil
	}

	// Create each person
	client := rt.RESTClient()
	var created, failed int

	for i, input := range inputs {
		person, err := client.CreatePerson(ctx, &input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create person %d (%s %s): %v\n",
				i+1, input.Name.FirstName, input.Name.LastName, err)
			failed++
			continue
		}
		created++
		if rt.Debug {
			fmt.Printf("Created: %s (%s %s)\n", person.ID, person.Name.FirstName, person.Name.LastName)
		}
	}

	fmt.Printf("Import complete: %d created, %d failed\n", created, failed)
	return nil
}

func detectFormat(filename, explicit string) string {
	if explicit != "" {
		return strings.ToLower(explicit)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return "json"
	case ".csv":
		return "csv"
	default:
		return ""
	}
}

func parseJSONInput(r io.Reader) ([]rest.CreatePersonInput, error) {
	var inputs []rest.CreatePersonInput
	dec := json.NewDecoder(r)
	if err := dec.Decode(&inputs); err != nil {
		return nil, err
	}
	return inputs, nil
}

func parseCSVInput(r io.Reader) ([]rest.CreatePersonInput, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Parse header row to find column indices
	header := records[0]
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Map expected column names (case-insensitive)
	getCol := func(name string, row []string) string {
		if idx, ok := colMap[strings.ToLower(name)]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	var inputs []rest.CreatePersonInput
	for i := 1; i < len(records); i++ {
		row := records[i]
		input := rest.CreatePersonInput{
			Name: types.Name{
				FirstName: getCol("firstname", row),
				LastName:  getCol("lastname", row),
			},
			Email: types.Email{
				PrimaryEmail: getCol("email", row),
			},
			Phone: types.Phone{
				PrimaryPhoneNumber: getCol("phone", row),
			},
			JobTitle: getCol("jobtitle", row),
			City:     getCol("city", row),
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

func buildImportSummary(inputs []rest.CreatePersonInput) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Would import %d people:\n", len(inputs)))
	for i, input := range inputs {
		name := fmt.Sprintf("%s %s", input.Name.FirstName, input.Name.LastName)
		if name == " " {
			name = "(no name)"
		}
		email := input.Email.PrimaryEmail
		if email == "" {
			email = "(no email)"
		}
		sb.WriteString(fmt.Sprintf("  %d. %s <%s>\n", i+1, name, email))
	}
	return sb.String()
}
