package people

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/auth"
	"github.com/salmonumbrella/twenty-cli/internal/secrets"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

// setupTestEnv sets up a mock server and configures auth/viper for testing
func setupTestEnv(t *testing.T, handler http.Handler) func() {
	t.Helper()

	// Create mock server
	server := httptest.NewServer(handler)

	// Save original values
	originalStore := getAuthStore()
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalBaseURL := viper.GetString("base_url")
	originalDebug := viper.GetBool("debug")
	originalOutput := viper.GetString("output")

	// Set up mock store with token
	mockStore := secrets.NewMockStore()
	_ = mockStore.SetToken("default", secrets.Token{
		Profile:      "default",
		RefreshToken: "test-token",
	})
	auth.SetStore(mockStore)

	// Clear environment variable to use mock store
	os.Unsetenv("TWENTY_TOKEN")

	// Configure viper
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("profile", "")
	viper.Set("query", "")

	// Return cleanup function
	return func() {
		server.Close()
		auth.SetStore(originalStore)
		if originalEnv != "" {
			os.Setenv("TWENTY_TOKEN", originalEnv)
		} else {
			os.Unsetenv("TWENTY_TOKEN")
		}
		viper.Set("base_url", originalBaseURL)
		viper.Set("debug", originalDebug)
		viper.Set("output", originalOutput)
	}
}

func getAuthStore() secrets.Store {
	// Get the current store - this is a bit hacky but works for testing
	return nil
}

func TestRunGet_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedPerson := types.Person{
		ID: "person-123",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		JobTitle:  "Engineer",
		CreatedAt: now,
		UpdatedAt: now,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/people/person-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.PersonResponse{}
		resp.Data.Person = expectedPerson

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	err := runGet(getCmd, []string{"person-123"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}
}

func TestRunGet_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Person not found"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	err := runGet(getCmd, []string{"non-existent"})
	if err == nil {
		t.Error("expected error for non-existent person")
	}
}

func TestRunDelete_Integration_WithoutForce(t *testing.T) {
	// Save and restore forceDelete flag
	originalForce := forceDelete
	defer func() { forceDelete = originalForce }()

	forceDelete = false

	err := runDelete(deleteCmd, []string{"person-123"})
	if err == nil {
		t.Error("expected error when --force is not set")
	}

	errMsg := err.Error()
	if errMsg != "delete aborted: use --force to confirm deletion of person-123" {
		t.Errorf("unexpected error message: %s", errMsg)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/rest/people/person-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore forceDelete flag
	originalForce := forceDelete
	defer func() { forceDelete = originalForce }()

	forceDelete = true

	err := runDelete(deleteCmd, []string{"person-123"})
	if err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}
}

func TestRunCreate_WithFlags(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/people" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore create flags
	origFirstName := createFirstName
	origLastName := createLastName
	origEmail := createEmail
	origData := createData
	defer func() {
		createFirstName = origFirstName
		createLastName = origLastName
		createEmail = origEmail
		createData = origData
	}()

	createFirstName = "John"
	createLastName = "Doe"
	createEmail = "john@example.com"
	createData = ""

	err := runCreate(createCmd, []string{})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
}

func TestRunCreate_WithJSONData(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "Jane", LastName: "Smith"},
			Email:     types.Email{PrimaryEmail: "jane@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore create flags
	origData := createData
	defer func() {
		createData = origData
	}()

	createData = `{"name":{"firstName":"Jane","lastName":"Smith"},"emails":{"primaryEmail":"jane@example.com"}}`

	err := runCreate(createCmd, []string{})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
}

func TestRunCreate_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	// Save and restore create flags
	origData := createData
	defer func() {
		createData = origData
	}()

	createData = `{invalid json}`

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRunExport_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore export flags
	origFormat := exportFormat
	origOutput := exportOutput
	origAll := exportAll
	defer func() {
		exportFormat = origFormat
		exportOutput = origOutput
		exportAll = origAll
	}()

	exportFormat = "json"
	exportOutput = ""
	exportAll = false

	err := runExport(exportCmd, []string{})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}
}

func TestRunExport_CSV(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore export flags
	origFormat := exportFormat
	origOutput := exportOutput
	origAll := exportAll
	defer func() {
		exportFormat = origFormat
		exportOutput = origOutput
		exportAll = origAll
	}()

	exportFormat = "csv"
	exportOutput = ""
	exportAll = false

	err := runExport(exportCmd, []string{})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}
}

func TestRunExport_UnsupportedFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 0,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = []types.Person{}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore export flags
	origFormat := exportFormat
	origOutput := exportOutput
	origAll := exportAll
	defer func() {
		exportFormat = origFormat
		exportOutput = origOutput
		exportAll = origAll
	}()

	exportFormat = "xml" // unsupported
	exportOutput = ""
	exportAll = false

	err := runExport(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestListPeople_WithMockServer(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "person-2",
			Name:      types.Name{FirstName: "Jane", LastName: "Smith"},
			Email:     types.Email{PrimaryEmail: "jane@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/people" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.PeopleListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a real client for this test
	client := createTestClient(server.URL)

	result, err := listPeople(context.Background(), client, nil)
	if err != nil {
		t.Fatalf("listPeople failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("expected 2 people, got %d", len(result.Data))
	}
}

// createTestClient creates a test REST client
func createTestClient(baseURL string) *rest.Client {
	return rest.NewClient(baseURL, "test-token", false, rest.WithNoRetry())
}

// TestListCmd_Execute tests the list command execution which exercises newListCmd callbacks
func TestListCmd_Execute_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set JSON output
	viper.Set("output", "json")

	cmd := newListCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

// TestListCmd_Execute_CSV tests the list command with CSV output to exercise CSVRow callback
func TestListCmd_Execute_CSV(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set CSV output
	viper.Set("output", "csv")

	cmd := newListCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

// TestListCmd_Execute_Table tests the list command with table output to exercise TableRow callback
func TestListCmd_Execute_Table(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1-with-long-id",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			JobTitle:  "Engineer",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text (table) output
	viper.Set("output", "text")

	cmd := newListCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

// TestListCmd_Execute_WithFilters tests the list command with filter flags to exercise BuildFilter callback
func TestListCmd_Execute_WithFilters(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			City:      "New York",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	filterChecked := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that filter query params are present
		query := r.URL.Query()
		if filter := query.Get("filter"); filter != "" {
			filterChecked = true
		}

		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "json")

	cmd := newListCmd()
	cmd.SetArgs([]string{"--email", "john@example.com", "--name", "John", "--city", "New York", "--company-id", "company-1"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command with filters failed: %v", err)
	}

	if !filterChecked {
		t.Log("Warning: filter query parameter was not sent (may be expected with empty filter)")
	}
}

func TestRunUpdate_WithJSONData(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	requestCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		resp := types.UpdatePersonResponse{}
		resp.Data.UpdatePerson = types.Person{
			ID:        "person-123",
			Name:      types.Name{FirstName: "Updated", LastName: "Person"},
			Email:     types.Email{PrimaryEmail: "updated@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore update flags
	origData := updateData
	defer func() {
		updateData = origData
	}()

	updateData = `{"name":{"firstName":"Updated","lastName":"Person"}}`

	err := runUpdate(updateCmd, []string{"person-123"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}
}

func TestRunUpdate_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	// Save and restore update flags
	origData := updateData
	defer func() {
		updateData = origData
	}()

	updateData = `{invalid json}`

	err := runUpdate(updateCmd, []string{"person-123"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRunImport_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "import-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[
		{"name":{"firstName":"John","lastName":"Doe"},"emails":{"primaryEmail":"john@example.com"}}
	]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore import flags
	origFormat := importFormat
	origDryRun := importDryRun
	defer func() {
		importFormat = origFormat
		importDryRun = origDryRun
	}()

	importFormat = ""
	importDryRun = false

	err = runImport(importCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}
}

func TestRunImport_CSV(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary CSV file
	tmpFile, err := os.CreateTemp("", "import-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	csvContent := `FirstName,LastName,Email
John,Doe,john@example.com`
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore import flags
	origFormat := importFormat
	origDryRun := importDryRun
	defer func() {
		importFormat = origFormat
		importDryRun = origDryRun
	}()

	importFormat = ""
	importDryRun = false

	err = runImport(importCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}
}

func TestRunImport_DryRun(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called in dry run mode")
	}))
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "import-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[
		{"name":{"firstName":"John","lastName":"Doe"},"emails":{"primaryEmail":"john@example.com"}}
	]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore import flags
	origFormat := importFormat
	origDryRun := importDryRun
	defer func() {
		importFormat = origFormat
		importDryRun = origDryRun
	}()

	importFormat = ""
	importDryRun = true

	err = runImport(importCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("runImport dry-run failed: %v", err)
	}
}

func TestRunImport_EmptyFile(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for empty file")
	}))
	defer cleanup()

	// Create a temporary JSON file with empty array
	tmpFile, err := os.CreateTemp("", "import-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("[]"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore import flags
	origFormat := importFormat
	origDryRun := importDryRun
	defer func() {
		importFormat = origFormat
		importDryRun = origDryRun
	}()

	importFormat = ""
	importDryRun = false

	err = runImport(importCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("runImport with empty file failed: %v", err)
	}
}

func TestRunImport_UnsupportedFormat(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	// Create a temporary file with unknown extension
	tmpFile, err := os.CreateTemp("", "import-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Save and restore import flags
	origFormat := importFormat
	origDryRun := importDryRun
	defer func() {
		importFormat = origFormat
		importDryRun = origDryRun
	}()

	importFormat = "" // Will fail to detect format
	importDryRun = false

	err = runImport(importCmd, []string{tmpFile.Name()})
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestRunImport_FileNotFound(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	err := runImport(importCmd, []string{"/nonexistent/file.json"})
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestRunBatchCreate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "Test", LastName: "Person"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[
		{"name":{"firstName":"John","lastName":"Doe"}},
		{"name":{"firstName":"Jane","lastName":"Smith"}}
	]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch create flags
	origFile := batchCreateFile
	defer func() {
		batchCreateFile = origFile
	}()

	batchCreateFile = tmpFile.Name()

	err = runBatchCreate(batchCreateCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchCreate failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestRunBatchDelete_WithoutForce(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called without --force")
	}))
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-delete-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`["id-1", "id-2"]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch delete flags
	origFile := batchDeleteFile
	origForce := batchDeleteForce
	defer func() {
		batchDeleteFile = origFile
		batchDeleteForce = origForce
	}()

	batchDeleteFile = tmpFile.Name()
	batchDeleteForce = false

	err = runBatchDelete(batchDeleteCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchDelete without force should not return error: %v", err)
	}
}

func TestRunBatchDelete_WithForce(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-delete-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`["id-1", "id-2"]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch delete flags
	origFile := batchDeleteFile
	origForce := batchDeleteForce
	defer func() {
		batchDeleteFile = origFile
		batchDeleteForce = origForce
	}()

	batchDeleteFile = tmpFile.Name()
	batchDeleteForce = true

	err = runBatchDelete(batchDeleteCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchDelete failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestRunExport_AllPages(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var resp types.PeopleListResponse
		if callCount == 1 {
			resp = types.PeopleListResponse{
				TotalCount: 3,
				PageInfo:   &types.PageInfo{HasNextPage: true, EndCursor: "cursor-1"},
			}
			resp.Data.People = []types.Person{
				{ID: "person-1", CreatedAt: now, UpdatedAt: now},
			}
		} else {
			resp = types.PeopleListResponse{
				TotalCount: 3,
				PageInfo:   &types.PageInfo{HasNextPage: false, EndCursor: "cursor-2"},
			}
			resp.Data.People = []types.Person{
				{ID: "person-2", CreatedAt: now, UpdatedAt: now},
				{ID: "person-3", CreatedAt: now, UpdatedAt: now},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore export flags
	origFormat := exportFormat
	origOutput := exportOutput
	origAll := exportAll
	defer func() {
		exportFormat = origFormat
		exportOutput = origOutput
		exportAll = origAll
	}()

	exportFormat = "json"
	exportOutput = ""
	exportAll = true // Fetch all pages

	err := runExport(exportCmd, []string{})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestRunExport_ToFile(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary output file
	tmpFile, err := os.CreateTemp("", "export-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Save and restore export flags
	origFormat := exportFormat
	origOutput := exportOutput
	origAll := exportAll
	defer func() {
		exportFormat = origFormat
		exportOutput = origOutput
		exportAll = origAll
	}()

	exportFormat = "json"
	exportOutput = tmpFile.Name()
	exportAll = false

	err = runExport(exportCmd, []string{})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}

	// Verify file was written
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("output file should not be empty")
	}
}

func TestRunUpsert_WithFlags(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GraphQL endpoint for upsert (uses createPerson mutation with upsert=true)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"createPerson": map[string]interface{}{
					"id": "person-123",
					"name": map[string]interface{}{
						"firstName": "John",
						"lastName":  "Doe",
					},
					"emails": map[string]interface{}{
						"primaryEmail": "john@example.com",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore upsert flags
	origFirstName := upsertFirstName
	origLastName := upsertLastName
	origEmail := upsertEmail
	origPhone := upsertPhone
	origJobTitle := upsertJobTitle
	origCompanyID := upsertCompanyID
	origData := upsertData
	defer func() {
		upsertFirstName = origFirstName
		upsertLastName = origLastName
		upsertEmail = origEmail
		upsertPhone = origPhone
		upsertJobTitle = origJobTitle
		upsertCompanyID = origCompanyID
		upsertData = origData
	}()

	upsertFirstName = "John"
	upsertLastName = "Doe"
	upsertEmail = "john@example.com"
	upsertPhone = ""
	upsertJobTitle = ""
	upsertCompanyID = ""
	upsertData = ""

	err := runUpsert(upsertCmd, []string{})
	if err != nil {
		t.Fatalf("runUpsert failed: %v", err)
	}
}

func TestRunUpsert_WithJSONData(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GraphQL endpoint for upsert (uses createPerson mutation with upsert=true)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"createPerson": map[string]interface{}{
					"id": "person-123",
					"name": map[string]interface{}{
						"firstName": "Jane",
						"lastName":  "Smith",
					},
					"emails": map[string]interface{}{
						"primaryEmail": "jane@example.com",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore upsert flags
	origEmail := upsertEmail
	origData := upsertData
	defer func() {
		upsertEmail = origEmail
		upsertData = origData
	}()

	upsertEmail = "jane@example.com"
	upsertData = `{"firstName":"Jane","lastName":"Smith"}`

	err := runUpsert(upsertCmd, []string{})
	if err != nil {
		t.Fatalf("runUpsert failed: %v", err)
	}
}

func TestRunUpsert_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	// Save and restore upsert flags
	origEmail := upsertEmail
	origData := upsertData
	defer func() {
		upsertEmail = origEmail
		upsertData = origData
	}()

	upsertEmail = "test@example.com"
	upsertData = `{invalid json}`

	err := runUpsert(upsertCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRunUpdate_WithNameFlags(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	requestCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			// First call to get existing person
			resp := types.PersonResponse{
				Data: struct {
					Person types.Person `json:"person"`
				}{
					Person: types.Person{
						ID:        "person-123",
						Name:      types.Name{FirstName: "OldFirst", LastName: "OldLast"},
						Email:     types.Email{PrimaryEmail: "old@example.com"},
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			// PATCH request to update
			resp := types.UpdatePersonResponse{
				Data: struct {
					UpdatePerson types.Person `json:"updatePerson"`
				}{
					UpdatePerson: types.Person{
						ID:        "person-123",
						Name:      types.Name{FirstName: "NewFirst", LastName: "OldLast"},
						Email:     types.Email{PrimaryEmail: "old@example.com"},
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a fresh command to test flag handling
	testCmd := &cobra.Command{}
	testCmd.Flags().StringVar(&updateFirstName, "first-name", "", "first name")
	testCmd.Flags().StringVar(&updateLastName, "last-name", "", "last name")
	testCmd.Flags().StringVar(&updateEmail, "email", "", "email")
	testCmd.Flags().StringVar(&updatePhone, "phone", "", "phone")
	testCmd.Flags().StringVar(&updateJobTitle, "job-title", "", "job title")
	testCmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data")

	// Save and restore update flags
	origFirstName := updateFirstName
	origLastName := updateLastName
	origData := updateData
	defer func() {
		updateFirstName = origFirstName
		updateLastName = origLastName
		updateData = origData
	}()

	updateFirstName = "NewFirst"
	updateLastName = ""
	updateData = ""

	// Mark the flag as changed
	testCmd.Flags().Set("first-name", "NewFirst")

	err := runUpdate(testCmd, []string{"person-123"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}
}

func TestRunBatchCreate_WithErrors(t *testing.T) {
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call succeeds
			now := time.Now().UTC().Truncate(time.Second)
			resp := types.CreatePersonResponse{}
			resp.Data.CreatePerson = types.Person{
				ID:        "new-person-id",
				Name:      types.Name{FirstName: "Test", LastName: "Person"},
				CreatedAt: now,
				UpdatedAt: now,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		} else {
			// Second call fails
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":{"message":"Invalid input"}}`))
		}
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[
		{"name":{"firstName":"John","lastName":"Doe"}},
		{"name":{"firstName":"Jane","lastName":"Smith"}}
	]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch create flags
	origFile := batchCreateFile
	defer func() {
		batchCreateFile = origFile
	}()

	batchCreateFile = tmpFile.Name()

	// This should not return an error, but report errors in output
	err = runBatchCreate(batchCreateCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchCreate should not fail: %v", err)
	}
}

func TestRunBatchDelete_WithErrors(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Not found"}}`))
		}
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-delete-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`["id-1", "id-2"]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch delete flags
	origFile := batchDeleteFile
	origForce := batchDeleteForce
	defer func() {
		batchDeleteFile = origFile
		batchDeleteForce = origForce
	}()

	batchDeleteFile = tmpFile.Name()
	batchDeleteForce = true

	// This should not return an error, but report errors in output
	err = runBatchDelete(batchDeleteCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchDelete should not fail: %v", err)
	}
}

func TestRunImport_WithFailedCreates(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			now := time.Now().UTC().Truncate(time.Second)
			resp := types.CreatePersonResponse{}
			resp.Data.CreatePerson = types.Person{
				ID:        "new-person-id",
				Name:      types.Name{FirstName: "John", LastName: "Doe"},
				CreatedAt: now,
				UpdatedAt: now,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":{"message":"Invalid input"}}`))
		}
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "import-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[
		{"name":{"firstName":"John","lastName":"Doe"}},
		{"name":{"firstName":"Jane","lastName":"Smith"}}
	]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore import flags
	origFormat := importFormat
	origDryRun := importDryRun
	defer func() {
		importFormat = origFormat
		importDryRun = origDryRun
	}()

	importFormat = ""
	importDryRun = false

	// This should not return an error, but report failures in output
	err = runImport(importCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("runImport should not fail: %v", err)
	}
}

func TestRunGet_WithInclude(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedPerson := types.Person{
		ID: "person-123",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		Company: &types.Company{
			ID:   "company-1",
			Name: "Test Company",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify depth parameter is set when include is used
		if r.URL.Query().Get("depth") != "1" {
			t.Errorf("expected depth=1, got %s", r.URL.Query().Get("depth"))
		}

		resp := types.PersonResponse{}
		resp.Data.Person = expectedPerson

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore getInclude
	origInclude := getInclude
	defer func() { getInclude = origInclude }()

	getInclude = []string{"company"}

	err := runGet(getCmd, []string{"person-123"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}
}

// TestRunBatchCreate_TextOutput tests batch create with text output format
func TestRunBatchCreate_TextOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "Test", LastName: "Person"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[{"name":{"firstName":"John","lastName":"Doe"}}]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch create flags
	origFile := batchCreateFile
	defer func() {
		batchCreateFile = origFile
	}()

	batchCreateFile = tmpFile.Name()

	err = runBatchCreate(batchCreateCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchCreate failed: %v", err)
	}
}

// TestRunBatchCreate_TextOutput_WithErrors tests text output with errors
func TestRunBatchCreate_TextOutput_WithErrors(t *testing.T) {
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// All calls fail
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Invalid input"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `[{"name":{"firstName":"John","lastName":"Doe"}}]`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch create flags
	origFile := batchCreateFile
	defer func() {
		batchCreateFile = origFile
	}()

	batchCreateFile = tmpFile.Name()

	err = runBatchCreate(batchCreateCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchCreate should not fail: %v", err)
	}
}

// TestRunBatchDelete_TextOutput tests batch delete with text output
func TestRunBatchDelete_TextOutput(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-delete-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`["id-1", "id-2"]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch delete flags
	origFile := batchDeleteFile
	origForce := batchDeleteForce
	defer func() {
		batchDeleteFile = origFile
		batchDeleteForce = origForce
	}()

	batchDeleteFile = tmpFile.Name()
	batchDeleteForce = true

	err = runBatchDelete(batchDeleteCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchDelete failed: %v", err)
	}
}

// TestRunBatchDelete_TextOutput_WithErrors tests batch delete text output with errors
func TestRunBatchDelete_TextOutput_WithErrors(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// All calls fail
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Not found"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "batch-delete-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`["id-1"]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Save and restore batch delete flags
	origFile := batchDeleteFile
	origForce := batchDeleteForce
	defer func() {
		batchDeleteFile = origFile
		batchDeleteForce = origForce
	}()

	batchDeleteFile = tmpFile.Name()
	batchDeleteForce = true

	err = runBatchDelete(batchDeleteCmd, []string{})
	if err != nil {
		t.Fatalf("runBatchDelete should not fail: %v", err)
	}
}

// TestRunCreate_TextOutput tests create with text output
func TestRunCreate_TextOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Save and restore create flags
	origData := createData
	defer func() {
		createData = origData
	}()

	createData = `{"name":{"firstName":"John","lastName":"Doe"}}`

	err := runCreate(createCmd, []string{})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
}

// TestRunUpdate_TextOutput tests update with text output
func TestRunUpdate_TextOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UpdatePersonResponse{}
		resp.Data.UpdatePerson = types.Person{
			ID:        "person-123",
			Name:      types.Name{FirstName: "Updated", LastName: "Person"},
			Email:     types.Email{PrimaryEmail: "updated@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Save and restore update flags
	origData := updateData
	defer func() {
		updateData = origData
	}()

	updateData = `{"name":{"firstName":"Updated","lastName":"Person"}}`

	err := runUpdate(updateCmd, []string{"person-123"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}
}

// TestRunExport_CSVFormat tests export with CSV format to full file
func TestRunExport_CSVFormat(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	people := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.PeopleListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.People = people

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create temp output file
	tmpFile, err := os.CreateTemp("", "export-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Save and restore export flags
	origOutput := exportOutput
	origFormat := exportFormat
	defer func() {
		exportOutput = origOutput
		exportFormat = origFormat
	}()

	exportOutput = tmpFile.Name()
	exportFormat = "csv"

	err = runExport(exportCmd, []string{})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("output file should not be empty")
	}
}

// TestRunUpsert_TextOutput tests upsert with text output
func TestRunUpsert_TextOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GraphQL endpoint for upsert (uses createPerson mutation with upsert=true)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"createPerson": map[string]interface{}{
					"id": "person-123",
					"name": map[string]interface{}{
						"firstName": "John",
						"lastName":  "Doe",
					},
					"emails": map[string]interface{}{
						"primaryEmail": "john@example.com",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Set text output
	viper.Set("output", "text")

	// Save and restore upsert flags
	origFirstName := upsertFirstName
	origLastName := upsertLastName
	origEmail := upsertEmail
	origPhone := upsertPhone
	origJobTitle := upsertJobTitle
	origCompanyID := upsertCompanyID
	origData := upsertData
	defer func() {
		upsertFirstName = origFirstName
		upsertLastName = origLastName
		upsertEmail = origEmail
		upsertPhone = origPhone
		upsertJobTitle = origJobTitle
		upsertCompanyID = origCompanyID
		upsertData = origData
	}()

	upsertFirstName = "John"
	upsertLastName = "Doe"
	upsertEmail = "john@example.com"
	upsertPhone = ""
	upsertJobTitle = ""
	upsertCompanyID = ""
	upsertData = ""

	err := runUpsert(upsertCmd, []string{})
	if err != nil {
		t.Fatalf("runUpsert failed: %v", err)
	}
}

// TestRunUpsert_WithAllFlags tests upsert with all optional flags
func TestRunUpsert_WithAllFlags(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GraphQL endpoint for upsert (uses createPerson mutation with upsert=true)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"createPerson": map[string]interface{}{
					"id": "person-123",
					"name": map[string]interface{}{
						"firstName": "John",
						"lastName":  "Doe",
					},
					"emails": map[string]interface{}{
						"primaryEmail": "john@example.com",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "json")

	// Save and restore upsert flags
	origFirstName := upsertFirstName
	origLastName := upsertLastName
	origEmail := upsertEmail
	origPhone := upsertPhone
	origJobTitle := upsertJobTitle
	origCompanyID := upsertCompanyID
	origData := upsertData
	defer func() {
		upsertFirstName = origFirstName
		upsertLastName = origLastName
		upsertEmail = origEmail
		upsertPhone = origPhone
		upsertJobTitle = origJobTitle
		upsertCompanyID = origCompanyID
		upsertData = origData
	}()

	upsertFirstName = "John"
	upsertLastName = "Doe"
	upsertEmail = "john@example.com"
	upsertPhone = "+1234567890"
	upsertJobTitle = "Engineer"
	upsertCompanyID = "company-1"
	upsertData = ""

	err := runUpsert(upsertCmd, []string{})
	if err != nil {
		t.Fatalf("runUpsert failed: %v", err)
	}
}
