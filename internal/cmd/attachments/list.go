package attachments

import (
	"context"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func listAttachments(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Attachment], error) {
	return client.ListAttachments(ctx, opts)
}

// truncateID truncates an ID for table display
func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}

// truncateName truncates a name for table display
func truncateName(name string) string {
	if len(name) > 30 {
		return name[:30] + "..."
	}
	return name
}

// truncateFullPath truncates a full path for table display
func truncateFullPath(fullPath string) string {
	if len(fullPath) > 40 {
		return fullPath[:40] + "..."
	}
	return fullPath
}

// formatTableRow formats an attachment for table display
func formatTableRow(a types.Attachment) []string {
	return []string{truncateID(a.ID), truncateName(a.Name), a.Type, truncateFullPath(a.FullPath)}
}

// getOptionalString safely extracts a string from a pointer
func getOptionalString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// formatCSVRow formats an attachment for CSV export
func formatCSVRow(a types.Attachment) []string {
	return []string{
		a.ID,
		a.Name,
		a.Type,
		a.FullPath,
		getOptionalString(a.CompanyID),
		getOptionalString(a.PersonID),
		getOptionalString(a.ActivityID),
		getOptionalString(a.TaskID),
		getOptionalString(a.NoteID),
		getOptionalString(a.AuthorID),
		a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		a.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

var listCmd = builder.NewListCommand(builder.ListConfig[types.Attachment]{
	Use:          "list",
	Short:        "List attachments",
	Long:         "List all attachments in your Twenty workspace.",
	ListFunc:     listAttachments,
	TableHeaders: []string{"ID", "NAME", "TYPE", "FULL_PATH"},
	TableRow:     formatTableRow,
	CSVHeaders:   []string{"id", "name", "type", "fullPath", "companyId", "personId", "activityId", "taskId", "noteId", "authorId", "createdAt", "updatedAt"},
	CSVRow:       formatCSVRow,
})
