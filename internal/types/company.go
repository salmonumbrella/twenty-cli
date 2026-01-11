package types

import "time"

// SecondaryLink represents a secondary link in a Link field
type SecondaryLink struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

// Link represents a link field in Twenty
type Link struct {
	PrimaryLinkLabel string          `json:"primaryLinkLabel"`
	PrimaryLinkUrl   string          `json:"primaryLinkUrl"`
	SecondaryLinks   []SecondaryLink `json:"secondaryLinks"`
}

// Address represents an address in Twenty
type Address struct {
	AddressCity     string `json:"addressCity"`
	AddressStreet1  string `json:"addressStreet1"`
	AddressStreet2  string `json:"addressStreet2"`
	AddressState    string `json:"addressState"`
	AddressCountry  string `json:"addressCountry"`
	AddressPostcode string `json:"addressPostcode"`
}

type Company struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	DomainName    Link      `json:"domainName,omitempty"`
	Address       Address   `json:"address,omitempty"`
	Employees     *int      `json:"employees,omitempty"`
	LinkedinLink  Link      `json:"linkedinLink,omitempty"`
	XLink         Link      `json:"xLink,omitempty"`
	AnnualRevenue *int      `json:"annualRevenue,omitempty"`
	IdealCustomer bool      `json:"idealCustomerProfile,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	// Relations (populated when depth=1)
	People        []Person      `json:"people,omitempty"`
	Opportunities []Opportunity `json:"opportunities,omitempty"`
}

// CompaniesListResponse is the Twenty API response for listing companies
type CompaniesListResponse struct {
	Data struct {
		Companies []Company `json:"companies"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// CompanyResponse is the Twenty API response for a single company
type CompanyResponse struct {
	Data struct {
		Company Company `json:"company"`
	} `json:"data"`
}

// CreateCompanyResponse is the Twenty API response for creating a company
type CreateCompanyResponse struct {
	Data struct {
		CreateCompany Company `json:"createCompany"`
	} `json:"data"`
}

// UpdateCompanyResponse is the Twenty API response for updating a company
type UpdateCompanyResponse struct {
	Data struct {
		UpdateCompany Company `json:"updateCompany"`
	} `json:"data"`
}
