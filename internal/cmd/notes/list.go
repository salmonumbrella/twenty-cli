package notes

import (
	"context"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listNotes(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Note], error) {
	return client.ListNotes(ctx, opts)
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Note]{
	Use:          "list",
	Short:        "List notes",
	Long:         "List all notes in your Twenty workspace.",
	ListFunc:     listNotes,
	TableHeaders: []string{"ID", "TITLE", "BODY", "CREATED AT"},
	TableRow: func(n types.Note) []string {
		id := n.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		body := n.Body()
		if len(body) > 40 {
			body = body[:40] + "..."
		}
		return []string{id, n.Title, body, n.CreatedAt.Format("2006-01-02")}
	},
	CSVHeaders: []string{"id", "title", "body", "createdAt", "updatedAt"},
	CSVRow: func(n types.Note) []string {
		return []string{
			n.ID,
			n.Title,
			n.Body(),
			n.CreatedAt.Format("2006-01-02T15:04:05Z"),
			n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	},
})
