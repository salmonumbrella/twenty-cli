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
	updateTitle      string
	updateDueAt      string
	updateStatus     string
	updateAssigneeID string
	updateData       string
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cobra.Command {
	return builder.NewUpdateTypedCommand(builder.UpdateTypedConfig[rest.UpdateTaskInput, types.Task]{
		Use:   "update",
		Short: "Update a task",
		BuildInput: func(cmd *cobra.Command, id string) (*rest.UpdateTaskInput, error) {
			var input rest.UpdateTaskInput
			if updateData != "" {
				if err := json.Unmarshal([]byte(updateData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			if cmd.Flags().Changed("title") {
				input.Title = &updateTitle
			}
			if cmd.Flags().Changed("due-at") {
				input.DueAt = &updateDueAt
			}
			if cmd.Flags().Changed("status") {
				input.Status = &updateStatus
			}
			if cmd.Flags().Changed("assignee-id") {
				input.AssigneeID = &updateAssigneeID
			}
			return &input, nil
		},
		UpdateFunc: func(ctx context.Context, client *rest.Client, id string, input *rest.UpdateTaskInput) (*types.Task, error) {
			return client.UpdateTask(ctx, id, input)
		},
		OutputText: func(task *types.Task) string {
			return fmt.Sprintf("Updated task: %s (%s)", task.ID, task.Title)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&updateTitle, "title", "", "task title")
			cmd.Flags().StringVar(&updateDueAt, "due-at", "", "due date (ISO 8601 format)")
			cmd.Flags().StringVar(&updateStatus, "status", "", "task status")
			cmd.Flags().StringVar(&updateAssigneeID, "assignee-id", "", "assignee user ID")
			cmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := updateCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}
