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

func TestClient_ListWebhooks(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedWebhooks := []types.Webhook{
		{
			ID:          "webhook-1",
			TargetURL:   "https://example.com/webhook1",
			Operation:   "*.created",
			Description: "Notify on all creations",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "webhook-2",
			TargetURL:   "https://example.com/webhook2",
			Operation:   "person.updated",
			Description: "Notify on person updates",
			IsActive:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/webhooks" {
			t.Errorf("expected path /rest/webhooks, got %s", r.URL.Path)
		}

		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedWebhooks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListWebhooks(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 webhooks, got %d", len(result.Data))
	}
	if result.Data[0].ID != "webhook-1" {
		t.Errorf("expected first webhook ID 'webhook-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].TargetURL != "https://example.com/webhook1" {
		t.Errorf("expected targetUrl 'https://example.com/webhook1', got %s", result.Data[0].TargetURL)
	}
	if result.Data[0].Operation != "*.created" {
		t.Errorf("expected operation '*.created', got %s", result.Data[0].Operation)
	}
	if !result.Data[0].IsActive {
		t.Error("expected first webhook to be active")
	}
	if result.Data[1].IsActive {
		t.Error("expected second webhook to be inactive")
	}
	// Webhooks API returns plain array, no pagination info
	if result.PageInfo != nil {
		t.Error("expected PageInfo to be nil for webhooks")
	}
}

func TestClient_ListWebhooks_WithOptions(t *testing.T) {
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
				// Twenty webhooks API returns a plain array
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]types.Webhook{})
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false, WithNoRetry())
			_, err := client.ListWebhooks(context.Background(), tt.opts)
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

func TestClient_ListWebhooks_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListWebhooks(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}

func TestClient_GetWebhook(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedWebhook := types.Webhook{
		ID:          "webhook-456",
		TargetURL:   "https://example.com/my-webhook",
		Operation:   "company.deleted",
		Description: "Handle company deletions",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	t.Run("basic get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/webhooks/webhook-456" {
				t.Errorf("expected path /rest/webhooks/webhook-456, got %s", r.URL.Path)
			}

			// GetWebhook returns Webhook directly, not wrapped in response struct
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedWebhook)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		webhook, err := client.GetWebhook(context.Background(), "webhook-456")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if webhook.ID != "webhook-456" {
			t.Errorf("expected ID 'webhook-456', got %s", webhook.ID)
		}
		if webhook.TargetURL != "https://example.com/my-webhook" {
			t.Errorf("expected targetUrl 'https://example.com/my-webhook', got %s", webhook.TargetURL)
		}
		if webhook.Operation != "company.deleted" {
			t.Errorf("expected operation 'company.deleted', got %s", webhook.Operation)
		}
		if !webhook.IsActive {
			t.Error("expected webhook to be active")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Webhook not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		_, err := client.GetWebhook(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent webhook, got nil")
		}
	})
}

func TestClient_CreateWebhook(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	input := &CreateWebhookInput{
		TargetURL:   "https://example.com/new-webhook",
		Operation:   "*.created",
		Description: "Handle all creations",
		Secret:      "my-secret-key",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/webhooks" {
			t.Errorf("expected path /rest/webhooks, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput CreateWebhookInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.TargetURL != "https://example.com/new-webhook" {
			t.Errorf("expected targetUrl 'https://example.com/new-webhook', got %s", receivedInput.TargetURL)
		}
		if receivedInput.Operation != "*.created" {
			t.Errorf("expected operation '*.created', got %s", receivedInput.Operation)
		}
		if receivedInput.Secret != "my-secret-key" {
			t.Errorf("expected secret 'my-secret-key', got %s", receivedInput.Secret)
		}

		// Return created webhook
		createdWebhook := types.Webhook{
			ID:          "new-webhook-id",
			TargetURL:   input.TargetURL,
			Operation:   input.Operation,
			Description: input.Description,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdWebhook)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	webhook, err := client.CreateWebhook(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if webhook.ID != "new-webhook-id" {
		t.Errorf("expected ID 'new-webhook-id', got %s", webhook.ID)
	}
	if webhook.TargetURL != "https://example.com/new-webhook" {
		t.Errorf("expected targetUrl 'https://example.com/new-webhook', got %s", webhook.TargetURL)
	}
	if webhook.Operation != "*.created" {
		t.Errorf("expected operation '*.created', got %s", webhook.Operation)
	}
	if !webhook.IsActive {
		t.Error("expected webhook to be active")
	}
}

func TestClient_CreateWebhook_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"validation_error","message":"Invalid target URL"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.CreateWebhook(context.Background(), &CreateWebhookInput{
		TargetURL: "invalid-url",
		Operation: "*.created",
	})

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestClient_DeleteWebhook(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		var receivedMethod, receivedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteWebhook(context.Background(), "webhook-to-delete")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
		if receivedPath != "/rest/webhooks/webhook-to-delete" {
			t.Errorf("expected path /rest/webhooks/webhook-to-delete, got %s", receivedPath)
		}
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Webhook not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteWebhook(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent webhook, got nil")
		}
	})
}
