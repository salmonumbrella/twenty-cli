package notes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var (
	updateTitle string
	updateBody  string
	updateData  string
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cobra.Command {
	return builder.NewUpdateTypedCommand(builder.UpdateTypedConfig[rest.UpdateNoteInput, types.Note]{
		Use:   "update",
		Short: "Update a note",
		BuildInput: func(cmd *cobra.Command, id string) (*rest.UpdateNoteInput, error) {
			var input rest.UpdateNoteInput
			if updateData != "" {
				if err := json.Unmarshal([]byte(updateData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			if cmd.Flags().Changed("title") {
				input.Title = &updateTitle
			}
			if cmd.Flags().Changed("body") {
				input.BodyV2 = &rest.BodyV2Input{
					Markdown: updateBody,
				}
			}
			return &input, nil
		},
		UpdateFunc: func(ctx context.Context, client *rest.Client, id string, input *rest.UpdateNoteInput) (*types.Note, error) {
			return client.UpdateNote(ctx, id, input)
		},
		OutputText: func(note *types.Note) string {
			return fmt.Sprintf("Updated note: %s (%s)", note.ID, note.Title)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&updateTitle, "title", "", "note title")
			cmd.Flags().StringVar(&updateBody, "body", "", "note body/content")
			cmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := updateCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to update note: %w", err)
	}
	return nil
}
