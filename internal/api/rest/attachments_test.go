package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestClient_ListAttachments(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	companyID := "company-1"
	personID := "person-1"
	expectedAttachments := []types.Attachment{
		{
			ID:        "attachment-1",
			Name:      "document.pdf",
			FullPath:  "/files/document.pdf",
			Type:      "application/pdf",
			CompanyID: &companyID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "attachment-2",
			Name:      "image.png",
			FullPath:  "/files/image.png",
			Type:      "image/png",
			PersonID:  &personID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/attachments" {
			t.Errorf("expected path /rest/attachments, got %s", r.URL.Path)
		}

		resp := types.AttachmentsListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false, EndCursor: "cursor-2"},
		}
		resp.Data.Attachments = expectedAttachments

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListAttachments(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(result.Data))
	}
	if result.Data[0].ID != "attachment-1" {
		t.Errorf("expected first attachment ID 'attachment-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Name != "document.pdf" {
		t.Errorf("expected name 'document.pdf', got %s", result.Data[0].Name)
	}
	if result.Data[0].Type != "application/pdf" {
		t.Errorf("expected type 'application/pdf', got %s", result.Data[0].Type)
	}
	if result.Data[0].CompanyID == nil || *result.Data[0].CompanyID != "company-1" {
		t.Errorf("expected companyId 'company-1', got %v", result.Data[0].CompanyID)
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if result.PageInfo.EndCursor != "cursor-2" {
		t.Errorf("expected EndCursor 'cursor-2', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_ListAttachments_WithOptions(t *testing.T) {
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
			name: "combined options",
			opts: &ListOptions{
				Limit:  25,
				Cursor: "xyz789",
				Sort:   "createdAt",
				Order:  "desc",
			},
			expectedParams: map[string]string{
				"limit":              "25",
				"starting_after":     "xyz789",
				"order_by":           "createdAt",
				"order_by_direction": "desc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.RawQuery
				resp := types.AttachmentsListResponse{TotalCount: 0}
				resp.Data.Attachments = []types.Attachment{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false, WithNoRetry())
			_, err := client.ListAttachments(context.Background(), tt.opts)
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

func TestClient_ListAttachments_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListAttachments(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}

func TestClient_GetAttachment(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	taskID := "task-123"
	expectedAttachment := types.Attachment{
		ID:        "attachment-456",
		Name:      "report.xlsx",
		FullPath:  "/files/reports/report.xlsx",
		Type:      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		TaskID:    &taskID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("basic get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/attachments/attachment-456" {
				t.Errorf("expected path /rest/attachments/attachment-456, got %s", r.URL.Path)
			}

			resp := types.AttachmentResponse{}
			resp.Data.Attachment = expectedAttachment
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		attachment, err := client.GetAttachment(context.Background(), "attachment-456")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attachment.ID != "attachment-456" {
			t.Errorf("expected ID 'attachment-456', got %s", attachment.ID)
		}
		if attachment.Name != "report.xlsx" {
			t.Errorf("expected name 'report.xlsx', got %s", attachment.Name)
		}
		if attachment.FullPath != "/files/reports/report.xlsx" {
			t.Errorf("expected fullPath '/files/reports/report.xlsx', got %s", attachment.FullPath)
		}
		if attachment.TaskID == nil || *attachment.TaskID != "task-123" {
			t.Errorf("expected taskId 'task-123', got %v", attachment.TaskID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Attachment not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		_, err := client.GetAttachment(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent attachment, got nil")
		}
	})
}
