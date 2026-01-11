package rest

import (
	"context"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

type CreatePersonInput struct {
	Name     types.Name  `json:"name"`
	Email    types.Email `json:"emails,omitempty"`
	Phone    types.Phone `json:"phones,omitempty"`
	JobTitle string      `json:"jobTitle,omitempty"`
	City     string      `json:"city,omitempty"`
}

func (c *Client) CreatePerson(ctx context.Context, input *CreatePersonInput) (*types.Person, error) {
	// Twenty API returns {"data":{"createPerson":{...}}} format
	var apiResp types.CreatePersonResponse
	if err := c.Post(ctx, "/rest/people", input, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.CreatePerson, nil
}

type UpdatePersonInput struct {
	Name     *types.Name  `json:"name,omitempty"`
	Email    *types.Email `json:"emails,omitempty"`
	Phone    *types.Phone `json:"phones,omitempty"`
	JobTitle *string      `json:"jobTitle,omitempty"`
	City     *string      `json:"city,omitempty"`
}

func (c *Client) UpdatePerson(ctx context.Context, id string, input *UpdatePersonInput) (*types.Person, error) {
	// Twenty API returns {"data":{"updatePerson":{...}}} format
	var apiResp types.UpdatePersonResponse
	if err := c.Patch(ctx, "/rest/people/"+id, input, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.UpdatePerson, nil
}

func (c *Client) DeletePerson(ctx context.Context, id string) error {
	return c.Delete(ctx, "/rest/people/"+id)
}
