package companies

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var getInclude string

func newGetCmd() *cobra.Command {
	return builder.NewGetCommand(builder.GetConfig[types.Company]{
		Use:   "get",
		Short: "Get a company by ID",
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Company, error) {
			opts := &rest.ListOptions{
				Include: strings.FieldsFunc(getInclude, func(r rune) bool { return r == ',' }),
			}
			return client.GetCompany(ctx, id, opts)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&getInclude, "include", "", "relations to include (comma-separated)")
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(c types.Company) [][]string {
			return companyTextRows(c)
		},
		CSVHeaders: []string{"id", "name", "domain", "employees", "city", "createdAt", "updatedAt"},
		CSVRows: func(c types.Company) [][]string {
			return [][]string{companyCSVRow(c)}
		},
	})
}

var getCmd = newGetCmd()

func runGet(cmd *cobra.Command, args []string) error {
	if err := getCmd.RunE(getCmd, args); err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}
	return nil
}

func companyTextRows(c types.Company) [][]string {
	employees := ""
	if c.Employees != nil {
		employees = fmt.Sprintf("%d", *c.Employees)
	}
	annualRevenue := ""
	if c.AnnualRevenue != nil {
		annualRevenue = fmt.Sprintf("%d", *c.AnnualRevenue)
	}
	return [][]string{
		{"ID:", c.ID},
		{"Name:", c.Name},
		{"Domain:", fmt.Sprintf("%v", c.DomainName)},
		{"Address:", fmt.Sprintf("%v", c.Address)},
		{"Employees:", employees},
		{"LinkedIn:", fmt.Sprintf("%v", c.LinkedinLink)},
		{"X (Twitter):", fmt.Sprintf("%v", c.XLink)},
		{"Annual Revenue:", annualRevenue},
		{"Ideal Customer:", fmt.Sprintf("%t", c.IdealCustomer)},
		{"Created:", c.CreatedAt.Format("2006-01-02 15:04:05")},
	}
}

func companyCSVRow(c types.Company) []string {
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
}

func outputCompany(c *types.Company, format, query string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, c, query)
	case "csv":
		employees := 0
		if c.Employees != nil {
			employees = *c.Employees
		}
		headers := []string{"id", "name", "domain", "employees", "city", "createdAt", "updatedAt"}
		rows := [][]string{{
			c.ID,
			c.Name,
			c.DomainName.PrimaryLinkUrl,
			fmt.Sprintf("%d", employees),
			c.Address.AddressCity,
			c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}}
		return outfmt.WriteCSV(os.Stdout, headers, rows)
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		rows := companyTextRows(*c)
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
