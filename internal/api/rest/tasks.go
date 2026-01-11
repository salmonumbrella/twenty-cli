package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListTasks(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Task], error) {
	path := "/rest/tasks"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	var apiResp types.TasksListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Task]{
		Data:       apiResp.Data.Tasks,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

func (c *Client) GetTask(ctx context.Context, id string) (*types.Task, error) {
	path := fmt.Sprintf("/rest/tasks/%s", id)

	var apiResp types.TaskResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.Task, nil
}

type CreateTaskInput struct {
	Title      string `json:"title"`
	DueAt      string `json:"dueAt,omitempty"`
	Status     string `json:"status,omitempty"`
	AssigneeID string `json:"assigneeId,omitempty"`
}

func (c *Client) CreateTask(ctx context.Context, input *CreateTaskInput) (*types.Task, error) {
	path := "/rest/tasks"

	var apiResp types.CreateTaskResponse
	if err := c.Post(ctx, path, input, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.CreateTask, nil
}

type UpdateTaskInput struct {
	Title      *string `json:"title,omitempty"`
	DueAt      *string `json:"dueAt,omitempty"`
	Status     *string `json:"status,omitempty"`
	AssigneeID *string `json:"assigneeId,omitempty"`
}

func (c *Client) UpdateTask(ctx context.Context, id string, input *UpdateTaskInput) (*types.Task, error) {
	path := fmt.Sprintf("/rest/tasks/%s", id)

	var apiResp types.UpdateTaskResponse
	if err := c.Patch(ctx, path, input, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data.UpdateTask, nil
}

func (c *Client) DeleteTask(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/tasks/%s", id)

	var apiResp types.DeleteTaskResponse
	if err := c.do(ctx, "DELETE", path, nil, &apiResp); err != nil {
		return err
	}

	return nil
}
