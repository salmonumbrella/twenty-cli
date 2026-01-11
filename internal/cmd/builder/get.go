package builder

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

// GetFunc is the function signature for API methods that retrieve a single resource.
// It takes a context, REST client, and resource ID, and returns a pointer to
// the resource of type T.
//
// Example implementation:
//
//	func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
//	    return client.GetPerson(ctx, id)
//	}
type GetFunc[T any] func(ctx context.Context, client *rest.Client, id string) (*T, error)

// GetConfig provides all the configuration needed to create a get command
// for retrieving a single resource of type T by its ID. Unlike [ListConfig],
// the table output displays the resource as key-value pairs (one row per field)
// rather than as a single row in a table.
type GetConfig[T any] struct {
	// Use is the command name (e.g., "get").
	// The final command will be "Use <id>" (e.g., "get <id>").
	Use string

	// Short is a brief description shown in help text.
	Short string

	// Long is an extended description shown in detailed help.
	Long string

	// GetFunc is the API function that retrieves a single resource by ID.
	GetFunc GetFunc[T]

	// TableHeaders are the column headers for text/table output format.
	// Typically ["FIELD", "VALUE"] for key-value display.
	TableHeaders []string

	// TableRows converts a single resource into multiple rows for table output.
	// Each inner slice represents one row (e.g., one field). This differs from
	// [ListConfig.TableRow] which returns a single row per resource.
	//
	// Example:
	//
	//	func(p types.Person) [][]string {
	//	    return [][]string{
	//	        {"ID", p.ID},
	//	        {"Name", p.Name.FirstName},
	//	        {"Email", p.Email},
	//	    }
	//	}
	TableRows func(T) [][]string

	// CSVHeaders are optional column headers for CSV output.
	// If nil, TableHeaders is used instead.
	CSVHeaders []string

	// CSVRows converts a single resource into rows for CSV output.
	// If nil, TableRows is used instead.
	CSVRows func(T) [][]string

	// ExtraFlags is an optional function to register resource-specific flags.
	ExtraFlags func(cmd *cobra.Command)
}

// NewGetCommand creates a new Cobra command for retrieving a single resource
// of type T by its ID. The command requires exactly one argument (the resource ID)
// and supports multiple output formats (text, JSON, YAML, CSV).
//
// The command authenticates using the stored OAuth token and reads configuration
// from viper (base_url, debug, output format).
//
// NewGetCommand panics if GetFunc or TableRows is nil.
func NewGetCommand[T any](cfg GetConfig[T]) *cobra.Command {
	if cfg.GetFunc == nil {
		panic("builder: GetConfig.GetFunc is required")
	}
	if cfg.TableRows == nil {
		panic("builder: GetConfig.TableRows is required")
	}

	cmd := &cobra.Command{
		Use:   cfg.Use + " <id>",
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cfg, args[0])
		},
	}

	if cfg.ExtraFlags != nil {
		cfg.ExtraFlags(cmd)
	}

	return cmd
}

func runGet[T any](cfg GetConfig[T], id string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	result, err := cfg.GetFunc(ctx, client, id)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	return outputGet(*result, rt.Output, rt.Query, cfg.TableHeaders, cfg.TableRows, cfg.CSVHeaders, cfg.CSVRows)
}

func outputGet[T any](data T, format, query string, headers []string, rowsFunc func(T) [][]string, csvHeaders []string, csvRows func(T) [][]string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, data, query)
	case "yaml":
		return outfmt.WriteYAML(os.Stdout, data, query)
	case "csv":
		if csvHeaders == nil {
			csvHeaders = headers
		}
		if csvRows == nil {
			csvRows = rowsFunc
		}
		rows := csvRows(data)
		return outfmt.WriteCSV(os.Stdout, csvHeaders, rows)
	default: // text/table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		rows := rowsFunc(data)
		for _, row := range rows {
			for i, cell := range row {
				if i > 0 {
					fmt.Fprint(w, "\t")
				}
				fmt.Fprint(w, cell)
			}
			fmt.Fprintln(w)
		}
		return w.Flush()
	}
}
