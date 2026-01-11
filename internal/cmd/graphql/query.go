package graphql

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	gqlQuery     string
	gqlFile      string
	gqlVars      string
	gqlVarsFile  string
	gqlOperation string
	gqlEndpoint  string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Run a GraphQL query",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGraphQL()
	},
}

var mutateCmd = &cobra.Command{
	Use:   "mutate",
	Short: "Run a GraphQL mutation",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGraphQL()
	},
}

func init() {
	for _, c := range []*cobra.Command{queryCmd, mutateCmd} {
		c.Flags().StringVarP(&gqlQuery, "query", "q", "", "GraphQL query string")
		c.Flags().StringVarP(&gqlFile, "file", "f", "", "GraphQL query file (use - for stdin)")
		c.Flags().StringVar(&gqlVars, "variables", "", "JSON variables")
		c.Flags().StringVar(&gqlVarsFile, "variables-file", "", "JSON variables file (use - for stdin)")
		c.Flags().StringVar(&gqlOperation, "operation", "", "operation name")
		c.Flags().StringVar(&gqlEndpoint, "endpoint", "graphql", "GraphQL endpoint path (graphql or metadata)")
	}
}

func runGraphQL() error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	gql, err := readQuery()
	if err != nil {
		return err
	}

	vars, err := readVariables()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"query": gql,
	}
	if len(vars) > 0 {
		payload["variables"] = vars
	}
	if gqlOperation != "" {
		payload["operationName"] = gqlOperation
	}

	path := normalizeEndpoint(gqlEndpoint)

	client := rt.RESTClient()
	raw, err := client.DoRaw(ctx, "POST", path, payload)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}

func readQuery() (string, error) {
	if gqlFile != "" {
		var raw []byte
		var err error
		if gqlFile == "-" {
			raw, err = io.ReadAll(os.Stdin)
		} else {
			raw, err = os.ReadFile(gqlFile)
		}
		if err != nil {
			return "", fmt.Errorf("read query file: %w", err)
		}
		return strings.TrimSpace(string(raw)), nil
	}

	if gqlQuery == "" {
		return "", fmt.Errorf("missing GraphQL query; use --query or --file")
	}
	return gqlQuery, nil
}

func readVariables() (map[string]interface{}, error) {
	vars, err := shared.ReadJSONMap(gqlVars, gqlVarsFile)
	if err != nil {
		if strings.Contains(err.Error(), "read json file") {
			return nil, fmt.Errorf("read variables file: %w", err)
		}
		return nil, fmt.Errorf("invalid JSON variables: %w", err)
	}
	return vars, nil
}

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = "graphql"
	}
	if strings.HasPrefix(endpoint, "/") {
		return endpoint
	}
	return "/" + endpoint
}
