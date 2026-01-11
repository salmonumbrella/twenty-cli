package people

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

// peopleFilterFlags holds filter flag values for the list command.
// Using a struct with closures avoids package-level mutable state.
type peopleFilterFlags struct {
	email   string
	name    string
	city    string
	company string
}

func listPeople(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
	return client.ListPeople(ctx, opts)
}

func newListCmd() *cobra.Command {
	flags := &peopleFilterFlags{}

	return builder.NewListCommand(builder.ListConfig[types.Person]{
		Use:      "list",
		Short:    "List people",
		Long:     "List all people in your Twenty workspace.",
		ListFunc: listPeople,
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&flags.email, "email", "", "filter by email (contains)")
			cmd.Flags().StringVar(&flags.name, "name", "", "filter by name (contains)")
			cmd.Flags().StringVar(&flags.city, "city", "", "filter by city")
			cmd.Flags().StringVar(&flags.company, "company-id", "", "filter by company ID")
		},
		BuildFilter: func() map[string]interface{} {
			filter := make(map[string]interface{})
			if flags.email != "" {
				filter["emails"] = map[string]interface{}{
					"primaryEmail": map[string]string{"ilike": "%" + flags.email + "%"},
				}
			}
			if flags.name != "" {
				filter["name"] = map[string]interface{}{
					"firstName": map[string]string{"ilike": "%" + flags.name + "%"},
				}
			}
			if flags.city != "" {
				filter["city"] = map[string]string{"eq": flags.city}
			}
			if flags.company != "" {
				filter["companyId"] = map[string]string{"eq": flags.company}
			}
			return filter
		},
		TableHeaders: []string{"ID", "NAME", "EMAIL", "JOB TITLE"},
		TableRow: func(p types.Person) []string {
			id := p.ID
			if len(id) > 8 {
				id = id[:8] + "..."
			}
			name := fmt.Sprintf("%s %s", p.Name.FirstName, p.Name.LastName)
			return []string{id, name, p.Email.PrimaryEmail, p.JobTitle}
		},
		CSVHeaders: []string{"id", "firstName", "lastName", "email", "phone", "jobTitle", "city", "companyId", "createdAt", "updatedAt"},
		CSVRow: func(p types.Person) []string {
			return []string{
				p.ID,
				p.Name.FirstName,
				p.Name.LastName,
				p.Email.PrimaryEmail,
				p.Phone.PrimaryPhoneNumber,
				p.JobTitle,
				p.City,
				p.CompanyID,
				p.CreatedAt.Format("2006-01-02T15:04:05Z"),
				p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}
		},
	})
}

var listCmd = newListCmd()
