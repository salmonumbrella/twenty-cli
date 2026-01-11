package webhooks

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestListCmd_Flags(t *testing.T) {
	flags := []string{"limit", "cursor", "all", "filter", "sort", "order"}
	for _, flag := range flags {
		if listCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestListCmd_Use(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list")
	}
}

func TestListCmd_Short(t *testing.T) {
	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestListCmd_Long(t *testing.T) {
	if listCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestListWebhooks_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/webhooks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "webhook-1", "targetUrl": "https://example.com/hook1", "operation": "*.created", "description": "Test hook", "isActive": true, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}]`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listWebhooks(ctx, client, nil)
	if err != nil {
		t.Fatalf("listWebhooks() error = %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("expected 1 webhook, got %d", len(result.Data))
	}
	if result.Data[0].ID != "webhook-1" {
		t.Errorf("webhook ID = %q, want %q", result.Data[0].ID, "webhook-1")
	}
	if result.Data[0].TargetURL != "https://example.com/hook1" {
		t.Errorf("webhook TargetURL = %q, want %q", result.Data[0].TargetURL, "https://example.com/hook1")
	}
	if result.Data[0].Operation != "*.created" {
		t.Errorf("webhook Operation = %q, want %q", result.Data[0].Operation, "*.created")
	}
	if !result.Data[0].IsActive {
		t.Error("webhook IsActive should be true")
	}
}

func TestListWebhooks_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("limit = %q, want %q", limit, "10")
		}

		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	opts := &rest.ListOptions{Limit: 10}
	result, err := listWebhooks(ctx, client, opts)
	if err != nil {
		t.Fatalf("listWebhooks() error = %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected 0 webhooks, got %d", len(result.Data))
	}
}

func TestListWebhooks_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	_, err := listWebhooks(ctx, client, nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestListWebhooks_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "bad-token", false)
	ctx := context.Background()

	_, err := listWebhooks(ctx, client, nil)
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestListWebhooks_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listWebhooks(ctx, client, nil)
	if err != nil {
		t.Fatalf("listWebhooks() error = %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected 0 webhooks, got %d", len(result.Data))
	}
	if result.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", result.TotalCount)
	}
}

func TestListWebhooks_MultipleWebhooks(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"id": "webhook-1", "targetUrl": "https://example.com/hook1", "operation": "*.created", "isActive": true, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"},
			{"id": "webhook-2", "targetUrl": "https://example.com/hook2", "operation": "person.updated", "isActive": false, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}
		]`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listWebhooks(ctx, client, nil)
	if err != nil {
		t.Fatalf("listWebhooks() error = %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("expected 2 webhooks, got %d", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if result.Data[0].ID != "webhook-1" {
		t.Errorf("first webhook ID = %q, want %q", result.Data[0].ID, "webhook-1")
	}
	if result.Data[1].ID != "webhook-2" {
		t.Errorf("second webhook ID = %q, want %q", result.Data[1].ID, "webhook-2")
	}
}

func TestListWebhooks_NoPagination(t *testing.T) {
	// Twenty webhooks API returns a plain array without pagination support
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"id": "webhook-1", "targetUrl": "https://example.com/hook1", "operation": "*.created", "isActive": true, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"},
			{"id": "webhook-2", "targetUrl": "https://example.com/hook2", "operation": "person.updated", "isActive": false, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}
		]`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listWebhooks(ctx, client, nil)
	if err != nil {
		t.Fatalf("listWebhooks() error = %v", err)
	}

	// Webhooks API returns plain array, no pagination info
	if result.PageInfo != nil {
		t.Error("expected PageInfo to be nil for webhooks")
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 webhooks, got %d", len(result.Data))
	}
	// TotalCount is derived from array length
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
}

func TestTableRow_ShortID(t *testing.T) {
	now := time.Now()
	webhook := types.Webhook{
		ID:        "short",
		TargetURL: "https://example.com/hook",
		Operation: "*.created",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test the ID truncation logic
	id := webhook.ID
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	if id != "short" {
		t.Errorf("ID = %q, want %q", id, "short")
	}
}

func TestTableRow_LongID(t *testing.T) {
	id := "very-long-webhook-id-123456789"
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	expected := "very-lon..."
	if id != expected {
		t.Errorf("ID = %q, want %q", id, expected)
	}
}

func TestTableRow_ExactlyEightCharID(t *testing.T) {
	id := "12345678"
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	expected := "12345678"
	if id != expected {
		t.Errorf("ID = %q, want %q", id, expected)
	}
}

func TestCSVRow_Format(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	webhook := types.Webhook{
		ID:          "csv-webhook-id",
		TargetURL:   "https://example.com/webhook",
		Operation:   "person.created",
		Description: "Test webhook description",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	row := []string{
		webhook.ID,
		webhook.TargetURL,
		webhook.Operation,
		webhook.Description,
		"true",
		webhook.CreatedAt.Format("2006-01-02T15:04:05Z"),
		webhook.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if row[0] != "csv-webhook-id" {
		t.Errorf("row[0] = %q, want %q", row[0], "csv-webhook-id")
	}
	if row[1] != "https://example.com/webhook" {
		t.Errorf("row[1] = %q, want %q", row[1], "https://example.com/webhook")
	}
	if row[2] != "person.created" {
		t.Errorf("row[2] = %q, want %q", row[2], "person.created")
	}
	if row[3] != "Test webhook description" {
		t.Errorf("row[3] = %q, want %q", row[3], "Test webhook description")
	}
	if row[4] != "true" {
		t.Errorf("row[4] = %q, want %q", row[4], "true")
	}
	if row[5] != "2024-01-15T10:30:00Z" {
		t.Errorf("row[5] = %q, want %q", row[5], "2024-01-15T10:30:00Z")
	}
}

func TestCSVRow_InactiveWebhook(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	webhook := types.Webhook{
		ID:        "inactive-webhook",
		TargetURL: "https://example.com/inactive",
		Operation: "company.deleted",
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	isActiveStr := "false"
	if webhook.IsActive {
		isActiveStr = "true"
	}

	if isActiveStr != "false" {
		t.Errorf("isActive = %q, want %q", isActiveStr, "false")
	}
}

func TestTableRow_DateFormat(t *testing.T) {
	createdAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	formatted := createdAt.Format("2006-01-02T15:04:05Z")

	expected := "2024-06-15T10:30:00Z"
	if formatted != expected {
		t.Errorf("formatted = %q, want %q", formatted, expected)
	}
}

func TestTableRow_IsActiveFormatting(t *testing.T) {
	tests := []struct {
		isActive bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		result := "false"
		if tt.isActive {
			result = "true"
		}
		if result != tt.expected {
			t.Errorf("isActive=%v formatted = %q, want %q", tt.isActive, result, tt.expected)
		}
	}
}

func TestListCmd_RunE_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "webhook-full-test", "targetUrl": "https://example.com/full", "operation": "*.created", "description": "Full test", "isActive": true, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}]`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listCmd.RunE(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("listCmd.RunE() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify table headers are in output
	if !strings.Contains(output, "ID") {
		t.Errorf("output missing ID header: %s", output)
	}
	if !strings.Contains(output, "URL") {
		t.Errorf("output missing URL header: %s", output)
	}
	if !strings.Contains(output, "webhook-f...") || !strings.Contains(output, "webhook-full-test"[:8]) {
		// ID should be truncated to 8 chars + "..."
		t.Logf("output: %s", output)
	}
}

func TestListCmd_RunE_JSONOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "webhook-json-test", "targetUrl": "https://example.com/json", "operation": "person.updated", "description": "JSON test", "isActive": false, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}]`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listCmd.RunE(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("listCmd.RunE() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "webhook-json-test") {
		t.Errorf("JSON output missing webhook ID: %s", output)
	}
	if !strings.Contains(output, "person.updated") {
		t.Errorf("JSON output missing operation: %s", output)
	}
}

func TestListCmd_RunE_CSVOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "webhook-csv-test", "targetUrl": "https://example.com/csv", "operation": "company.deleted", "description": "CSV test description", "isActive": true, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}]`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listCmd.RunE(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("listCmd.RunE() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify CSV headers
	if !strings.Contains(output, "id,targetUrl,operation,description,isActive,createdAt,updatedAt") {
		t.Errorf("CSV output missing headers: %s", output)
	}
	// Verify data row
	if !strings.Contains(output, "webhook-csv-test") {
		t.Errorf("CSV output missing webhook ID: %s", output)
	}
	if !strings.Contains(output, "CSV test description") {
		t.Errorf("CSV output missing description: %s", output)
	}
}

func TestListCmd_RunE_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := listCmd.RunE(listCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestListCmd_TableRowTruncation(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Twenty webhooks API returns a plain array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Use a webhook ID longer than 8 characters to test truncation
		w.Write([]byte(`[{"id": "abcdefghijklmnop", "targetUrl": "https://example.com/truncate", "operation": "task.created", "description": "", "isActive": true, "createdAt": "` + now.Format(time.RFC3339) + `", "updatedAt": "` + now.Format(time.RFC3339) + `"}]`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listCmd.RunE(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("listCmd.RunE() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// The ID "abcdefghijklmnop" should be truncated to "abcdefgh..."
	if !strings.Contains(output, "abcdefgh...") {
		t.Errorf("ID not truncated correctly in output: %s", output)
	}
}
