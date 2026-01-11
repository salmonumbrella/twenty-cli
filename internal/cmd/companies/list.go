package companies

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listCompanies(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Company], error) {
	return client.ListCompanies(ctx, opts)
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Company]{
	Use:          "list",
	Short:        "List companies",
	Long:         "List all companies in your Twenty workspace.",
	ListFunc:     listCompanies,
	TableHeaders: []string{"ID", "NAME", "DOMAIN", "EMPLOYEES"},
	TableRow: func(c types.Company) []string {
		id := c.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		employees := 0
		if c.Employees != nil {
			employees = *c.Employees
		}
		return []string{id, c.Name, c.DomainName.PrimaryLinkUrl, fmt.Sprintf("%d", employees)}
	},
	CSVHeaders: []string{"id", "name", "domain", "employees", "city", "createdAt", "updatedAt"},
	CSVRow: func(c types.Company) []string {
		employees := 0
		if c.Employees != nil {
			employees = *c.Employees
		}
		return []string{
			c.ID,
			c.Name,
			c.DomainName.PrimaryLinkUrl,
			fmt.Sprintf("%d", employees),
			c.Address.AddressCity,
			c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	},
})
