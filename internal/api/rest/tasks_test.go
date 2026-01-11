package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestClient_ListTasks(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedTasks := []types.Task{
		{
			ID:         "task-1",
			Title:      "Review PR",
			DueAt:      "2024-12-31T23:59:59Z",
			Status:     "TODO",
			AssigneeID: "user-1",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "task-2",
			Title:      "Write documentation",
			DueAt:      "2025-01-15T12:00:00Z",
			Status:     "IN_PROGRESS",
			AssigneeID: "user-2",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/tasks" {
			t.Errorf("expected path /rest/tasks, got %s", r.URL.Path)
		}

		resp := types.TasksListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false, EndCursor: "cursor-2"},
		}
		resp.Data.Tasks = expectedTasks

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListTasks(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Data))
	}
	if result.Data[0].ID != "task-1" {
		t.Errorf("expected first task ID 'task-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Title != "Review PR" {
		t.Errorf("expected first task title 'Review PR', got %s", result.Data[0].Title)
	}
	if result.Data[1].Status != "IN_PROGRESS" {
		t.Errorf("expected second task status 'IN_PROGRESS', got %s", result.Data[1].Status)
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if result.PageInfo.EndCursor != "cursor-2" {
		t.Errorf("expected EndCursor 'cursor-2', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_ListTasks_WithOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		expectedParams map[string]string
	}{
		{
			name: "with limit",
			opts: &ListOptions{Limit: 10},
			expectedParams: map[string]string{
				"limit": "10",
			},
		},
		{
			name: "with cursor",
			opts: &ListOptions{Cursor: "abc123"},
			expectedParams: map[string]string{
				"starting_after": "abc123",
			},
		},
		{
			name: "with sort and order",
			opts: &ListOptions{Sort: "createdAt", Order: "desc"},
			expectedParams: map[string]string{
				"order_by":           "createdAt",
				"order_by_direction": "desc",
			},
		},
		{
			name: "combined options",
			opts: &ListOptions{
				Limit:  25,
				Cursor: "xyz789",
				Sort:   "title",
				Order:  "asc",
			},
			expectedParams: map[string]string{
				"limit":              "25",
				"starting_after":     "xyz789",
				"order_by":           "title",
				"order_by_direction": "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.RawQuery
				resp := types.TasksListResponse{TotalCount: 0}
				resp.Data.Tasks = []types.Task{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false, WithNoRetry())
			_, err := client.ListTasks(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for param, expected := range tt.expectedParams {
				if !strings.Contains(receivedQuery, param+"="+expected) {
					t.Errorf("expected query to contain %s=%s, got query: %s", param, expected, receivedQuery)
				}
			}
		})
	}
}

func TestClient_GetTask(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedTask := types.Task{
		ID:         "task-123",
		Title:      "Complete project",
		DueAt:      "2025-02-28T18:00:00Z",
		Status:     "TODO",
		AssigneeID: "user-456",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/tasks/task-123" {
			t.Errorf("expected path /rest/tasks/task-123, got %s", r.URL.Path)
		}

		resp := types.TaskResponse{}
		resp.Data.Task = expectedTask
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	task, err := client.GetTask(context.Background(), "task-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task-123" {
		t.Errorf("expected ID 'task-123', got %s", task.ID)
	}
	if task.Title != "Complete project" {
		t.Errorf("expected title 'Complete project', got %s", task.Title)
	}
	if task.Status != "TODO" {
		t.Errorf("expected status 'TODO', got %s", task.Status)
	}
	if task.AssigneeID != "user-456" {
		t.Errorf("expected assigneeId 'user-456', got %s", task.AssigneeID)
	}
}

func TestClient_GetTask_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Task not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.GetTask(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

func TestClient_CreateTask(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	input := &CreateTaskInput{
		Title:      "New Task",
		DueAt:      "2025-03-15T12:00:00Z",
		Status:     "TODO",
		AssigneeID: "user-789",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/tasks" {
			t.Errorf("expected path /rest/tasks, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput CreateTaskInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Title != "New Task" {
			t.Errorf("expected title 'New Task', got %s", receivedInput.Title)
		}
		if receivedInput.Status != "TODO" {
			t.Errorf("expected status 'TODO', got %s", receivedInput.Status)
		}
		if receivedInput.AssigneeID != "user-789" {
			t.Errorf("expected assigneeId 'user-789', got %s", receivedInput.AssigneeID)
		}

		// Return created task
		resp := types.CreateTaskResponse{}
		resp.Data.CreateTask = types.Task{
			ID:         "new-task-id",
			Title:      input.Title,
			DueAt:      input.DueAt,
			Status:     input.Status,
			AssigneeID: input.AssigneeID,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	task, err := client.CreateTask(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "new-task-id" {
		t.Errorf("expected ID 'new-task-id', got %s", task.ID)
	}
	if task.Title != "New Task" {
		t.Errorf("expected title 'New Task', got %s", task.Title)
	}
	if task.Status != "TODO" {
		t.Errorf("expected status 'TODO', got %s", task.Status)
	}
}

func TestClient_CreateTask_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"validation_error","message":"Title is required"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.CreateTask(context.Background(), &CreateTaskInput{})

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestClient_UpdateTask(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	updatedTitle := "Updated Task Title"
	updatedStatus := "IN_PROGRESS"
	input := &UpdateTaskInput{
		Title:  &updatedTitle,
		Status: &updatedStatus,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/tasks/task-123" {
			t.Errorf("expected path /rest/tasks/task-123, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput UpdateTaskInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Title == nil || *receivedInput.Title != "Updated Task Title" {
			t.Errorf("expected title 'Updated Task Title', got %v", receivedInput.Title)
		}
		if receivedInput.Status == nil || *receivedInput.Status != "IN_PROGRESS" {
			t.Errorf("expected status 'IN_PROGRESS', got %v", receivedInput.Status)
		}

		// Return updated task
		resp := types.UpdateTaskResponse{}
		resp.Data.UpdateTask = types.Task{
			ID:        "task-123",
			Title:     *input.Title,
			Status:    *input.Status,
			CreatedAt: now,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	task, err := client.UpdateTask(context.Background(), "task-123", input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task-123" {
		t.Errorf("expected ID 'task-123', got %s", task.ID)
	}
	if task.Title != "Updated Task Title" {
		t.Errorf("expected title 'Updated Task Title', got %s", task.Title)
	}
	if task.Status != "IN_PROGRESS" {
		t.Errorf("expected status 'IN_PROGRESS', got %s", task.Status)
	}
}

func TestClient_UpdateTask_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Task not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	title := "Updated"
	_, err := client.UpdateTask(context.Background(), "non-existent", &UpdateTaskInput{
		Title: &title,
	})

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

func TestClient_DeleteTask(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		var receivedMethod, receivedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			resp := types.DeleteTaskResponse{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteTask(context.Background(), "task-to-delete")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
		if receivedPath != "/rest/tasks/task-to-delete" {
			t.Errorf("expected path /rest/tasks/task-to-delete, got %s", receivedPath)
		}
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Task not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteTask(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent task, got nil")
		}
	})
}

func TestClient_ListTasks_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListTasks(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}
