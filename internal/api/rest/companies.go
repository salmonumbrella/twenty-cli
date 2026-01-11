package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListCompanies(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Company], error) {
	path := "/rest/companies"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	// Twenty API returns {"data":{"companies":[...]}} format
	var apiResp types.CompaniesListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Company]{
		Data:       apiResp.Data.Companies,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

func (c *Client) GetCompany(ctx context.Context, id string, opts *ListOptions) (*types.Company, error) {
	path := fmt.Sprintf("/rest/companies/%s", id)
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}
	// Twenty API returns {"data":{"company":{...}}} format
	var apiResp types.CompanyResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.Company, nil
}

// LinkInput represents a link field for API input
type LinkInput struct {
	PrimaryLinkLabel string `json:"primaryLinkLabel,omitempty"`
	PrimaryLinkUrl   string `json:"primaryLinkUrl,omitempty"`
}

type CreateCompanyInput struct {
	Name          string     `json:"name"`
	DomainName    *LinkInput `json:"domainName,omitempty"`
	Address       string     `json:"address,omitempty"`
	Employees     int        `json:"employees,omitempty"`
	LinkedinLink  *LinkInput `json:"linkedinLink,omitempty"`
	XLink         *LinkInput `json:"xLink,omitempty"`
	AnnualRevenue int        `json:"annualRevenue,omitempty"`
	IdealCustomer bool       `json:"idealCustomerProfile,omitempty"`
}

func (c *Client) CreateCompany(ctx context.Context, input *CreateCompanyInput) (*types.Company, error) {
	// Twenty API returns {"data":{"createCompany":{...}}} format
	var apiResp types.CreateCompanyResponse
	if err := c.Post(ctx, "/rest/companies", input, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.CreateCompany, nil
}

type UpdateCompanyInput struct {
	Name          *string    `json:"name,omitempty"`
	DomainName    *LinkInput `json:"domainName,omitempty"`
	Address       *string    `json:"address,omitempty"`
	Employees     *int       `json:"employees,omitempty"`
	LinkedinLink  *LinkInput `json:"linkedinLink,omitempty"`
	XLink         *LinkInput `json:"xLink,omitempty"`
	AnnualRevenue *int       `json:"annualRevenue,omitempty"`
	IdealCustomer *bool      `json:"idealCustomerProfile,omitempty"`
}

func (c *Client) UpdateCompany(ctx context.Context, id string, input *UpdateCompanyInput) (*types.Company, error) {
	// Twenty API returns {"data":{"updateCompany":{...}}} format
	var apiResp types.UpdateCompanyResponse
	if err := c.Patch(ctx, "/rest/companies/"+id, input, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.UpdateCompany, nil
}

func (c *Client) DeleteCompany(ctx context.Context, id string) error {
	return c.Delete(ctx, "/rest/companies/"+id)
}
