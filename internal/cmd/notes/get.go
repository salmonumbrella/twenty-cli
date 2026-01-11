package notes

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
	return builder.NewGetCommand(builder.GetConfig[types.Note]{
		Use:   "get",
		Short: "Get a note by ID",
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Note, error) {
			return client.GetNote(ctx, id)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(n types.Note) [][]string {
			return noteTextRows(n)
		},
		CSVHeaders: []string{"id", "title", "body", "createdAt", "updatedAt"},
		CSVRows: func(n types.Note) [][]string {
			return [][]string{noteCSVRow(n)}
		},
	})
}

var getCmd = newGetCmd()

func runGet(cmd *cobra.Command, args []string) error {
	if err := getCmd.RunE(getCmd, args); err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	return nil
}

func noteTextRows(n types.Note) [][]string {
	return [][]string{
		{"ID:", n.ID},
		{"Title:", n.Title},
		{"Body:", n.Body()},
		{"Created:", n.CreatedAt.Format("2006-01-02 15:04:05")},
		{"Updated:", n.UpdatedAt.Format("2006-01-02 15:04:05")},
	}
}

func noteCSVRow(n types.Note) []string {
	return []string{
		n.ID,
		n.Title,
		n.Body(),
		n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func outputNote(n *types.Note, format, query string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, n, query)
	case "csv":
		headers := []string{"id", "title", "body", "createdAt", "updatedAt"}
		rows := [][]string{noteCSVRow(*n)}
		return outfmt.WriteCSV(os.Stdout, headers, rows)
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		rows := noteTextRows(*n)
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
