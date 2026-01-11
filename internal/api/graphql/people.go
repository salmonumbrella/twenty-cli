package graphql

import (
	"context"

	"github.com/shurcooL/graphql"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

// UpsertPersonInput is the user-friendly input for upserting a person
type UpsertPersonInput struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	JobTitle  string
	CompanyID string
}

// UpsertPerson creates or updates a person based on email matching
func (c *Client) UpsertPerson(ctx context.Context, input *UpsertPersonInput) (*types.Person, error) {
	var mutation UpsertPersonMutation

	// Build the GraphQL input using PersonCreateInput (Twenty's actual type)
	data := PersonCreateInput{
		Name: &PersonNameInput{
			FirstName: graphql.String(input.FirstName),
			LastName:  graphql.String(input.LastName),
		},
		Emails: &PersonEmailsInput{
			PrimaryEmail: graphql.String(input.Email),
		},
	}

	// Add optional phone if provided
	if input.Phone != "" {
		data.Phones = &PersonPhonesInput{
			PrimaryPhoneNumber: graphql.String(input.Phone),
		}
	}

	// Add optional job title if provided
	if input.JobTitle != "" {
		jobTitle := graphql.String(input.JobTitle)
		data.JobTitle = &jobTitle
	}

	// Add optional company ID if provided
	if input.CompanyID != "" {
		companyID := graphql.String(input.CompanyID)
		data.CompanyID = &companyID
	}

	variables := map[string]interface{}{
		"data": data,
	}

	if err := c.Mutate(ctx, &mutation, variables); err != nil {
		return nil, err
	}

	return &types.Person{
		ID: string(mutation.CreatePerson.ID),
		Name: types.Name{
			FirstName: string(mutation.CreatePerson.Name.FirstName),
			LastName:  string(mutation.CreatePerson.Name.LastName),
		},
		Email: types.Email{
			PrimaryEmail: string(mutation.CreatePerson.Emails.PrimaryEmail),
		},
	}, nil
}
