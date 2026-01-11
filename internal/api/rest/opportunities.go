package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListOpportunities(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Opportunity], error) {
	path := "/rest/opportunities"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	var apiResp types.OpportunitiesListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Opportunity]{
		Data:       apiResp.Data.Opportunities,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

func (c *Client) GetOpportunity(ctx context.Context, id string) (*types.Opportunity, error) {
	path := fmt.Sprintf("/rest/opportunities/%s", id)

	var apiResp types.OpportunityResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.Opportunity, nil
}

type CreateOpportunityInput struct {
	Name             string          `json:"name"`
	Amount           *types.Currency `json:"amount,omitempty"`
	CloseDate        string          `json:"closeDate,omitempty"`
	Stage            string          `json:"stage,omitempty"`
	Probability      int             `json:"probability,omitempty"`
	CompanyID        string          `json:"companyId,omitempty"`
	PointOfContactID string          `json:"pointOfContactId,omitempty"`
}

func (c *Client) CreateOpportunity(ctx context.Context, input *CreateOpportunityInput) (*types.Opportunity, error) {
	path := "/rest/opportunities"

	var apiResp types.CreateOpportunityResponse
	if err := c.Post(ctx, path, input, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.CreateOpportunity, nil
}

type UpdateOpportunityInput struct {
	Name             *string         `json:"name,omitempty"`
	Amount           *types.Currency `json:"amount,omitempty"`
	CloseDate        *string         `json:"closeDate,omitempty"`
	Stage            *string         `json:"stage,omitempty"`
	Probability      *int            `json:"probability,omitempty"`
	CompanyID        *string         `json:"companyId,omitempty"`
	PointOfContactID *string         `json:"pointOfContactId,omitempty"`
}

func (c *Client) UpdateOpportunity(ctx context.Context, id string, input *UpdateOpportunityInput) (*types.Opportunity, error) {
	path := fmt.Sprintf("/rest/opportunities/%s", id)

	var apiResp types.UpdateOpportunityResponse
	if err := c.Patch(ctx, path, input, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.UpdateOpportunity, nil
}

func (c *Client) DeleteOpportunity(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/opportunities/%s", id)

	var apiResp types.DeleteOpportunityResponse
	if err := c.do(ctx, "DELETE", path, nil, &apiResp); err != nil {
		return err
	}

	return nil
}
