package people

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var (
	exportFormat string
	exportOutput string
	exportAll    bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export people to file",
	Long:  "Export all people from Twenty CRM to JSON or CSV format.",
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "output format: json or csv")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file (default: stdout)")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "fetch all records (paginate through everything)")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}
	client := rt.RESTClient()

	// Fetch all people with pagination
	var allPeople []types.Person
	cursor := ""
	limit := 100 // Use larger limit for export

	for {
		opts := &rest.ListOptions{
			Limit:  limit,
			Cursor: cursor,
		}
		result, err := client.ListPeople(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list people: %w", err)
		}

		allPeople = append(allPeople, result.Data...)

		if !exportAll || result.PageInfo == nil || !result.PageInfo.HasNextPage {
			break
		}
		cursor = result.PageInfo.EndCursor
	}

	// Determine output writer
	var writer io.Writer = os.Stdout
	if exportOutput != "" {
		f, err := os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = f
	}

	// Format and write output
	switch exportFormat {
	case "csv":
		if err := formatPeopleAsCSV(allPeople, writer); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
	case "json":
		if err := formatPeopleAsJSON(allPeople, writer, rt.Query); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s (use 'json' or 'csv')", exportFormat)
	}

	if exportOutput != "" {
		fmt.Fprintf(os.Stderr, "Exported %d people to %s\n", len(allPeople), exportOutput)
	}

	return nil
}

func formatPeopleAsJSON(people []types.Person, w io.Writer, query string) error {
	return outfmt.WriteJSON(w, people, query)
}

func formatPeopleAsCSV(people []types.Person, w io.Writer) error {
	writer := csv.NewWriter(w)

	// Write header
	header := []string{"ID", "FirstName", "LastName", "Email", "JobTitle", "City", "Phone"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, p := range people {
		row := []string{
			p.ID,
			p.Name.FirstName,
			p.Name.LastName,
			p.Email.PrimaryEmail,
			p.JobTitle,
			p.City,
			p.Phone.PrimaryPhoneNumber,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}
	return nil
}
