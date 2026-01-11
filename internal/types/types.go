package types

import "time"

type Name struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Email struct {
	PrimaryEmail     string   `json:"primaryEmail"`
	AdditionalEmails []string `json:"additionalEmails,omitempty"`
}

type Phone struct {
	PrimaryPhoneNumber     string   `json:"primaryPhoneNumber"`
	AdditionalPhoneNumbers []string `json:"additionalPhoneNumbers,omitempty"`
}

type Person struct {
	ID        string    `json:"id"`
	Name      Name      `json:"name"`
	Email     Email     `json:"emails"`
	Phone     Phone     `json:"phones,omitempty"`
	JobTitle  string    `json:"jobTitle,omitempty"`
	City      string    `json:"city,omitempty"`
	CompanyID string    `json:"companyId,omitempty"`
	Company   *Company  `json:"company,omitempty"` // Included relation
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListResponse is the generic response wrapper
// Note: Twenty API returns {"data":{"resourceName":[...]}} so we need custom handling
type ListResponse[T any] struct {
	Data       []T       `json:"-"` // Will be populated manually
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// PeopleListResponse is the specific response for people list
type PeopleListResponse struct {
	Data struct {
		People []Person `json:"people"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// PersonResponse is the specific response for single person
type PersonResponse struct {
	Data struct {
		Person Person `json:"person"`
	} `json:"data"`
}

// CreatePersonResponse is the response for creating a person
type CreatePersonResponse struct {
	Data struct {
		CreatePerson Person `json:"createPerson"`
	} `json:"data"`
}

// UpdatePersonResponse is the response for updating a person
type UpdatePersonResponse struct {
	Data struct {
		UpdatePerson Person `json:"updatePerson"`
	} `json:"data"`
}

// DeletePersonResponse is the response for deleting a person
type DeletePersonResponse struct {
	Data struct {
		DeletePerson Person `json:"deletePerson"`
	} `json:"data"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}
