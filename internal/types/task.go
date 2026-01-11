package types

import "time"

type Task struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	DueAt      string    `json:"dueAt,omitempty"`
	Status     string    `json:"status,omitempty"`
	AssigneeID string    `json:"assigneeId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TasksListResponse represents the API response for listing tasks
type TasksListResponse struct {
	Data struct {
		Tasks []Task `json:"tasks"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// TaskResponse represents the API response for a single task
type TaskResponse struct {
	Data struct {
		Task Task `json:"task"`
	} `json:"data"`
}

// CreateTaskResponse represents the API response for creating a task
type CreateTaskResponse struct {
	Data struct {
		CreateTask Task `json:"createTask"`
	} `json:"data"`
}

// UpdateTaskResponse represents the API response for updating a task
type UpdateTaskResponse struct {
	Data struct {
		UpdateTask Task `json:"updateTask"`
	} `json:"data"`
}

// DeleteTaskResponse represents the API response for deleting a task
type DeleteTaskResponse struct {
	Data struct {
		DeleteTask Task `json:"deleteTask"`
	} `json:"data"`
}
