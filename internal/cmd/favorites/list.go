package favorites

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listFavorites(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Favorite], error) {
	return client.ListFavorites(ctx, opts)
}

// Helper to get first non-nil ID from the favorite
func getRecordID(f types.Favorite) string {
	switch {
	case f.CompanyID != nil:
		return *f.CompanyID
	case f.PersonID != nil:
		return *f.PersonID
	case f.OpportunityID != nil:
		return *f.OpportunityID
	case f.TaskID != nil:
		return *f.TaskID
	case f.NoteID != nil:
		return *f.NoteID
	case f.ViewID != nil:
		return *f.ViewID
	case f.WorkflowID != nil:
		return *f.WorkflowID
	case f.RocketID != nil:
		return *f.RocketID
	default:
		return ""
	}
}

// Helper to get the type of the favorited record
func getRecordType(f types.Favorite) string {
	switch {
	case f.CompanyID != nil:
		return "company"
	case f.PersonID != nil:
		return "person"
	case f.OpportunityID != nil:
		return "opportunity"
	case f.TaskID != nil:
		return "task"
	case f.NoteID != nil:
		return "note"
	case f.ViewID != nil:
		return "view"
	case f.WorkflowID != nil:
		return "workflow"
	case f.RocketID != nil:
		return "rocket"
	default:
		return ""
	}
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Favorite]{
	Use:          "list",
	Short:        "List favorites",
	Long:         "List all favorites in your Twenty workspace.",
	ListFunc:     listFavorites,
	TableHeaders: []string{"ID", "TYPE", "RECORD_ID", "POSITION"},
	TableRow: func(f types.Favorite) []string {
		id := f.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		recordID := getRecordID(f)
		if len(recordID) > 8 {
			recordID = recordID[:8] + "..."
		}
		return []string{id, getRecordType(f), recordID, fmt.Sprintf("%.0f", f.Position)}
	},
	CSVHeaders: []string{"id", "type", "recordId", "position", "workspaceMemberId", "createdAt", "updatedAt"},
	CSVRow: func(f types.Favorite) []string {
		return []string{
			f.ID,
			getRecordType(f),
			getRecordID(f),
			fmt.Sprintf("%.0f", f.Position),
			f.WorkspaceMemberID,
			f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	},
})
