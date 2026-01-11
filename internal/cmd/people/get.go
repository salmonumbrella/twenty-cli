package people

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var getInclude []string

func newGetCmd() *cobra.Command {
	return builder.NewGetCommand(builder.GetConfig[types.Person]{
		Use:   "get",
		Short: "Get a person by ID",
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			var opts *rest.GetPersonOptions
			if len(getInclude) > 0 {
				opts = &rest.GetPersonOptions{Include: getInclude}
			}
			return client.GetPerson(ctx, id, opts)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringSliceVar(&getInclude, "include", nil, "include relations (e.g., company)")
		},
		TableHeaders: []string{"id", "firstName", "lastName", "email", "phone", "jobTitle", "city", "companyId", "createdAt", "updatedAt"},
		TableRows: func(p types.Person) [][]string {
			return personTextRows(p)
		},
		CSVHeaders: []string{"id", "firstName", "lastName", "email", "phone", "jobTitle", "city", "companyId", "createdAt", "updatedAt"},
		CSVRows: func(p types.Person) [][]string {
			return [][]string{personCSVRow(p)}
		},
	})
}

var getCmd = newGetCmd()

func runGet(cmd *cobra.Command, args []string) error {
	return getCmd.RunE(getCmd, args)
}

func personTextRows(p types.Person) [][]string {
	rows := [][]string{
		{"ID", p.ID},
		{"Name", fmt.Sprintf("%s %s", p.Name.FirstName, p.Name.LastName)},
		{"Email", p.Email.PrimaryEmail},
		{"Job Title", p.JobTitle},
	}
	if p.Company != nil {
		rows = append(rows, []string{"Company", p.Company.Name})
	}
	rows = append(rows, []string{"Created", p.CreatedAt.Format("2006-01-02 15:04:05")})
	return rows
}

func personCSVRow(p types.Person) []string {
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
}

func outputPerson(out io.Writer, p *types.Person, format, query string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(out, p, query)
	case "csv":
		headers := []string{"id", "firstName", "lastName", "email", "phone", "jobTitle", "city", "companyId", "createdAt", "updatedAt"}
		rows := [][]string{personCSVRow(*p)}
		return outfmt.WriteCSV(out, headers, rows)
	default:
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		rows := personTextRows(*p)
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
