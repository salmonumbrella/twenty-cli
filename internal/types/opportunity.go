package types

import "time"

// Currency represents a monetary amount with currency code
type Currency struct {
	AmountMicros string `json:"amountMicros"`
	CurrencyCode string `json:"currencyCode"`
}

type Opportunity struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Amount           *Currency `json:"amount,omitempty"`
	CloseDate        string    `json:"closeDate,omitempty"`
	Stage            string    `json:"stage,omitempty"`
	Probability      int       `json:"probability,omitempty"`
	CompanyID        string    `json:"companyId,omitempty"`
	PointOfContactID string    `json:"pointOfContactId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	// Relations (populated when depth=1)
	Company        *Company `json:"company,omitempty"`
	PointOfContact *Person  `json:"pointOfContact,omitempty"`
}

// OpportunitiesListResponse represents the API response for listing opportunities
type OpportunitiesListResponse struct {
	Data struct {
		Opportunities []Opportunity `json:"opportunities"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// OpportunityResponse represents the API response for a single opportunity
type OpportunityResponse struct {
	Data struct {
		Opportunity Opportunity `json:"opportunity"`
	} `json:"data"`
}

// CreateOpportunityResponse represents the API response for creating an opportunity
type CreateOpportunityResponse struct {
	Data struct {
		CreateOpportunity Opportunity `json:"createOpportunity"`
	} `json:"data"`
}

// UpdateOpportunityResponse represents the API response for updating an opportunity
type UpdateOpportunityResponse struct {
	Data struct {
		UpdateOpportunity Opportunity `json:"updateOpportunity"`
	} `json:"data"`
}

// DeleteOpportunityResponse represents the API response for deleting an opportunity
type DeleteOpportunityResponse struct {
	Data struct {
		DeleteOpportunity Opportunity `json:"deleteOpportunity"`
	} `json:"data"`
}
