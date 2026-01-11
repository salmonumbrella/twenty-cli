package builder

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

// CreateConfig provides configuration for building a create command.
// Unlike [ListConfig] and [GetConfig], create commands are not generic
// because they accept arbitrary JSON data and return raw JSON responses.
//
// The generated command accepts data via --data (inline JSON) or --file
// (path to JSON file, use "-" for stdin).
type CreateConfig struct {
	// Use is the command name (e.g., "create").
	Use string

	// Short is a brief description shown in help text.
	Short string

	// Long is an extended description shown in detailed help.
	Long string

	// Resource is the API resource name used in the REST endpoint.
	// For example, "people" results in POST /rest/people.
	Resource string

	// AllowNonObject allows JSON arrays or scalars as payloads.
	// By default, payloads must be JSON objects.
	AllowNonObject bool

	// MissingPayloadMessage overrides the error when no payload is provided.
	MissingPayloadMessage string

	// InvalidObjectMessage overrides the error when payload is not a JSON object.
	InvalidObjectMessage string

	// JSONOutput forces JSON output (respecting query) instead of raw formats.
	JSONOutput bool

	// SkipErrorWrap returns API errors directly instead of wrapping them.
	SkipErrorWrap bool

	// ExtraFlags is an optional function to register resource-specific flags.
	ExtraFlags func(cmd *cobra.Command)
}

// UpdateConfig provides configuration for building an update command.
// The generated command requires a resource ID as an argument and accepts
// patch data via --data (inline JSON) or --file (path to JSON file).
type UpdateConfig struct {
	// Use is the command name (e.g., "update").
	// The final command will be "Use <id>" (e.g., "update <id>").
	Use string

	// Short is a brief description shown in help text.
	Short string

	// Long is an extended description shown in detailed help.
	Long string

	// Resource is the API resource name used in the REST endpoint.
	// For example, "people" with ID "abc" results in PATCH /rest/people/abc.
	Resource string

	// IDArg customizes the placeholder used in help (default: "id").
	IDArg string

	// AllowNonObject allows JSON arrays or scalars as payloads.
	// By default, payloads must be JSON objects.
	AllowNonObject bool

	// MissingPayloadMessage overrides the error when no payload is provided.
	MissingPayloadMessage string

	// InvalidObjectMessage overrides the error when payload is not a JSON object.
	InvalidObjectMessage string

	// JSONOutput forces JSON output (respecting query) instead of raw formats.
	JSONOutput bool

	// SkipErrorWrap returns API errors directly instead of wrapping them.
	SkipErrorWrap bool

	// ExtraFlags is an optional function to register resource-specific flags.
	ExtraFlags func(cmd *cobra.Command)
}

// DeleteConfig provides configuration for building a delete command.
// The generated command requires a resource ID as an argument and prompts
// for confirmation unless --force is specified.
type DeleteConfig struct {
	// Use is the command name (e.g., "delete").
	// The final command will be "Use <id>" (e.g., "delete <id>").
	Use string

	// Short is a brief description shown in help text.
	Short string

	// Long is an extended description shown in detailed help.
	Long string

	// Resource is the API resource name used in the REST endpoint.
	// For example, "people" with ID "abc" results in DELETE /rest/people/abc.
	Resource string
}

// NewCreateCommand creates a new Cobra command for creating a resource.
// The command accepts JSON data via --data or --file flags and POSTs it
// to the configured resource endpoint.
//
// The command authenticates using the stored OAuth token and reads configuration
// from viper (base_url, debug, output format). The response is output in the
// requested format (text, JSON, YAML, or CSV).
//
// NewCreateCommand panics if Resource is empty.
func NewCreateCommand(cfg CreateConfig) *cobra.Command {
	if cfg.Resource == "" {
		panic("builder: CreateConfig.Resource is required")
	}

	var dataFlag string
	var dataFile string

	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cfg, dataFlag, dataFile)
		},
	}

	cmd.Flags().StringVarP(&dataFlag, "data", "d", "", "JSON data")
	cmd.Flags().StringVarP(&dataFile, "file", "f", "", "JSON file (use - for stdin)")
	if cfg.ExtraFlags != nil {
		cfg.ExtraFlags(cmd)
	}

	return cmd
}

// NewUpdateCommand creates a new Cobra command for updating a resource.
// The command requires a resource ID as an argument and accepts JSON patch
// data via --data or --file flags. It sends a PATCH request to the configured
// resource endpoint.
//
// The command authenticates using the stored OAuth token and reads configuration
// from viper (base_url, debug, output format). The response is output in the
// requested format (text, JSON, YAML, or CSV).
//
// NewUpdateCommand panics if Resource is empty.
func NewUpdateCommand(cfg UpdateConfig) *cobra.Command {
	if cfg.Resource == "" {
		panic("builder: UpdateConfig.Resource is required")
	}

	var dataFlag string
	var dataFile string

	idArg := cfg.IDArg
	if idArg == "" {
		idArg = "id"
	}

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s>", cfg.Use, idArg),
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cfg, args[0], dataFlag, dataFile)
		},
	}

	cmd.Flags().StringVarP(&dataFlag, "data", "d", "", "JSON data")
	cmd.Flags().StringVarP(&dataFile, "file", "f", "", "JSON file (use - for stdin)")
	if cfg.ExtraFlags != nil {
		cfg.ExtraFlags(cmd)
	}

	return cmd
}

// NewDeleteCommand creates a new Cobra command for deleting a resource.
// The command requires a resource ID as an argument and prompts for
// confirmation before proceeding. Use --force to skip the confirmation prompt.
//
// The command authenticates using the stored OAuth token and reads configuration
// from viper (base_url, debug). On success, it prints a confirmation message.
//
// NewDeleteCommand panics if Resource is empty.
func NewDeleteCommand(cfg DeleteConfig) *cobra.Command {
	if cfg.Resource == "" {
		panic("builder: DeleteConfig.Resource is required")
	}

	var force bool

	cmd := &cobra.Command{
		Use:   cfg.Use + " <id>",
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cfg.Resource, args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&force, "yes", false, "skip confirmation prompt (alias for --force)")

	return cmd
}

func runCreate(cfg CreateConfig, dataFlag, dataFile string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	data, err := readJSONPayload(dataFlag, dataFile, payloadOptions{
		allowNonObject:    cfg.AllowNonObject,
		missingPayloadMsg: cfg.MissingPayloadMessage,
		invalidObjectMsg:  cfg.InvalidObjectMessage,
	})
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	path := fmt.Sprintf("/rest/%s", cfg.Resource)

	result, err := client.DoRaw(ctx, "POST", path, data)
	if err != nil {
		if cfg.SkipErrorWrap {
			return err
		}
		return fmt.Errorf("failed to create: %w", err)
	}

	return writeOutput(result, rt.Output, rt.Query, cfg.JSONOutput)
}

func runUpdate(cfg UpdateConfig, id, dataFlag, dataFile string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	data, err := readJSONPayload(dataFlag, dataFile, payloadOptions{
		allowNonObject:    cfg.AllowNonObject,
		missingPayloadMsg: cfg.MissingPayloadMessage,
		invalidObjectMsg:  cfg.InvalidObjectMessage,
	})
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	path := fmt.Sprintf("/rest/%s/%s", cfg.Resource, id)

	result, err := client.DoRaw(ctx, "PATCH", path, data)
	if err != nil {
		if cfg.SkipErrorWrap {
			return err
		}
		return fmt.Errorf("failed to update: %w", err)
	}

	return writeOutput(result, rt.Output, rt.Query, cfg.JSONOutput)
}

func runDelete(resource, id string, force bool) error {
	if !force {
		fmt.Printf("Delete %s %s? [y/N] ", resource, id)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Canceled")
			return nil
		}
	}

	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()
	path := fmt.Sprintf("/rest/%s/%s", resource, id)

	if err := client.Delete(ctx, path); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	fmt.Printf("Deleted %s %s\n", resource, id)
	return nil
}

type payloadOptions struct {
	allowNonObject    bool
	missingPayloadMsg string
	invalidObjectMsg  string
}

func readJSONPayload(dataFlag, dataFile string, opts payloadOptions) (interface{}, error) {
	payload, err := shared.ReadJSONInput(dataFlag, dataFile)
	if err != nil {
		return nil, err
	}
	if payload == nil {
		msg := opts.missingPayloadMsg
		if msg == "" {
			msg = "missing JSON payload; use --data or --file"
		}
		return nil, errors.New(msg)
	}
	if !opts.allowNonObject {
		obj, ok := payload.(map[string]interface{})
		if !ok {
			msg := opts.invalidObjectMsg
			if msg == "" {
				msg = "expected JSON object"
			}
			return nil, errors.New(msg)
		}
		return obj, nil
	}
	return payload, nil
}

func writeOutput(data json.RawMessage, format, query string, jsonOnly bool) error {
	if jsonOnly {
		return shared.WriteJSONOutput(os.Stdout, format, query, data)
	}
	return shared.WriteRawOutput(os.Stdout, format, query, data)
}
