package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTask_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "task-123",
		"title": "Follow up with client",
		"dueAt": "2024-07-15T09:00:00Z",
		"status": "TODO",
		"assigneeId": "user-456",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var task Task
	err := json.Unmarshal([]byte(jsonData), &task)
	if err != nil {
		t.Fatalf("failed to unmarshal Task: %v", err)
	}

	if task.ID != "task-123" {
		t.Errorf("expected ID='task-123', got %q", task.ID)
	}
	if task.Title != "Follow up with client" {
		t.Errorf("expected Title='Follow up with client', got %q", task.Title)
	}
	if task.DueAt != "2024-07-15T09:00:00Z" {
		t.Errorf("expected DueAt='2024-07-15T09:00:00Z', got %q", task.DueAt)
	}
	if task.Status != "TODO" {
		t.Errorf("expected Status='TODO', got %q", task.Status)
	}
	if task.AssigneeID != "user-456" {
		t.Errorf("expected AssigneeID='user-456', got %q", task.AssigneeID)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestTask_JSONMarshal(t *testing.T) {
	task := Task{
		ID:         "task-789",
		Title:      "Review contract",
		DueAt:      "2024-08-01T12:00:00Z",
		Status:     "IN_PROGRESS",
		AssigneeID: "user-123",
		CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("failed to marshal Task: %v", err)
	}

	// Round-trip: unmarshal back
	var parsed Task
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized Task: %v", err)
	}

	if parsed.ID != task.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", task.ID, parsed.ID)
	}
	if parsed.Title != task.Title {
		t.Errorf("round-trip failed: expected Title=%q, got %q", task.Title, parsed.Title)
	}
	if parsed.Status != task.Status {
		t.Errorf("round-trip failed: expected Status=%q, got %q", task.Status, parsed.Status)
	}
	if parsed.AssigneeID != task.AssigneeID {
		t.Errorf("round-trip failed: expected AssigneeID=%q, got %q", task.AssigneeID, parsed.AssigneeID)
	}
}

func TestTask_MinimalFields(t *testing.T) {
	jsonData := `{
		"id": "task-minimal",
		"title": "Simple task",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var task Task
	err := json.Unmarshal([]byte(jsonData), &task)
	if err != nil {
		t.Fatalf("failed to unmarshal Task: %v", err)
	}

	if task.ID != "task-minimal" {
		t.Errorf("expected ID='task-minimal', got %q", task.ID)
	}
	if task.Title != "Simple task" {
		t.Errorf("expected Title='Simple task', got %q", task.Title)
	}
	if task.DueAt != "" {
		t.Errorf("expected DueAt to be empty, got %q", task.DueAt)
	}
	if task.Status != "" {
		t.Errorf("expected Status to be empty, got %q", task.Status)
	}
	if task.AssigneeID != "" {
		t.Errorf("expected AssigneeID to be empty, got %q", task.AssigneeID)
	}
}

func TestTasksListResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"tasks": [
				{
					"id": "t1",
					"title": "Task One",
					"status": "TODO",
					"createdAt": "2024-01-15T10:30:00Z",
					"updatedAt": "2024-06-20T14:45:00Z"
				},
				{
					"id": "t2",
					"title": "Task Two",
					"status": "DONE",
					"dueAt": "2024-07-01T00:00:00Z",
					"createdAt": "2024-02-10T08:00:00Z",
					"updatedAt": "2024-05-15T12:00:00Z"
				}
			]
		},
		"totalCount": 2,
		"pageInfo": {
			"hasNextPage": true,
			"endCursor": "next-cursor"
		}
	}`

	var resp TasksListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal TasksListResponse: %v", err)
	}

	if len(resp.Data.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(resp.Data.Tasks))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected TotalCount=2, got %d", resp.TotalCount)
	}
	if resp.Data.Tasks[0].ID != "t1" {
		t.Errorf("expected first task ID='t1', got %q", resp.Data.Tasks[0].ID)
	}
	if resp.Data.Tasks[0].Title != "Task One" {
		t.Errorf("expected first task Title='Task One', got %q", resp.Data.Tasks[0].Title)
	}
	if resp.Data.Tasks[1].Status != "DONE" {
		t.Errorf("expected second task Status='DONE', got %q", resp.Data.Tasks[1].Status)
	}
	if resp.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if !resp.PageInfo.HasNextPage {
		t.Error("expected HasNextPage=true")
	}
	if resp.PageInfo.EndCursor != "next-cursor" {
		t.Errorf("expected EndCursor='next-cursor', got %q", resp.PageInfo.EndCursor)
	}
}

func TestTaskResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"task": {
				"id": "task-single",
				"title": "Single Task",
				"status": "IN_PROGRESS",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp TaskResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal TaskResponse: %v", err)
	}

	if resp.Data.Task.ID != "task-single" {
		t.Errorf("expected ID='task-single', got %q", resp.Data.Task.ID)
	}
	if resp.Data.Task.Title != "Single Task" {
		t.Errorf("expected Title='Single Task', got %q", resp.Data.Task.Title)
	}
}

func TestCreateTaskResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"createTask": {
				"id": "new-task",
				"title": "New Task",
				"status": "TODO",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp CreateTaskResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateTaskResponse: %v", err)
	}

	if resp.Data.CreateTask.ID != "new-task" {
		t.Errorf("expected ID='new-task', got %q", resp.Data.CreateTask.ID)
	}
	if resp.Data.CreateTask.Title != "New Task" {
		t.Errorf("expected Title='New Task', got %q", resp.Data.CreateTask.Title)
	}
}

func TestUpdateTaskResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"updateTask": {
				"id": "updated-task",
				"title": "Updated Task",
				"status": "DONE",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp UpdateTaskResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal UpdateTaskResponse: %v", err)
	}

	if resp.Data.UpdateTask.ID != "updated-task" {
		t.Errorf("expected ID='updated-task', got %q", resp.Data.UpdateTask.ID)
	}
	if resp.Data.UpdateTask.Status != "DONE" {
		t.Errorf("expected Status='DONE', got %q", resp.Data.UpdateTask.Status)
	}
}

func TestDeleteTaskResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"deleteTask": {
				"id": "deleted-task",
				"title": "Deleted Task",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp DeleteTaskResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal DeleteTaskResponse: %v", err)
	}

	if resp.Data.DeleteTask.ID != "deleted-task" {
		t.Errorf("expected ID='deleted-task', got %q", resp.Data.DeleteTask.ID)
	}
}

func TestTask_StatusValues(t *testing.T) {
	statuses := []string{"TODO", "IN_PROGRESS", "DONE", "CANCELED"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			jsonData := `{
				"id": "task-status-test",
				"title": "Test Task",
				"status": "` + status + `",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}`

			var task Task
			err := json.Unmarshal([]byte(jsonData), &task)
			if err != nil {
				t.Fatalf("failed to unmarshal Task: %v", err)
			}

			if task.Status != status {
				t.Errorf("expected Status=%q, got %q", status, task.Status)
			}
		})
	}
}
