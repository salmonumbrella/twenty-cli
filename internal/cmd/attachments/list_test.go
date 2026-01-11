package attachments

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestListAttachments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {"attachments": [
				{"id": "attach-1", "name": "document.pdf", "type": "application/pdf", "fullPath": "/files/document.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}
			]},
			"totalCount": 1
		}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listAttachments(ctx, client, nil)
	if err != nil {
		t.Fatalf("listAttachments() error = %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(result.Data))
	}
	if result.Data[0].ID != "attach-1" {
		t.Errorf("attachment ID = %q, want %q", result.Data[0].ID, "attach-1")
	}
	if result.Data[0].Name != "document.pdf" {
		t.Errorf("attachment Name = %q, want %q", result.Data[0].Name, "document.pdf")
	}
}

func TestListAttachments_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("limit = %q, want %q", limit, "10")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"attachments": []}, "totalCount": 0}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	opts := &rest.ListOptions{Limit: 10}
	result, err := listAttachments(ctx, client, opts)
	if err != nil {
		t.Fatalf("listAttachments() error = %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(result.Data))
	}
}

func TestListAttachments_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	_, err := listAttachments(ctx, client, nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestListAttachments_MultipleAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {"attachments": [
				{"id": "attach-1", "name": "doc1.pdf", "type": "application/pdf", "fullPath": "/files/doc1.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"},
				{"id": "attach-2", "name": "image.png", "type": "image/png", "fullPath": "/files/image.png", "createdAt": "2024-01-02T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z"},
				{"id": "attach-3", "name": "data.csv", "type": "text/csv", "fullPath": "/files/data.csv", "createdAt": "2024-01-03T00:00:00Z", "updatedAt": "2024-01-03T00:00:00Z"}
			]},
			"totalCount": 3
		}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listAttachments(ctx, client, nil)
	if err != nil {
		t.Fatalf("listAttachments() error = %v", err)
	}

	if len(result.Data) != 3 {
		t.Errorf("expected 3 attachments, got %d", len(result.Data))
	}
	if result.TotalCount != 3 {
		t.Errorf("totalCount = %d, want %d", result.TotalCount, 3)
	}
}

func TestTruncateID_Short(t *testing.T) {
	result := truncateID("short")
	if result != "short" {
		t.Errorf("truncateID() = %q, want %q", result, "short")
	}
}

func TestTruncateID_Long(t *testing.T) {
	result := truncateID("very-long-attachment-id-123456789")
	expected := "very-lon..."
	if result != expected {
		t.Errorf("truncateID() = %q, want %q", result, expected)
	}
}

func TestTruncateID_ExactlyEight(t *testing.T) {
	result := truncateID("12345678")
	if result != "12345678" {
		t.Errorf("truncateID() = %q, want %q", result, "12345678")
	}
}

func TestTruncateID_NineChars(t *testing.T) {
	result := truncateID("123456789")
	expected := "12345678..."
	if result != expected {
		t.Errorf("truncateID() = %q, want %q", result, expected)
	}
}

func TestTruncateName_Short(t *testing.T) {
	result := truncateName("test.pdf")
	if result != "test.pdf" {
		t.Errorf("truncateName() = %q, want %q", result, "test.pdf")
	}
}

func TestTruncateName_Long(t *testing.T) {
	result := truncateName("this-is-a-very-long-filename-that-exceeds-thirty-characters.pdf")
	expected := "this-is-a-very-long-filename-t..."
	if result != expected {
		t.Errorf("truncateName() = %q, want %q", result, expected)
	}
}

func TestTruncateName_ExactlyThirty(t *testing.T) {
	result := truncateName("123456789012345678901234567890")
	if result != "123456789012345678901234567890" {
		t.Errorf("truncateName() = %q, want %q", result, "123456789012345678901234567890")
	}
}

func TestTruncateFullPath_Short(t *testing.T) {
	result := truncateFullPath("/files/test.pdf")
	if result != "/files/test.pdf" {
		t.Errorf("truncateFullPath() = %q, want %q", result, "/files/test.pdf")
	}
}

func TestTruncateFullPath_Long(t *testing.T) {
	result := truncateFullPath("/very/long/path/to/deeply/nested/folder/structure/with/many/levels/file.pdf")
	expected := "/very/long/path/to/deeply/nested/folder/..."
	if result != expected {
		t.Errorf("truncateFullPath() = %q, want %q", result, expected)
	}
}

func TestTruncateFullPath_ExactlyForty(t *testing.T) {
	result := truncateFullPath("1234567890123456789012345678901234567890")
	if result != "1234567890123456789012345678901234567890" {
		t.Errorf("truncateFullPath() = %q, want %q", result, "1234567890123456789012345678901234567890")
	}
}

func TestGetOptionalString_Nil(t *testing.T) {
	result := getOptionalString(nil)
	if result != "" {
		t.Errorf("getOptionalString(nil) = %q, want empty string", result)
	}
}

func TestGetOptionalString_NonNil(t *testing.T) {
	value := "test-value"
	result := getOptionalString(&value)
	if result != "test-value" {
		t.Errorf("getOptionalString() = %q, want %q", result, "test-value")
	}
}

func TestFormatTableRow(t *testing.T) {
	now := time.Now()
	a := types.Attachment{
		ID:        "attach-123",
		Name:      "document.pdf",
		Type:      "application/pdf",
		FullPath:  "/files/document.pdf",
		CreatedAt: now,
		UpdatedAt: now,
	}

	row := formatTableRow(a)
	if len(row) != 4 {
		t.Fatalf("formatTableRow() returned %d columns, want 4", len(row))
	}
	if row[0] != "attach-1..." {
		t.Errorf("row[0] = %q, want %q", row[0], "attach-1...")
	}
	if row[1] != "document.pdf" {
		t.Errorf("row[1] = %q, want %q", row[1], "document.pdf")
	}
	if row[2] != "application/pdf" {
		t.Errorf("row[2] = %q, want %q", row[2], "application/pdf")
	}
	if row[3] != "/files/document.pdf" {
		t.Errorf("row[3] = %q, want %q", row[3], "/files/document.pdf")
	}
}

func TestFormatTableRow_LongValues(t *testing.T) {
	now := time.Now()
	a := types.Attachment{
		ID:        "very-long-attachment-id-123456789",
		Name:      "this-is-a-very-long-filename-that-exceeds-thirty-characters.pdf",
		Type:      "application/pdf",
		FullPath:  "/very/long/path/to/deeply/nested/folder/structure/with/many/levels/file.pdf",
		CreatedAt: now,
		UpdatedAt: now,
	}

	row := formatTableRow(a)
	if row[0] != "very-lon..." {
		t.Errorf("row[0] = %q, want %q", row[0], "very-lon...")
	}
	if row[1] != "this-is-a-very-long-filename-t..." {
		t.Errorf("row[1] = %q, want %q", row[1], "this-is-a-very-long-filename-t...")
	}
	if row[3] != "/very/long/path/to/deeply/nested/folder/..." {
		t.Errorf("row[3] = %q, want %q", row[3], "/very/long/path/to/deeply/nested/folder/...")
	}
}

func TestFormatCSVRow(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	companyID := "company-123"
	personID := "person-456"

	a := types.Attachment{
		ID:        "csv-attach",
		Name:      "report.pdf",
		Type:      "application/pdf",
		FullPath:  "/files/report.pdf",
		CompanyID: &companyID,
		PersonID:  &personID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	row := formatCSVRow(a)
	if len(row) != 12 {
		t.Fatalf("formatCSVRow() returned %d columns, want 12", len(row))
	}
	if row[0] != "csv-attach" {
		t.Errorf("row[0] = %q, want %q", row[0], "csv-attach")
	}
	if row[1] != "report.pdf" {
		t.Errorf("row[1] = %q, want %q", row[1], "report.pdf")
	}
	if row[4] != "company-123" {
		t.Errorf("row[4] = %q, want %q", row[4], "company-123")
	}
	if row[5] != "person-456" {
		t.Errorf("row[5] = %q, want %q", row[5], "person-456")
	}
	if row[10] != "2024-01-15T10:30:00Z" {
		t.Errorf("row[10] = %q, want %q", row[10], "2024-01-15T10:30:00Z")
	}
}

func TestFormatCSVRow_NilOptionals(t *testing.T) {
	now := time.Now()
	a := types.Attachment{
		ID:        "attach-nil",
		Name:      "test.pdf",
		Type:      "application/pdf",
		FullPath:  "/files/test.pdf",
		CreatedAt: now,
		UpdatedAt: now,
	}

	row := formatCSVRow(a)
	// Check all optional fields are empty strings
	if row[4] != "" {
		t.Errorf("row[4] (companyId) = %q, want empty string", row[4])
	}
	if row[5] != "" {
		t.Errorf("row[5] (personId) = %q, want empty string", row[5])
	}
	if row[6] != "" {
		t.Errorf("row[6] (activityId) = %q, want empty string", row[6])
	}
	if row[7] != "" {
		t.Errorf("row[7] (taskId) = %q, want empty string", row[7])
	}
	if row[8] != "" {
		t.Errorf("row[8] (noteId) = %q, want empty string", row[8])
	}
	if row[9] != "" {
		t.Errorf("row[9] (authorId) = %q, want empty string", row[9])
	}
}

func TestFormatCSVRow_AllOptionals(t *testing.T) {
	now := time.Now()
	companyID := "company-123"
	personID := "person-456"
	activityID := "activity-789"
	taskID := "task-abc"
	noteID := "note-def"
	authorID := "author-ghi"

	a := types.Attachment{
		ID:         "attach-all",
		Name:       "full.pdf",
		Type:       "application/pdf",
		FullPath:   "/files/full.pdf",
		CompanyID:  &companyID,
		PersonID:   &personID,
		ActivityID: &activityID,
		TaskID:     &taskID,
		NoteID:     &noteID,
		AuthorID:   &authorID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	row := formatCSVRow(a)
	if row[4] != "company-123" {
		t.Errorf("row[4] = %q, want %q", row[4], "company-123")
	}
	if row[5] != "person-456" {
		t.Errorf("row[5] = %q, want %q", row[5], "person-456")
	}
	if row[6] != "activity-789" {
		t.Errorf("row[6] = %q, want %q", row[6], "activity-789")
	}
	if row[7] != "task-abc" {
		t.Errorf("row[7] = %q, want %q", row[7], "task-abc")
	}
	if row[8] != "note-def" {
		t.Errorf("row[8] = %q, want %q", row[8], "note-def")
	}
	if row[9] != "author-ghi" {
		t.Errorf("row[9] = %q, want %q", row[9], "author-ghi")
	}
}

func TestListAttachments_WithCursor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("starting_after")
		if cursor != "prev-cursor" {
			t.Errorf("cursor = %q, want %q", cursor, "prev-cursor")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"attachments": []}, "totalCount": 0}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	opts := &rest.ListOptions{Cursor: "prev-cursor"}
	_, err := listAttachments(ctx, client, opts)
	if err != nil {
		t.Fatalf("listAttachments() error = %v", err)
	}
}

func TestListAttachments_WithSort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderBy := r.URL.Query().Get("order_by")
		if orderBy == "" {
			t.Error("order_by query parameter is empty")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"attachments": []}, "totalCount": 0}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	opts := &rest.ListOptions{Sort: "createdAt", Order: "desc"}
	_, err := listAttachments(ctx, client, opts)
	if err != nil {
		t.Fatalf("listAttachments() error = %v", err)
	}
}

func TestListAttachments_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	_, err := listAttachments(ctx, client, nil)
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}
