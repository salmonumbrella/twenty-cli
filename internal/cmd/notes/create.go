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
	createTitle string
	createBody  string
	createData  string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	return builder.NewCreateTypedCommand(builder.CreateTypedConfig[rest.CreateNoteInput, types.Note]{
		Use:   "create",
		Short: "Create a new note",
		BuildInput: func(cmd *cobra.Command) (*rest.CreateNoteInput, error) {
			var input rest.CreateNoteInput
			if createData != "" {
				if err := json.Unmarshal([]byte(createData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			input = rest.CreateNoteInput{
				Title: createTitle,
			}
			if createBody != "" {
				input.BodyV2 = &rest.BodyV2Input{
					Markdown: createBody,
				}
			}
			return &input, nil
		},
		CreateFunc: func(ctx context.Context, client *rest.Client, input *rest.CreateNoteInput) (*types.Note, error) {
			return client.CreateNote(ctx, input)
		},
		OutputText: func(note *types.Note) string {
			return fmt.Sprintf("Created note: %s (%s)", note.ID, note.Title)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&createTitle, "title", "", "note title")
			cmd.Flags().StringVar(&createBody, "body", "", "note body/content")
			cmd.Flags().StringVarP(&createData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := createCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}
