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
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

// ListFunc is the function signature for API methods that list resources.
// It takes a context, REST client, and list options, and returns a paginated
// response containing items of type T.
//
// Example implementation:
//
//	func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
//	    return client.ListPeople(ctx, opts)
//	}
type ListFunc[T any] func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[T], error)

// ListConfig provides all the configuration needed to create a list command
// for resources of type T. It specifies the command metadata, the API function
// to call, and how to format output in different modes (table, CSV, JSON).
type ListConfig[T any] struct {
	// Use is the command name (e.g., "list").
	Use string

	// Short is a brief description shown in help text.
	Short string

	// Long is an extended description shown in detailed help.
	Long string

	// ListFunc is the API function that retrieves the list of resources.
	// This is called for each page when paginating.
	ListFunc ListFunc[T]

	// TableHeaders are the column headers for text/table output format.
	// Example: []string{"ID", "NAME", "EMAIL"}
	TableHeaders []string

	// TableRow converts a single resource into a row of strings for table output.
	// The returned slice must have the same length as TableHeaders.
	TableRow func(T) []string

	// CSVHeaders are optional column headers for CSV output.
	// If nil, TableHeaders is used instead.
	CSVHeaders []string

	// CSVRow is an optional row formatter for CSV output.
	// If nil, TableRow is used instead. This is useful when CSV needs
	// different columns than the table view (e.g., including more fields).
	CSVRow func(T) []string

	// ExtraFlags is an optional function to register resource-specific flags.
	// For example, a people list command might add --email or --company flags.
	ExtraFlags func(cmd *cobra.Command)

	// BuildFilter is an optional function that builds filter conditions from
	// the extra flags. It returns a map that will be merged with any user-provided
	// --filter flag. This enables type-safe filtering via dedicated flags.
	BuildFilter func() map[string]interface{}
}

// NewListCommand creates a new Cobra command for listing resources of type T.
// The returned command includes all standard list flags (pagination, filtering,
// sorting) and supports multiple output formats (text, JSON, YAML, CSV).
//
// When --all is specified, the command automatically paginates through all
// results. Otherwise, it returns a single page based on --limit and --cursor.
//
// The command authenticates using the stored OAuth token and reads configuration
// from viper (base_url, debug, output format).
//
// NewListCommand panics if ListFunc or TableRow is nil.
func NewListCommand[T any](cfg ListConfig[T]) *cobra.Command {
	if cfg.ListFunc == nil {
		panic("builder: ListConfig.ListFunc is required")
	}
	if cfg.TableRow == nil {
		panic("builder: ListConfig.TableRow is required")
	}

	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cfg, opts)
		},
	}

	opts.RegisterFlags(cmd)
	if cfg.ExtraFlags != nil {
		cfg.ExtraFlags(cmd)
	}

	return cmd
}

func runList[T any](cfg ListConfig[T], opts *ListOptions) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	restOpts, err := opts.ToRESTOptions()
	if err != nil {
		return err
	}

	// Apply extra filters if provided
	if cfg.BuildFilter != nil {
		extraFilter := cfg.BuildFilter()
		if restOpts.Filter == nil {
			restOpts.Filter = make(map[string]interface{})
		}
		for k, v := range extraFilter {
			restOpts.Filter[k] = v
		}
	}

	var allData []T
	cursor := opts.Cursor

	for {
		restOpts.Cursor = cursor
		result, err := cfg.ListFunc(ctx, client, restOpts)
		if err != nil {
			return fmt.Errorf("failed to list: %w", err)
		}

		allData = append(allData, result.Data...)

		if !opts.All || result.PageInfo == nil || !result.PageInfo.HasNextPage {
			break
		}
		cursor = result.PageInfo.EndCursor
	}

	// Use CSV-specific config if provided, otherwise fall back to table config
	csvHeaders := cfg.CSVHeaders
	if csvHeaders == nil {
		csvHeaders = cfg.TableHeaders
	}
	csvRow := cfg.CSVRow
	if csvRow == nil {
		csvRow = cfg.TableRow
	}

	return outputList(allData, rt.Output, rt.Query, cfg.TableHeaders, cfg.TableRow, csvHeaders, csvRow)
}

func outputList[T any](data []T, format, query string, headers []string, rowFunc func(T) []string, csvHeaders []string, csvRow func(T) []string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, data, query)
	case "yaml":
		return outfmt.WriteYAML(os.Stdout, data, query)
	case "csv":
		var rows [][]string
		for _, item := range data {
			rows = append(rows, csvRow(item))
		}
		return outfmt.WriteCSV(os.Stdout, csvHeaders, rows)
	default: // text/table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		// Print headers
		for i, h := range headers {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, h)
		}
		fmt.Fprintln(w)
		// Print rows
		for _, item := range data {
			row := rowFunc(item)
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
