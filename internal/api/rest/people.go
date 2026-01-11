package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListPeople(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Person], error) {
	path := "/rest/people"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	// Twenty API returns {"data":{"people":[...]}} format
	var apiResp types.PeopleListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	// Convert to generic ListResponse
	result := &types.ListResponse[types.Person]{
		Data:       apiResp.Data.People,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

// GetPersonOptions contains options for getting a person
type GetPersonOptions struct {
	Include []string // e.g., ["company"]
}

func (c *Client) GetPerson(ctx context.Context, id string, opts *GetPersonOptions) (*types.Person, error) {
	path := fmt.Sprintf("/rest/people/%s", id)

	// Twenty API uses depth=1 to include relations, not include parameter
	if opts != nil && len(opts.Include) > 0 {
		path += "?depth=1"
	}

	// Twenty API returns {"data":{"person":{...}}} format
	var apiResp types.PersonResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.Person, nil
}
