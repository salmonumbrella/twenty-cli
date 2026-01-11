package records

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var (
	exportFormat     string
	exportOutput     string
	exportAll        bool
	exportLimit      int
	exportFilter     string
	exportFilterFile string
	exportParams     []string
	exportSort       string
	exportOrder      string
	exportFields     string
	exportInclude    string
)

var exportCmd = &cobra.Command{
	Use:   "export <object>",
	Short: "Export records to JSON",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "export format (json)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file (default: stdout)")
	exportCmd.Flags().BoolVar(&exportAll, "all", true, "export all records")
	exportCmd.Flags().IntVar(&exportLimit, "limit", 200, "page size for export")
	exportCmd.Flags().StringVar(&exportFilter, "filter", "", "JSON filter")
	exportCmd.Flags().StringVar(&exportFilterFile, "filter-file", "", "JSON filter file (use - for stdin)")
	exportCmd.Flags().StringArrayVar(&exportParams, "param", nil, "additional query params (key=value)")
	exportCmd.Flags().StringVar(&exportSort, "sort", "", "sort field")
	exportCmd.Flags().StringVar(&exportOrder, "order", "", "sort order (asc or desc)")
	exportCmd.Flags().StringVar(&exportFields, "fields", "", "fields to select (comma-separated)")
	exportCmd.Flags().StringVar(&exportInclude, "include", "", "relations to include (comma-separated)")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	if exportFormat != "json" {
		return fmt.Errorf("unsupported export format %q (only json)", exportFormat)
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

	cursor := ""
	var all []interface{}

	for {
		params, err := parseQueryParams(
			exportLimit,
			cursor,
			exportFilter,
			exportFilterFile,
			exportSort,
			exportOrder,
			exportFields,
			exportInclude,
			exportParams,
		)
		if err != nil {
			return err
		}

		path := buildPath(plural, "")
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		raw, err := client.DoRaw(ctx, "GET", path, nil)
		if err != nil {
			return err
		}

		items, pageInfo, err := extractList(raw, plural)
		if err != nil {
			return err
		}
		all = append(all, items...)

		if !exportAll || pageInfo == nil || !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
		if cursor == "" {
			break
		}
	}

	var w io.Writer = os.Stdout
	if exportOutput != "" {
		f, err := os.Create(exportOutput)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	if err := outfmt.WriteJSON(w, all, rt.Query); err != nil {
		return err
	}

	if exportOutput != "" {
		fmt.Fprintf(os.Stderr, "Exported %d records to %s\n", len(all), exportOutput)
	}
	return nil
}
