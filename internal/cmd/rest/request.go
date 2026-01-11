package rest

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

var (
	bodyData     string
	bodyDataFile string
)

var requestCmd = &cobra.Command{
	Use:   "request <method> <path>",
	Short: "Send a raw REST request",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRequest(args[0], args[1])
	},
}

var getCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Send a GET request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRequest("GET", args[0])
	},
}

var postCmd = &cobra.Command{
	Use:   "post <path>",
	Short: "Send a POST request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRequest("POST", args[0])
	},
}

var patchCmd = &cobra.Command{
	Use:   "patch <path>",
	Short: "Send a PATCH request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRequest("PATCH", args[0])
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Send a DELETE request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRequest("DELETE", args[0])
	},
}

func init() {
	for _, c := range []*cobra.Command{requestCmd, postCmd, patchCmd, deleteCmd} {
		c.Flags().StringVarP(&bodyData, "data", "d", "", "JSON request body")
		c.Flags().StringVarP(&bodyDataFile, "data-file", "f", "", "JSON file for request body (use - for stdin)")
	}
}

func runRequest(method, path string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	body, err := parseBody()
	if err != nil {
		return err
	}

	path = normalizePath(path)
	raw, err := client.DoRaw(ctx, strings.ToUpper(method), path, body)
	if err != nil {
		return err
	}

	return shared.WriteJSONOutput(os.Stdout, rt.Output, rt.Query, raw)
}

func parseBody() (interface{}, error) {
	body, err := shared.ReadJSONInput(bodyData, bodyDataFile)
	if err != nil {
		if strings.Contains(err.Error(), "read json file") {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		return nil, fmt.Errorf("invalid JSON body: %w", err)
	}
	return body, nil
}

func normalizePath(path string) string {
	if path == "" {
		return "/rest"
	}
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Auto-prepend /rest/ for paths that don't already have a known API prefix
	if !strings.HasPrefix(path, "/rest") && !strings.HasPrefix(path, "/graphql") && !strings.HasPrefix(path, "/metadata") {
		path = "/rest" + path
	}
	return path
}
