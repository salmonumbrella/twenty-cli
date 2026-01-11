package graphql

import "github.com/shurcooL/graphql"

// Common GraphQL types and queries/mutations for Twenty CRM

// PersonName represents the name field structure for a person
type PersonName struct {
	FirstName graphql.String `json:"firstName"`
	LastName  graphql.String `json:"lastName"`
}

// PersonEmails represents the emails field structure for a person
type PersonEmails struct {
	PrimaryEmail graphql.String `json:"primaryEmail"`
}

// PersonPhones represents the phones field structure for a person
type PersonPhones struct {
	PrimaryPhoneNumber      graphql.String `json:"primaryPhoneNumber"`
	PrimaryPhoneCountryCode graphql.String `json:"primaryPhoneCountryCode"`
}

// UpsertPersonMutation represents the mutation to upsert a person using createPerson with upsert=true
type UpsertPersonMutation struct {
	CreatePerson struct {
		ID     graphql.String
		Name   PersonName
		Emails PersonEmails
	} `graphql:"createPerson(data: $data, upsert: true)"`
}

// PersonCreateInput is the input type for creating/upserting a person (matches Twenty's PersonCreateInput)
type PersonCreateInput struct {
	Name      *PersonNameInput   `json:"name,omitempty"`
	Emails    *PersonEmailsInput `json:"emails,omitempty"`
	Phones    *PersonPhonesInput `json:"phones,omitempty"`
	CompanyID *graphql.String    `json:"companyId,omitempty"`
	JobTitle  *graphql.String    `json:"jobTitle,omitempty"`
}

// PersonNameInput is the input type for person name fields
type PersonNameInput struct {
	FirstName graphql.String `json:"firstName"`
	LastName  graphql.String `json:"lastName"`
}

// PersonEmailsInput is the input type for person email fields
type PersonEmailsInput struct {
	PrimaryEmail graphql.String `json:"primaryEmail"`
}

// PersonPhonesInput is the input type for person phone fields
type PersonPhonesInput struct {
	PrimaryPhoneNumber      graphql.String `json:"primaryPhoneNumber"`
	PrimaryPhoneCountryCode graphql.String `json:"primaryPhoneCountryCode"`
}

// UpsertCompanyMutation represents the mutation to upsert a company
type UpsertCompanyMutation struct {
	UpsertCompany struct {
		ID   graphql.String
		Name graphql.String
	} `graphql:"upsertCompany(data: $data)"`
}

// CompanyInput is the input type for creating/upserting a company
type CompanyInput struct {
	Name       graphql.String  `json:"name"`
	DomainName *graphql.String `json:"domainName,omitempty"`
}

// FindManyPeopleQuery represents a query to find multiple people with filters
type FindManyPeopleQuery struct {
	People struct {
		Edges []struct {
			Node struct {
				ID        graphql.String
				Name      PersonName
				Emails    PersonEmails
				CreatedAt graphql.String
				UpdatedAt graphql.String
			}
		}
		PageInfo PageInfo
	} `graphql:"people(filter: $filter, first: $first, after: $after)"`
}

// FindManyCompaniesQuery represents a query to find multiple companies with filters
type FindManyCompaniesQuery struct {
	Companies struct {
		Edges []struct {
			Node struct {
				ID         graphql.String
				Name       graphql.String
				DomainName graphql.String
				CreatedAt  graphql.String
				UpdatedAt  graphql.String
			}
		}
		PageInfo PageInfo
	} `graphql:"companies(filter: $filter, first: $first, after: $after)"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     graphql.Boolean
	HasPreviousPage graphql.Boolean
	StartCursor     graphql.String
	EndCursor       graphql.String
}
