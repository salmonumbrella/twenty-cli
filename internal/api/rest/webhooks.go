package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListWebhooks(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Webhook], error) {
	path := "/rest/webhooks"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	// Twenty webhooks API returns a plain array, not the standard paginated response
	var webhooks []types.Webhook
	if err := c.Get(ctx, path, &webhooks); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Webhook]{
		Data:       webhooks,
		TotalCount: len(webhooks),
		PageInfo:   nil,
	}
	return result, nil
}

func (c *Client) GetWebhook(ctx context.Context, id string) (*types.Webhook, error) {
	path := fmt.Sprintf("/rest/webhooks/%s", id)
	var result types.Webhook
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type CreateWebhookInput struct {
	TargetURL   string `json:"targetUrl"`
	Operation   string `json:"operation"`
	Description string `json:"description,omitempty"`
	Secret      string `json:"secret,omitempty"`
}

func (c *Client) CreateWebhook(ctx context.Context, input *CreateWebhookInput) (*types.Webhook, error) {
	var result types.Webhook
	if err := c.Post(ctx, "/rest/webhooks", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	return c.Delete(ctx, "/rest/webhooks/"+id)
}
