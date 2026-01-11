package webhooks

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var (
	createURL         string
	createOperation   string
	createDescription string
	createSecret      string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	return builder.NewCreateTypedCommand(builder.CreateTypedConfig[rest.CreateWebhookInput, types.Webhook]{
		Use:   "create",
		Short: "Create a new webhook",
		Long:  "Create a new webhook to receive notifications for CRM events.",
		BuildInput: func(cmd *cobra.Command) (*rest.CreateWebhookInput, error) {
			input := &rest.CreateWebhookInput{
				TargetURL:   createURL,
				Operation:   createOperation,
				Description: createDescription,
				Secret:      createSecret,
			}
			return input, nil
		},
		CreateFunc: func(ctx context.Context, client *rest.Client, input *rest.CreateWebhookInput) (*types.Webhook, error) {
			return client.CreateWebhook(ctx, input)
		},
		OutputText: func(webhook *types.Webhook) string {
			return fmt.Sprintf("Created webhook: %s (%s -> %s)", webhook.ID, webhook.Operation, webhook.TargetURL)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&createURL, "url", "", "target URL for the webhook (required)")
			cmd.Flags().StringVar(&createOperation, "operation", "", "operation to subscribe to (e.g., person.created, *.updated) (required)")
			cmd.Flags().StringVar(&createDescription, "description", "", "description of the webhook")
			cmd.Flags().StringVar(&createSecret, "secret", "", "secret for webhook signature validation (required)")

			_ = cmd.MarkFlagRequired("url")
			_ = cmd.MarkFlagRequired("operation")
			_ = cmd.MarkFlagRequired("secret")
		},
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := createCmd.RunE(cmd, args); err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}
	return nil
}
