package records

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var (
	listLimit      int
	listCursor     string
	listAll        bool
	listFilter     string
	listFilterFile string
	listSort       string
	listOrder      string
	listFields     string
	listInclude    string
	listParams     []string
)

var listCmd = &cobra.Command{
	Use:   "list <object>",
	Short: "List records for any object",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 20, "max records to return")
	listCmd.Flags().StringVar(&listCursor, "cursor", "", "pagination cursor")
	listCmd.Flags().BoolVar(&listAll, "all", false, "fetch all records")
	listCmd.Flags().StringVar(&listFilter, "filter", "", "JSON filter")
	listCmd.Flags().StringVar(&listFilterFile, "filter-file", "", "JSON filter file (use - for stdin)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "sort field")
	listCmd.Flags().StringVar(&listOrder, "order", "", "sort order (asc or desc)")
	listCmd.Flags().StringVar(&listFields, "fields", "", "fields to select (comma-separated)")
	listCmd.Flags().StringVar(&listInclude, "include", "", "relations to include (comma-separated)")
	listCmd.Flags().StringArrayVar(&listParams, "param", nil, "additional query params (key=value)")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	params, err := parseQueryParams(listLimit, listCursor, listFilter, listFilterFile, listSort, listOrder, listFields, listInclude, listParams)
	if err != nil {
		return err
	}

	cursor := listCursor
	var all []interface{}
	var lastPageInfo interface{}

	for {
		if cursor != "" {
			params.Set("starting_after", cursor)
		}
		path := buildPath(plural, "")
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		raw, err := client.DoRaw(ctx, "GET", path, nil)
		if err != nil {
			return err
		}

		if !listAll {
			return outputRecords(raw, rt.Output, rt.Query)
		}

		items, pageInfo, err := extractList(raw, plural)
		if err != nil {
			return err
		}
		all = append(all, items...)
		if pageInfo == nil || !pageInfo.HasNextPage {
			lastPageInfo = pageInfo
			break
		}
		cursor = pageInfo.EndCursor
		if cursor == "" {
			break
		}
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			plural: all,
		},
		"totalCount": len(all),
	}
	if lastPageInfo != nil {
		payload["pageInfo"] = lastPageInfo
	}

	return outputRecords(payload, rt.Output, rt.Query)
}

func outputRecords(data interface{}, format, query string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, data, query)
	case "yaml":
		return outfmt.WriteYAML(os.Stdout, data, query)
	case "csv":
		return outfmt.WriteCSVFromJSON(os.Stdout, data)
	default:
		return outfmt.WriteTableFromJSON(os.Stdout, data)
	}
}
