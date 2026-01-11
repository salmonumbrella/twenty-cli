package opportunities

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listOpportunities(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Opportunity], error) {
	return client.ListOpportunities(ctx, opts)
}

func formatCurrency(c *types.Currency) string {
	if c == nil {
		return "-"
	}
	// Convert micros to dollars (divide by 1,000,000)
	// AmountMicros is a string, so we need to parse it
	var amount float64
	if _, err := fmt.Sscanf(c.AmountMicros, "%f", &amount); err == nil {
		amount = amount / 1000000
		return fmt.Sprintf("%.2f %s", amount, c.CurrencyCode)
	}
	return fmt.Sprintf("%s %s", c.AmountMicros, c.CurrencyCode)
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Opportunity]{
	Use:          "list",
	Short:        "List opportunities",
	Long:         "List all opportunities in your Twenty workspace.",
	ListFunc:     listOpportunities,
	TableHeaders: []string{"ID", "NAME", "STAGE", "AMOUNT", "CLOSE DATE"},
	TableRow: func(o types.Opportunity) []string {
		id := o.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		return []string{id, o.Name, o.Stage, formatCurrency(o.Amount), o.CloseDate}
	},
	CSVHeaders: []string{"id", "name", "stage", "amount", "probability", "closeDate", "createdAt", "updatedAt"},
	CSVRow: func(o types.Opportunity) []string {
		return []string{
			o.ID,
			o.Name,
			o.Stage,
			formatCurrency(o.Amount),
			fmt.Sprintf("%d", o.Probability),
			o.CloseDate,
			o.CreatedAt.Format("2006-01-02T15:04:05Z"),
			o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	},
})
