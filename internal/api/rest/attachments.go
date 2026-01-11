package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListAttachments(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Attachment], error) {
	path := "/rest/attachments"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	var apiResp types.AttachmentsListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Attachment]{
		Data:       apiResp.Data.Attachments,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

func (c *Client) GetAttachment(ctx context.Context, id string) (*types.Attachment, error) {
	path := fmt.Sprintf("/rest/attachments/%s", id)
	var apiResp types.AttachmentResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.Attachment, nil
}
