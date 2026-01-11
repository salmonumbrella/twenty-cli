package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

// CreateTypedConfig configures a typed create command.
type CreateTypedConfig[In any, Out any] struct {
	Use   string
	Short string
	Long  string

	BuildInput func(cmd *cobra.Command) (*In, error)
	CreateFunc func(ctx context.Context, client *rest.Client, input *In) (*Out, error)
	OutputText func(*Out) string
	ExtraFlags func(cmd *cobra.Command)
}

// UpdateTypedConfig configures a typed update command.
type UpdateTypedConfig[In any, Out any] struct {
	Use   string
	Short string
	Long  string

	BuildInput func(cmd *cobra.Command, id string) (*In, error)
	// BuildInputWithClient can be used when input building depends on API reads.
	BuildInputWithClient func(ctx context.Context, client *rest.Client, cmd *cobra.Command, id string) (*In, error)
	UpdateFunc           func(ctx context.Context, client *rest.Client, id string, input *In) (*Out, error)
	OutputText           func(*Out) string
	ExtraFlags           func(cmd *cobra.Command)
}

// NewCreateTypedCommand creates a new create command that uses typed input/output.
func NewCreateTypedCommand[In any, Out any](cfg CreateTypedConfig[In, Out]) *cobra.Command {
	if cfg.BuildInput == nil {
		panic("builder: CreateTypedConfig.BuildInput is required")
	}
	if cfg.CreateFunc == nil {
		panic("builder: CreateTypedConfig.CreateFunc is required")
	}

	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateTyped(cfg, cmd)
		},
	}

	if cfg.ExtraFlags != nil {
		cfg.ExtraFlags(cmd)
	}

	return cmd
}

// NewUpdateTypedCommand creates a new update command that uses typed input/output.
func NewUpdateTypedCommand[In any, Out any](cfg UpdateTypedConfig[In, Out]) *cobra.Command {
	if cfg.BuildInput == nil && cfg.BuildInputWithClient == nil {
		panic("builder: UpdateTypedConfig.BuildInput or BuildInputWithClient is required")
	}
	if cfg.UpdateFunc == nil {
		panic("builder: UpdateTypedConfig.UpdateFunc is required")
	}

	cmd := &cobra.Command{
		Use:   cfg.Use + " <id>",
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateTyped(cfg, cmd, args[0])
		},
	}

	if cfg.ExtraFlags != nil {
		cfg.ExtraFlags(cmd)
	}

	return cmd
}

func runCreateTyped[In any, Out any](cfg CreateTypedConfig[In, Out], cmd *cobra.Command) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	input, err := cfg.BuildInput(cmd)
	if err != nil {
		return err
	}

	result, err := cfg.CreateFunc(ctx, rt.RESTClient(), input)
	if err != nil {
		return err
	}

	if rt.Output == "json" {
		return outfmt.WriteJSON(os.Stdout, result, rt.Query)
	}
	if cfg.OutputText != nil {
		fmt.Fprintln(os.Stdout, cfg.OutputText(result))
		return nil
	}
	return outfmt.WriteJSON(os.Stdout, result, rt.Query)
}

func runUpdateTyped[In any, Out any](cfg UpdateTypedConfig[In, Out], cmd *cobra.Command, id string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	var input *In
	if cfg.BuildInputWithClient != nil {
		input, err = cfg.BuildInputWithClient(ctx, client, cmd, id)
	} else {
		input, err = cfg.BuildInput(cmd, id)
	}
	if err != nil {
		return err
	}

	result, err := cfg.UpdateFunc(ctx, client, id, input)
	if err != nil {
		return err
	}

	if rt.Output == "json" {
		return outfmt.WriteJSON(os.Stdout, result, rt.Query)
	}
	if cfg.OutputText != nil {
		fmt.Fprintln(os.Stdout, cfg.OutputText(result))
		return nil
	}
	return outfmt.WriteJSON(os.Stdout, result, rt.Query)
}
