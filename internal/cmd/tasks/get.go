package tasks

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
	return builder.NewGetCommand(builder.GetConfig[types.Task]{
		Use:   "get",
		Short: "Get a task by ID",
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Task, error) {
			return client.GetTask(ctx, id)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(t types.Task) [][]string {
			return taskTextRows(t)
		},
		CSVHeaders: []string{"id", "title", "status", "dueAt", "assigneeId", "createdAt", "updatedAt"},
		CSVRows: func(t types.Task) [][]string {
			return [][]string{taskCSVRow(t)}
		},
	})
}

var getCmd = newGetCmd()

func runGet(cmd *cobra.Command, args []string) error {
	if err := getCmd.RunE(getCmd, args); err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	return nil
}

func taskTextRows(t types.Task) [][]string {
	return [][]string{
		{"ID:", t.ID},
		{"Title:", t.Title},
		{"Due At:", t.DueAt},
		{"Status:", t.Status},
		{"Assignee ID:", t.AssigneeID},
		{"Created:", t.CreatedAt.Format("2006-01-02 15:04:05")},
	}
}

func taskCSVRow(t types.Task) []string {
	return []string{
		t.ID,
		t.Title,
		t.Status,
		t.DueAt,
		t.AssigneeID,
		t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func outputTask(t *types.Task, format, query string) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(os.Stdout, t, query)
	case "csv":
		headers := []string{"id", "title", "status", "dueAt", "assigneeId", "createdAt", "updatedAt"}
		rows := [][]string{taskCSVRow(*t)}
		return outfmt.WriteCSV(os.Stdout, headers, rows)
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		rows := taskTextRows(*t)
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
