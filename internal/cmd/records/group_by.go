package records

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	groupByData       string
	groupByDataFile   string
	groupByFilter     string
	groupByFilterFile string
	groupByParams     []string
)

var groupByCmd = &cobra.Command{
	Use:   "group-by <object>",
	Short: "Group records",
	Args:  cobra.ExactArgs(1),
	RunE:  runGroupBy,
}

func init() {
	groupByCmd.Flags().StringVarP(&groupByData, "data", "d", "", "JSON payload for group-by")
	groupByCmd.Flags().StringVarP(&groupByDataFile, "file", "f", "", "JSON file payload for group-by (use - for stdin)")
	groupByCmd.Flags().StringVar(&groupByFilter, "filter", "", "JSON filter")
	groupByCmd.Flags().StringVar(&groupByFilterFile, "filter-file", "", "JSON filter file (use - for stdin)")
	groupByCmd.Flags().StringArrayVar(&groupByParams, "param", nil, "additional query params (key=value)")
}

func runGroupBy(cmd *cobra.Command, args []string) error {
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

	payload, err := shared.ReadJSONInput(groupByData, groupByDataFile)
	if err != nil {
		return err
	}

	path := buildPath(plural, "group-by")
	method := "GET"
	var body interface{}

	if payload != nil {
		method = "POST"
		body = payload
	} else {
		params, err := parseQueryParams(0, "", groupByFilter, groupByFilterFile, "", "", "", "", groupByParams)
		if err != nil {
			return err
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	raw, err := client.DoRaw(ctx, method, path, body)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
