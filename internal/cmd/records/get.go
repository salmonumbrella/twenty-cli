package records

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	getFields  string
	getInclude string
	getParams  []string
)

var getCmd = &cobra.Command{
	Use:   "get <object> <id>",
	Short: "Get a record by ID",
	Args:  cobra.ExactArgs(2),
	RunE:  runGet,
}

func init() {
	getCmd.Flags().StringVar(&getFields, "fields", "", "fields to select (comma-separated)")
	getCmd.Flags().StringVar(&getInclude, "include", "", "relations to include (comma-separated)")
	getCmd.Flags().StringArrayVar(&getParams, "param", nil, "additional query params (key=value)")
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	object := args[0]
	id := args[1]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	params, err := parseQueryParams(0, "", "", "", "", "", getFields, getInclude, getParams)
	if err != nil {
		return err
	}

	path := buildPath(plural, id)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	raw, err := client.DoRaw(ctx, "GET", path, nil)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
