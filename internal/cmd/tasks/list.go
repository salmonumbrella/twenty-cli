package tasks

import (
	"context"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listTasks(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Task], error) {
	return client.ListTasks(ctx, opts)
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Task]{
	Use:          "list",
	Short:        "List tasks",
	Long:         "List all tasks in your Twenty workspace.",
	ListFunc:     listTasks,
	TableHeaders: []string{"ID", "TITLE", "STATUS", "DUE AT", "ASSIGNEE ID"},
	TableRow: func(t types.Task) []string {
		id := t.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		assigneeID := t.AssigneeID
		if len(assigneeID) > 8 {
			assigneeID = assigneeID[:8] + "..."
		}
		return []string{id, t.Title, t.Status, t.DueAt, assigneeID}
	},
	CSVHeaders: []string{"id", "title", "status", "dueAt", "assigneeId", "createdAt", "updatedAt"},
	CSVRow: func(t types.Task) []string {
		return []string{
			t.ID,
			t.Title,
			t.Status,
			t.DueAt,
			t.AssigneeID,
			t.CreatedAt.Format("2006-01-02T15:04:05Z"),
			t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	},
})
