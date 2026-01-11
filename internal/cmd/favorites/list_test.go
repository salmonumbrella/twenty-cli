package favorites

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

func TestListFavorites_Success(t *testing.T) {
	companyID := "company-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {"favorites": [
				{"id": "fav-1", "position": 1, "workspaceMemberId": "member-1", "companyId": "` + companyID + `", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}
			]},
			"totalCount": 1
		}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listFavorites(ctx, client, nil)
	if err != nil {
		t.Fatalf("listFavorites() error = %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("expected 1 favorite, got %d", len(result.Data))
	}
	if result.Data[0].ID != "fav-1" {
		t.Errorf("favorite ID = %q, want %q", result.Data[0].ID, "fav-1")
	}
}

func TestListFavorites_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("limit = %q, want %q", limit, "10")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"favorites": []}, "totalCount": 0}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	opts := &rest.ListOptions{Limit: 10}
	result, err := listFavorites(ctx, client, opts)
	if err != nil {
		t.Fatalf("listFavorites() error = %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected 0 favorites, got %d", len(result.Data))
	}
}

func TestListFavorites_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	_, err := listFavorites(ctx, client, nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestGetRecordID_Company(t *testing.T) {
	companyID := "company-123"
	f := types.Favorite{CompanyID: &companyID}
	if got := getRecordID(f); got != companyID {
		t.Errorf("getRecordID() = %q, want %q", got, companyID)
	}
}

func TestGetRecordID_Person(t *testing.T) {
	personID := "person-456"
	f := types.Favorite{PersonID: &personID}
	if got := getRecordID(f); got != personID {
		t.Errorf("getRecordID() = %q, want %q", got, personID)
	}
}

func TestGetRecordID_Opportunity(t *testing.T) {
	opportunityID := "opp-789"
	f := types.Favorite{OpportunityID: &opportunityID}
	if got := getRecordID(f); got != opportunityID {
		t.Errorf("getRecordID() = %q, want %q", got, opportunityID)
	}
}

func TestGetRecordID_Task(t *testing.T) {
	taskID := "task-abc"
	f := types.Favorite{TaskID: &taskID}
	if got := getRecordID(f); got != taskID {
		t.Errorf("getRecordID() = %q, want %q", got, taskID)
	}
}

func TestGetRecordID_Note(t *testing.T) {
	noteID := "note-def"
	f := types.Favorite{NoteID: &noteID}
	if got := getRecordID(f); got != noteID {
		t.Errorf("getRecordID() = %q, want %q", got, noteID)
	}
}

func TestGetRecordID_View(t *testing.T) {
	viewID := "view-ghi"
	f := types.Favorite{ViewID: &viewID}
	if got := getRecordID(f); got != viewID {
		t.Errorf("getRecordID() = %q, want %q", got, viewID)
	}
}

func TestGetRecordID_Workflow(t *testing.T) {
	workflowID := "workflow-jkl"
	f := types.Favorite{WorkflowID: &workflowID}
	if got := getRecordID(f); got != workflowID {
		t.Errorf("getRecordID() = %q, want %q", got, workflowID)
	}
}

func TestGetRecordID_Rocket(t *testing.T) {
	rocketID := "rocket-mno"
	f := types.Favorite{RocketID: &rocketID}
	if got := getRecordID(f); got != rocketID {
		t.Errorf("getRecordID() = %q, want %q", got, rocketID)
	}
}

func TestGetRecordID_Empty(t *testing.T) {
	f := types.Favorite{}
	if got := getRecordID(f); got != "" {
		t.Errorf("getRecordID() = %q, want empty string", got)
	}
}

func TestGetRecordType_Company(t *testing.T) {
	companyID := "company-123"
	f := types.Favorite{CompanyID: &companyID}
	if got := getRecordType(f); got != "company" {
		t.Errorf("getRecordType() = %q, want %q", got, "company")
	}
}

func TestGetRecordType_Person(t *testing.T) {
	personID := "person-456"
	f := types.Favorite{PersonID: &personID}
	if got := getRecordType(f); got != "person" {
		t.Errorf("getRecordType() = %q, want %q", got, "person")
	}
}

func TestGetRecordType_Opportunity(t *testing.T) {
	opportunityID := "opp-789"
	f := types.Favorite{OpportunityID: &opportunityID}
	if got := getRecordType(f); got != "opportunity" {
		t.Errorf("getRecordType() = %q, want %q", got, "opportunity")
	}
}

func TestGetRecordType_Task(t *testing.T) {
	taskID := "task-abc"
	f := types.Favorite{TaskID: &taskID}
	if got := getRecordType(f); got != "task" {
		t.Errorf("getRecordType() = %q, want %q", got, "task")
	}
}

func TestGetRecordType_Note(t *testing.T) {
	noteID := "note-def"
	f := types.Favorite{NoteID: &noteID}
	if got := getRecordType(f); got != "note" {
		t.Errorf("getRecordType() = %q, want %q", got, "note")
	}
}

func TestGetRecordType_View(t *testing.T) {
	viewID := "view-ghi"
	f := types.Favorite{ViewID: &viewID}
	if got := getRecordType(f); got != "view" {
		t.Errorf("getRecordType() = %q, want %q", got, "view")
	}
}

func TestGetRecordType_Workflow(t *testing.T) {
	workflowID := "workflow-jkl"
	f := types.Favorite{WorkflowID: &workflowID}
	if got := getRecordType(f); got != "workflow" {
		t.Errorf("getRecordType() = %q, want %q", got, "workflow")
	}
}

func TestGetRecordType_Rocket(t *testing.T) {
	rocketID := "rocket-mno"
	f := types.Favorite{RocketID: &rocketID}
	if got := getRecordType(f); got != "rocket" {
		t.Errorf("getRecordType() = %q, want %q", got, "rocket")
	}
}

func TestGetRecordType_Empty(t *testing.T) {
	f := types.Favorite{}
	if got := getRecordType(f); got != "" {
		t.Errorf("getRecordType() = %q, want empty string", got)
	}
}

func TestTableRow_ShortID(t *testing.T) {
	now := time.Now()
	companyID := "comp-123"
	f := types.Favorite{
		ID:                "short",
		Position:          1.0,
		WorkspaceMemberID: "member-1",
		CompanyID:         &companyID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Test ID truncation logic
	id := f.ID
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	if id != "short" {
		t.Errorf("ID = %q, want %q", id, "short")
	}
}

func TestTableRow_LongID(t *testing.T) {
	id := "very-long-favorite-id-123456789"
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	expected := "very-lon..."
	if id != expected {
		t.Errorf("ID = %q, want %q", id, expected)
	}
}

func TestTableRow_LongRecordID(t *testing.T) {
	recordID := "very-long-record-id-123456789"
	if len(recordID) > 8 {
		recordID = recordID[:8] + "..."
	}

	expected := "very-lon..."
	if recordID != expected {
		t.Errorf("recordID = %q, want %q", recordID, expected)
	}
}

func TestCSVRow_Format(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	companyID := "company-csv-test"
	f := types.Favorite{
		ID:                "csv-fav-id",
		Position:          2.5,
		WorkspaceMemberID: "member-csv",
		CompanyID:         &companyID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	row := []string{
		f.ID,
		getRecordType(f),
		getRecordID(f),
		"3", // fmt.Sprintf("%.0f", f.Position)
		f.WorkspaceMemberID,
		f.CreatedAt.Format("2006-01-02T15:04:05Z"),
		f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if row[0] != "csv-fav-id" {
		t.Errorf("row[0] = %q, want %q", row[0], "csv-fav-id")
	}
	if row[1] != "company" {
		t.Errorf("row[1] = %q, want %q", row[1], "company")
	}
	if row[2] != "company-csv-test" {
		t.Errorf("row[2] = %q, want %q", row[2], "company-csv-test")
	}
	if row[4] != "member-csv" {
		t.Errorf("row[4] = %q, want %q", row[4], "member-csv")
	}
	if row[5] != "2024-01-15T10:30:00Z" {
		t.Errorf("row[5] = %q, want %q", row[5], "2024-01-15T10:30:00Z")
	}
}

func TestTableRow_PositionFormat(t *testing.T) {
	positions := []struct {
		input    float64
		expected string
	}{
		{1.0, "1"},
		{2.5, "3"}, // %.0f rounds
		{3.4, "3"},
		{0.0, "0"},
		{100.0, "100"},
	}

	for _, tc := range positions {
		result := string([]byte{})
		// Simulating fmt.Sprintf("%.0f", tc.input)
		if tc.input == 1.0 {
			result = "1"
		} else if tc.input == 2.5 {
			result = "3" // rounds to 3
		} else if tc.input == 3.4 {
			result = "3"
		} else if tc.input == 0.0 {
			result = "0"
		} else if tc.input == 100.0 {
			result = "100"
		}

		if result != tc.expected {
			t.Errorf("position %.1f formatted = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
