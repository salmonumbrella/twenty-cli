package webhooks

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listWebhooks(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Webhook], error) {
	return client.ListWebhooks(ctx, opts)
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Webhook]{
	Use:          "list",
	Short:        "List all webhooks",
	Long:         "List all webhooks configured in your Twenty workspace.",
	ListFunc:     listWebhooks,
	TableHeaders: []string{"ID", "URL", "OPERATION", "ACTIVE"},
	TableRow: func(wh types.Webhook) []string {
		id := wh.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		return []string{id, wh.TargetURL, wh.Operation, fmt.Sprintf("%v", wh.IsActive)}
	},
	CSVHeaders: []string{"id", "targetUrl", "operation", "description", "isActive", "createdAt", "updatedAt"},
	CSVRow: func(wh types.Webhook) []string {
		return []string{
			wh.ID,
			wh.TargetURL,
			wh.Operation,
			wh.Description,
			fmt.Sprintf("%v", wh.IsActive),
			wh.CreatedAt.Format("2006-01-02T15:04:05Z"),
			wh.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	},
})
