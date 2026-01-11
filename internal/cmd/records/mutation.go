package records

import (
	"context"
	"os"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

func runMutation(method, object, id, data, dataFile string, set []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	body, err := parseBody(data, dataFile, set)
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	plural, err := resolveObject(ctx, client, object, noResolve)
	if err != nil {
		return err
	}

	raw, err := client.DoRaw(ctx, method, buildPath(plural, id), body)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}
