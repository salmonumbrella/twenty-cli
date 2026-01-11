package opportunities

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

// createTestClient creates a test REST client
func createTestClient(baseURL string) *rest.Client {
	return rest.NewClient(baseURL, "test-token", false, rest.WithNoRetry())
}

// TestCmd tests the main opportunities command
func TestCmd(t *testing.T) {
	if Cmd == nil {
		t.Error("Cmd should not be nil")
	}

	if Cmd.Use != "opportunities" {
		t.Errorf("Cmd.Use = %q, want %q", Cmd.Use, "opportunities")
	}

	if Cmd.Short == "" {
		t.Error("Cmd.Short should not be empty")
	}

	// Verify subcommands are registered
	subcommands := []string{"list", "get", "create", "update", "delete"}
	for _, name := range subcommands {
		found := false
		for _, cmd := range Cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

// TestCreateCmd tests the create command structure
func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"name", "amount", "currency", "close-date", "stage", "probability", "company-id", "contact-id", "data"}
	for _, flag := range flags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestCreateCmd_Use(t *testing.T) {
	if createCmd.Use != "create" {
		t.Errorf("Use = %q, want %q", createCmd.Use, "create")
	}
}

func TestCreateCmd_Short(t *testing.T) {
	if createCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCreateCmd_DataFlagShorthand(t *testing.T) {
	flag := createCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

// TestDeleteCmd tests the delete command structure
func TestDeleteCmd_Flags(t *testing.T) {
	if deleteCmd.Flags().Lookup("force") == nil {
		t.Error("force flag not registered")
	}
}

func TestDeleteCmd_ForceFlagShorthand(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("force flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestDeleteCmd_Use(t *testing.T) {
	if deleteCmd.Use != "delete <id>" {
		t.Errorf("Use = %q, want %q", deleteCmd.Use, "delete <id>")
	}
}

func TestDeleteCmd_Short(t *testing.T) {
	if deleteCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestDeleteCmd_Args(t *testing.T) {
	if deleteCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := deleteCmd.Args(deleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = deleteCmd.Args(deleteCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = deleteCmd.Args(deleteCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

// TestGetCmd tests the get command structure
func TestGetCmd_Use(t *testing.T) {
	if getCmd.Use != "get <id>" {
		t.Errorf("Use = %q, want %q", getCmd.Use, "get <id>")
	}
}

func TestGetCmd_Short(t *testing.T) {
	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestGetCmd_Args(t *testing.T) {
	if getCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = getCmd.Args(getCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = getCmd.Args(getCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

// TestUpdateCmd tests the update command structure
func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"name", "amount", "currency", "close-date", "stage", "probability", "company-id", "contact-id", "data"}
	for _, flag := range flags {
		if updateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update <id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	if updateCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = updateCmd.Args(updateCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestUpdateCmd_DataFlagShorthand(t *testing.T) {
	flag := updateCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

// TestFormatCurrency tests the formatCurrency function
func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency *types.Currency
		want     string
	}{
		{
			name:     "nil currency",
			currency: nil,
			want:     "-",
		},
		{
			name: "valid currency",
			currency: &types.Currency{
				AmountMicros: "50000000000",
				CurrencyCode: "USD",
			},
			want: "50000.00 USD",
		},
		{
			name: "zero amount",
			currency: &types.Currency{
				AmountMicros: "0",
				CurrencyCode: "EUR",
			},
			want: "0.00 EUR",
		},
		{
			name: "invalid amount string",
			currency: &types.Currency{
				AmountMicros: "invalid",
				CurrencyCode: "USD",
			},
			want: "invalid USD",
		},
		{
			name: "small amount",
			currency: &types.Currency{
				AmountMicros: "1000000",
				CurrencyCode: "GBP",
			},
			want: "1.00 GBP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCurrency(tt.currency)
			if got != tt.want {
				t.Errorf("formatCurrency() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestListOpportunities tests the listOpportunities function
func TestListOpportunities_WithMockServer(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	opportunities := []types.Opportunity{
		{
			ID:        "opp-1",
			Name:      "Deal 1",
			Stage:     "Negotiation",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "opp-2",
			Name:      "Deal 2",
			Stage:     "Closed",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/opportunities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.OpportunitiesListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.Opportunities = opportunities

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := createTestClient(server.URL)

	result, err := listOpportunities(context.Background(), client, nil)
	if err != nil {
		t.Fatalf("listOpportunities failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("expected 2 opportunities, got %d", len(result.Data))
	}
}

// TestRunGet tests the runGet function
func TestRunGet_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedOpp := types.Opportunity{
		ID:    "opp-123",
		Name:  "Big Deal",
		Stage: "Proposal",
		Amount: &types.Currency{
			AmountMicros: "100000000000",
			CurrencyCode: "USD",
		},
		Probability: 75,
		CloseDate:   "2024-12-31",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/opportunities/opp-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.OpportunityResponse{}
		resp.Data.Opportunity = expectedOpp

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	err := runGet(getCmd, []string{"opp-123"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}
}

func TestRunGet_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Opportunity not found"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	err := runGet(getCmd, []string{"non-existent"})
	if err == nil {
		t.Error("expected error for non-existent opportunity")
	}
}

// TestOutputOpportunity tests the outputOpportunity function
func TestOutputOpportunity_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	o := &types.Opportunity{
		ID:          "opp-1",
		Name:        "Test Deal",
		Stage:       "Proposal",
		Probability: 50,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputOpportunity(o, "json", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputOpportunity failed: %v", err)
	}

	buf.ReadFrom(r)

	// Verify output is valid JSON
	var parsed types.Opportunity
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.ID != "opp-1" {
		t.Errorf("ID = %q, want %q", parsed.ID, "opp-1")
	}
}

func TestOutputOpportunity_CSV(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	o := &types.Opportunity{
		ID:          "opp-1",
		Name:        "Test Deal",
		Stage:       "Proposal",
		Probability: 50,
		Amount: &types.Currency{
			AmountMicros: "50000000000",
			CurrencyCode: "USD",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputOpportunity(o, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputOpportunity failed: %v", err)
	}

	buf.ReadFrom(r)

	// Verify output is valid CSV
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("expected 2 rows, got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"id", "name", "stage", "amount", "probability", "createdAt", "updatedAt"}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check data
	if records[1][0] != "opp-1" {
		t.Errorf("ID = %q, want %q", records[1][0], "opp-1")
	}
}

func TestOutputOpportunity_Text(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	o := &types.Opportunity{
		ID:               "opp-1",
		Name:             "Test Deal",
		Stage:            "Proposal",
		Probability:      50,
		CloseDate:        "2024-12-31",
		CompanyID:        "company-123",
		PointOfContactID: "contact-456",
		Amount: &types.Currency{
			AmountMicros: "50000000000",
			CurrencyCode: "USD",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputOpportunity(o, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputOpportunity failed: %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	// Check that output contains expected fields
	if !strings.Contains(output, "opp-1") {
		t.Errorf("output should contain ID 'opp-1'")
	}
	if !strings.Contains(output, "Test Deal") {
		t.Errorf("output should contain name 'Test Deal'")
	}
	if !strings.Contains(output, "Proposal") {
		t.Errorf("output should contain stage 'Proposal'")
	}
	if !strings.Contains(output, "50%") {
		t.Errorf("output should contain probability '50%%'")
	}
	if !strings.Contains(output, "2024-12-31") {
		t.Errorf("output should contain close date")
	}
}

func TestOutputOpportunity_Default(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	o := &types.Opportunity{
		ID:        "opp-1",
		Name:      "Test Deal",
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Empty format should default to text
	err := outputOpportunity(o, "", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputOpportunity failed: %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "opp-1") {
		t.Errorf("output should contain ID")
	}
}

// TestRunCreate tests the runCreate function
func TestRunCreate_WithFlags(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/opportunities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.CreateOpportunityResponse{}
		resp.Data.CreateOpportunity = types.Opportunity{
			ID:        "new-opp-id",
			Name:      "New Deal",
			Stage:     "Qualification",
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
	origName := createName
	origAmount := createAmount
	origCurrency := createCurrency
	origCloseDate := createCloseDate
	origStage := createStage
	origProbability := createProbability
	origCompanyID := createCompanyID
	origContactID := createPointOfContactID
	origData := createData
	defer func() {
		createName = origName
		createAmount = origAmount
		createCurrency = origCurrency
		createCloseDate = origCloseDate
		createStage = origStage
		createProbability = origProbability
		createCompanyID = origCompanyID
		createPointOfContactID = origContactID
		createData = origData
	}()

	createName = "New Deal"
	createAmount = 50000.0
	createCurrency = "USD"
	createCloseDate = "2024-12-31"
	createStage = "Qualification"
	createProbability = 25
	createCompanyID = "company-1"
	createPointOfContactID = "contact-1"
	createData = ""

	err := runCreate(createCmd, []string{})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
}

func TestRunCreate_WithJSONData(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreateOpportunityResponse{}
		resp.Data.CreateOpportunity = types.Opportunity{
			ID:        "new-opp-id",
			Name:      "JSON Deal",
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

	createData = `{"name":"JSON Deal","stage":"Proposal","probability":50}`

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

func TestRunCreate_TextOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreateOpportunityResponse{}
		resp.Data.CreateOpportunity = types.Opportunity{
			ID:        "new-opp-id",
			Name:      "Text Deal",
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")

	// Save and restore create flags
	origData := createData
	defer func() {
		createData = origData
	}()

	createData = `{"name":"Text Deal"}`

	err := runCreate(createCmd, []string{})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
}

func TestRunCreate_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Invalid input"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore create flags
	origData := createData
	origName := createName
	defer func() {
		createData = origData
		createName = origName
	}()

	createData = ""
	createName = "Test"

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Error("expected error for API error")
	}
}

// TestRunUpdate tests the runUpdate function
func TestRunUpdate_WithJSONData(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		resp := types.UpdateOpportunityResponse{}
		resp.Data.UpdateOpportunity = types.Opportunity{
			ID:        "opp-123",
			Name:      "Updated Deal",
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

	updateData = `{"name":"Updated Deal"}`

	err := runUpdate(updateCmd, []string{"opp-123"})
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

	err := runUpdate(updateCmd, []string{"opp-123"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRunUpdate_WithFlags(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UpdateOpportunityResponse{}
		resp.Data.UpdateOpportunity = types.Opportunity{
			ID:        "opp-123",
			Name:      "Flag Deal",
			Stage:     "Closed Won",
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create a fresh command to test flag handling
	testCmd := &cobra.Command{}
	testCmd.Flags().StringVar(&updateName, "name", "", "name")
	testCmd.Flags().Float64Var(&updateAmount, "amount", 0, "amount")
	testCmd.Flags().StringVar(&updateCurrency, "currency", "USD", "currency")
	testCmd.Flags().StringVar(&updateCloseDate, "close-date", "", "close date")
	testCmd.Flags().StringVar(&updateStage, "stage", "", "stage")
	testCmd.Flags().IntVar(&updateProbability, "probability", 0, "probability")
	testCmd.Flags().StringVar(&updateCompanyID, "company-id", "", "company ID")
	testCmd.Flags().StringVar(&updatePointOfContactID, "contact-id", "", "contact ID")
	testCmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data")

	// Save and restore update flags
	origName := updateName
	origAmount := updateAmount
	origCloseDate := updateCloseDate
	origStage := updateStage
	origProbability := updateProbability
	origCompanyID := updateCompanyID
	origContactID := updatePointOfContactID
	origData := updateData
	defer func() {
		updateName = origName
		updateAmount = origAmount
		updateCloseDate = origCloseDate
		updateStage = origStage
		updateProbability = origProbability
		updateCompanyID = origCompanyID
		updatePointOfContactID = origContactID
		updateData = origData
	}()

	updateData = ""

	// Set flags via Flags().Set() to mark them as changed
	testCmd.Flags().Set("name", "Flag Deal")
	testCmd.Flags().Set("amount", "75000")
	testCmd.Flags().Set("close-date", "2025-01-15")
	testCmd.Flags().Set("stage", "Closed Won")
	testCmd.Flags().Set("probability", "100")
	testCmd.Flags().Set("company-id", "new-company")
	testCmd.Flags().Set("contact-id", "new-contact")

	err := runUpdate(testCmd, []string{"opp-123"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}
}

func TestRunUpdate_TextOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UpdateOpportunityResponse{}
		resp.Data.UpdateOpportunity = types.Opportunity{
			ID:        "opp-123",
			Name:      "Text Output Deal",
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")

	// Save and restore update flags
	origData := updateData
	defer func() {
		updateData = origData
	}()

	updateData = `{"name":"Text Output Deal"}`

	err := runUpdate(updateCmd, []string{"opp-123"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Opportunity not found"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore update flags
	origData := updateData
	defer func() {
		updateData = origData
	}()

	updateData = `{"name":"Test"}`

	err := runUpdate(updateCmd, []string{"non-existent"})
	if err == nil {
		t.Error("expected error for API error")
	}
}

// TestRunDelete tests the runDelete function
func TestRunDelete_WithoutForce(t *testing.T) {
	// Save original value
	originalForce := forceDelete
	defer func() { forceDelete = originalForce }()

	forceDelete = false

	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called without --force")
	}))
	defer cleanup()

	// runDelete should return nil but print warning when force is not set
	err := runDelete(deleteCmd, []string{"test-id"})
	if err != nil {
		t.Errorf("Expected no error when --force is not set, got %v", err)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/rest/opportunities/opp-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := types.DeleteOpportunityResponse{}
		resp.Data.DeleteOpportunity = types.Opportunity{ID: "opp-123"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore forceDelete flag
	originalForce := forceDelete
	defer func() { forceDelete = originalForce }()

	forceDelete = true

	err := runDelete(deleteCmd, []string{"opp-123"})
	if err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}
}

func TestRunDelete_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Opportunity not found"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore forceDelete flag
	originalForce := forceDelete
	defer func() { forceDelete = originalForce }()

	forceDelete = true

	err := runDelete(deleteCmd, []string{"non-existent"})
	if err == nil {
		t.Error("expected error for API error")
	}
}

// TestListCmd_Execute tests the list command execution
func TestListCmd_Execute_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	opportunities := []types.Opportunity{
		{
			ID:        "opp-1",
			Name:      "Deal 1",
			Stage:     "Proposal",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.OpportunitiesListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.Opportunities = opportunities

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "json")

	listCmd.SetArgs([]string{})
	err := listCmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestListCmd_Execute_CSV(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	opportunities := []types.Opportunity{
		{
			ID:    "opp-1",
			Name:  "Deal 1",
			Stage: "Proposal",
			Amount: &types.Currency{
				AmountMicros: "50000000000",
				CurrencyCode: "USD",
			},
			Probability: 50,
			CloseDate:   "2024-12-31",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.OpportunitiesListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.Opportunities = opportunities

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "csv")

	listCmd.SetArgs([]string{})
	err := listCmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestListCmd_Execute_Table(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	opportunities := []types.Opportunity{
		{
			ID:    "opp-1-with-long-id-12345",
			Name:  "Big Deal",
			Stage: "Negotiation",
			Amount: &types.Currency{
				AmountMicros: "100000000000",
				CurrencyCode: "USD",
			},
			CloseDate: "2024-12-31",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.OpportunitiesListResponse{
			TotalCount: 1,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.Opportunities = opportunities

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")

	listCmd.SetArgs([]string{})
	err := listCmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

// TestCreateOpportunityInput tests the CreateOpportunityInput struct
func TestCreateOpportunityInput_FromJSON(t *testing.T) {
	jsonData := `{
		"name": "Test Deal",
		"stage": "Proposal",
		"probability": 75,
		"closeDate": "2024-12-31",
		"companyId": "company-1",
		"pointOfContactId": "contact-1"
	}`

	var input rest.CreateOpportunityInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name != "Test Deal" {
		t.Errorf("Name = %q, want %q", input.Name, "Test Deal")
	}
	if input.Stage != "Proposal" {
		t.Errorf("Stage = %q, want %q", input.Stage, "Proposal")
	}
	if input.Probability != 75 {
		t.Errorf("Probability = %d, want %d", input.Probability, 75)
	}
}

func TestCreateOpportunityInput_WithAmount(t *testing.T) {
	jsonData := `{
		"name": "Amount Deal",
		"amount": {
			"amountMicros": "50000000000",
			"currencyCode": "USD"
		}
	}`

	var input rest.CreateOpportunityInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Amount == nil {
		t.Fatal("Amount should not be nil")
	}
	if input.Amount.AmountMicros != "50000000000" {
		t.Errorf("AmountMicros = %q, want %q", input.Amount.AmountMicros, "50000000000")
	}
	if input.Amount.CurrencyCode != "USD" {
		t.Errorf("CurrencyCode = %q, want %q", input.Amount.CurrencyCode, "USD")
	}
}

// TestUpdateOpportunityInput tests the UpdateOpportunityInput struct
func TestUpdateOpportunityInput_FromJSON(t *testing.T) {
	jsonData := `{
		"name": "Updated Deal",
		"stage": "Closed Won"
	}`

	var input rest.UpdateOpportunityInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if *input.Name != "Updated Deal" {
		t.Errorf("Name = %q, want %q", *input.Name, "Updated Deal")
	}
}

func TestUpdateOpportunityInput_PartialUpdate(t *testing.T) {
	name := "Partial Deal"
	input := rest.UpdateOpportunityInput{
		Name: &name,
	}

	if input.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if *input.Name != "Partial Deal" {
		t.Errorf("Name = %q, want %q", *input.Name, "Partial Deal")
	}
	if input.Stage != nil {
		t.Error("Stage should be nil when not updated")
	}
	if input.Amount != nil {
		t.Error("Amount should be nil when not updated")
	}
}

// TestRunGet_CSVOutput tests get command with CSV output
func TestRunGet_CSVOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedOpp := types.Opportunity{
		ID:          "opp-csv",
		Name:        "CSV Deal",
		Stage:       "Proposal",
		Probability: 50,
		Amount: &types.Currency{
			AmountMicros: "25000000000",
			CurrencyCode: "EUR",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.OpportunityResponse{}
		resp.Data.Opportunity = expectedOpp

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "csv")

	err := runGet(getCmd, []string{"opp-csv"})
	if err != nil {
		t.Fatalf("runGet with CSV output failed: %v", err)
	}
}

// TestRunGet_TextOutput tests get command with text output
func TestRunGet_TextOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedOpp := types.Opportunity{
		ID:          "opp-text",
		Name:        "Text Deal",
		Stage:       "Negotiation",
		Probability: 75,
		CloseDate:   "2025-06-30",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.OpportunityResponse{}
		resp.Data.Opportunity = expectedOpp

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")

	err := runGet(getCmd, []string{"opp-text"})
	if err != nil {
		t.Fatalf("runGet with text output failed: %v", err)
	}
}

// TestRunCreate_WithZeroAmount tests create with zero amount (should not include amount)
func TestRunCreate_WithZeroAmount(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	requestBodyChecked := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that amount is not included when zero
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["amount"]; ok {
			t.Error("amount should not be included when zero")
		}
		requestBodyChecked = true

		resp := types.CreateOpportunityResponse{}
		resp.Data.CreateOpportunity = types.Opportunity{
			ID:        "new-opp-id",
			Name:      "No Amount Deal",
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
	origName := createName
	origAmount := createAmount
	origData := createData
	defer func() {
		createName = origName
		createAmount = origAmount
		createData = origData
	}()

	createName = "No Amount Deal"
	createAmount = 0 // Zero amount
	createData = ""

	err := runCreate(createCmd, []string{})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	if !requestBodyChecked {
		t.Error("request body was not checked")
	}
}
