package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListNotes(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Note], error) {
	path := "/rest/notes"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	var apiResp types.NotesListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Note]{
		Data:       apiResp.Data.Notes,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

func (c *Client) GetNote(ctx context.Context, id string) (*types.Note, error) {
	path := fmt.Sprintf("/rest/notes/%s", id)

	var apiResp types.NoteResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.Note, nil
}

// BodyV2Input represents the body content for creating/updating notes
type BodyV2Input struct {
	Markdown string `json:"markdown,omitempty"`
}

type CreateNoteInput struct {
	Title  string       `json:"title,omitempty"`
	BodyV2 *BodyV2Input `json:"bodyV2,omitempty"`
}

func (c *Client) CreateNote(ctx context.Context, input *CreateNoteInput) (*types.Note, error) {
	path := "/rest/notes"

	var apiResp types.CreateNoteResponse
	if err := c.Post(ctx, path, input, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.CreateNote, nil
}

type UpdateNoteInput struct {
	Title  *string      `json:"title,omitempty"`
	BodyV2 *BodyV2Input `json:"bodyV2,omitempty"`
}

func (c *Client) UpdateNote(ctx context.Context, id string, input *UpdateNoteInput) (*types.Note, error) {
	path := fmt.Sprintf("/rest/notes/%s", id)

	var apiResp types.UpdateNoteResponse
	if err := c.Patch(ctx, path, input, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.UpdateNote, nil
}

func (c *Client) DeleteNote(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/notes/%s", id)

	var apiResp types.DeleteNoteResponse
	if err := c.do(ctx, "DELETE", path, nil, &apiResp); err != nil {
		return err
	}

	return nil
}
