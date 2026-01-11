package opportunities

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func newGetCmd() *cobra.Command {
	return builder.NewGetCommand(builder.GetConfig[types.Opportunity]{
		Use:   "get",
		Short: "Get an opportunity by ID",
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Opportunity, error) {
			return client.GetOpportunity(ctx, id)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(o types.Opportunity) [][]string {
			return opportunityTextRows(o)
		},
		CSVHeaders: []string{"id", "name", "stage", "amount", "probability", "createdAt", "updatedAt"},
		CSVRows: func(o types.Opportunity) [][]string {
			return [][]string{opportunityCSVRow(o)}
		},
	})
}

var getCmd = newGetCmd()

func runGet(cmd *cobra.Command, args []string) error {
	if err := getCmd.RunE(getCmd, args); err != nil {
		return fmt.Errorf("failed to get opportunity: %w", err)
	}
	return nil
}

func opportunityTextRows(o types.Opportunity) [][]string {
	return [][]string{
		{"ID:", o.ID},
		{"Name:", o.Name},
		{"Amount:", formatCurrency(o.Amount)},
		{"Close Date:", o.CloseDate},
		{"Stage:", o.Stage},
		{"Probability:", fmt.Sprintf("%d%%", o.Probability)},
		{"Company ID:", o.CompanyID},
		{"Point of Contact ID:", o.PointOfContactID},
		{"Created:", o.CreatedAt.Format("2006-01-02 15:04:05")},
	}
}

func opportunityCSVRow(o types.Opportunity) []string {
	return []string{
		o.ID,
		o.Name,
		o.Stage,
		formatCurrency(o.Amount),
		fmt.Sprintf("%d", o.Probability),
		o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func outputOpportunity(o *types.Opportunity, format, query string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, o, query)
	case "csv":
		headers := []string{"id", "name", "stage", "amount", "probability", "createdAt", "updatedAt"}
		rows := [][]string{opportunityCSVRow(*o)}
		return outfmt.WriteCSV(os.Stdout, headers, rows)
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		rows := opportunityTextRows(*o)
		for _, row := range rows {
			for i, cell := range row {
				if i > 0 {
					fmt.Fprint(w, "\t")
				}
				fmt.Fprint(w, cell)
			}
			fmt.Fprintln(w)
		}
		return w.Flush()
	}
}
