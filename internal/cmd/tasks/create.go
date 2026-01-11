package tasks

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
	createTitle      string
	createDueAt      string
	createStatus     string
	createAssigneeID string
	createData       string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	return builder.NewCreateTypedCommand(builder.CreateTypedConfig[rest.CreateTaskInput, types.Task]{
		Use:   "create",
		Short: "Create a new task",
		BuildInput: func(cmd *cobra.Command) (*rest.CreateTaskInput, error) {
			var input rest.CreateTaskInput
			if createData != "" {
				if err := json.Unmarshal([]byte(createData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			input = rest.CreateTaskInput{
				Title:      createTitle,
				DueAt:      createDueAt,
				Status:     createStatus,
				AssigneeID: createAssigneeID,
			}
			return &input, nil
		},
		CreateFunc: func(ctx context.Context, client *rest.Client, input *rest.CreateTaskInput) (*types.Task, error) {
			return client.CreateTask(ctx, input)
		},
		OutputText: func(task *types.Task) string {
			return fmt.Sprintf("Created task: %s (%s)", task.ID, task.Title)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&createTitle, "title", "", "task title")
			cmd.Flags().StringVar(&createDueAt, "due-at", "", "due date (ISO 8601 format)")
			cmd.Flags().StringVar(&createStatus, "status", "", "task status")
			cmd.Flags().StringVar(&createAssigneeID, "assignee-id", "", "assignee user ID")
			cmd.Flags().StringVarP(&createData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := createCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}
