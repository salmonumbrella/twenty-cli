package companies

import (
	"bytes"
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
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"name", "domain", "address", "employees", "linkedin", "x-link", "revenue", "ideal-customer", "data"}
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
	// Command should require exactly 1 argument
	if updateCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = updateCmd.Args(updateCmd, []string{"company-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = updateCmd.Args(updateCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
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

func TestUpdateCompanyInput_FromJSON(t *testing.T) {
	jsonData := `{
		"name": "Updated Corp",
		"employees": 200,
		"idealCustomerProfile": true
	}`

	var input rest.UpdateCompanyInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name == nil || *input.Name != "Updated Corp" {
		t.Errorf("Name = %v, want 'Updated Corp'", input.Name)
	}
	if input.Employees == nil || *input.Employees != 200 {
		t.Errorf("Employees = %v, want 200", input.Employees)
	}
}

func TestUpdateCompanyInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input rest.UpdateCompanyInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestUpdateCompanyInput_EmptyInput(t *testing.T) {
	input := rest.UpdateCompanyInput{}

	if input.Name != nil {
		t.Error("Name should be nil for empty input")
	}
	if input.Employees != nil {
		t.Error("Employees should be nil for empty input")
	}
	if input.DomainName != nil {
		t.Error("DomainName should be nil for empty input")
	}
}

func TestUpdateCompanyInput_PartialUpdate(t *testing.T) {
	name := "New Name"
	input := rest.UpdateCompanyInput{
		Name: &name,
	}

	if input.Name == nil || *input.Name != "New Name" {
		t.Errorf("Name = %v, want 'New Name'", input.Name)
	}
	if input.Employees != nil {
		t.Error("Employees should be nil when not updated")
	}
}

func TestRunUpdate_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 150

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies/company-123" {
			t.Errorf("expected path /rest/companies/company-123, got %s", r.URL.Path)
		}

		resp := types.UpdateCompanyResponse{}
		resp.Data.UpdateCompany = types.Company{
			ID:        "company-123",
			Name:      "Updated Corp",
			Employees: &employees,
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origData := updateData
	defer func() { updateData = origData }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"name": "Updated Corp", "employees": 150}`

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"company-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Updated company") {
		t.Errorf("output missing 'Updated company': %s", output)
	}
	if !strings.Contains(output, "company-123") {
		t.Errorf("output missing 'company-123': %s", output)
	}
}

func TestRunUpdate_JSONOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 200

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UpdateCompanyResponse{}
		resp.Data.UpdateCompany = types.Company{
			ID:        "json-update",
			Name:      "JSON Updated Corp",
			Employees: &employees,
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origData := updateData
	defer func() { updateData = origData }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"name": "JSON Updated Corp"}`

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"json-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify it's valid JSON
	var result types.Company
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if result.ID != "json-update" {
		t.Errorf("ID = %q, want %q", result.ID, "json-update")
	}
}

func TestRunUpdate_InvalidJSON(t *testing.T) {
	// Save original flag values
	origData := updateData
	defer func() { updateData = origData }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{invalid json}`

	err := runUpdate(updateCmd, []string{"company-123"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "invalid JSON data") {
		t.Errorf("error message should contain 'invalid JSON data', got: %v", err)
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Company not found"}}`))
	}))
	defer server.Close()

	// Save original flag values
	origData := updateData
	defer func() { updateData = origData }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"name": "Test"}`

	err := runUpdate(updateCmd, []string{"non-existent"})
	if err == nil {
		t.Fatal("expected error for non-existent company")
	}

	if !strings.Contains(err.Error(), "failed to update company") {
		t.Errorf("error message should contain 'failed to update company', got: %v", err)
	}
}

func TestRunUpdate_WithFlagsChanged(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 100

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UpdateCompanyResponse{}
		resp.Data.UpdateCompany = types.Company{
			ID:        "flag-update",
			Name:      "Flag Updated Corp",
			Employees: &employees,
			DomainName: types.Link{
				PrimaryLinkUrl: "https://flagupdated.com",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origName := updateName
	origDomain := updateDomain
	origData := updateData
	defer func() {
		updateName = origName
		updateDomain = origDomain
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Create a new command to test flag changes
	testCmd := &cobra.Command{
		Use:  "update <id>",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}
	testCmd.Flags().StringVar(&updateName, "name", "", "company name")
	testCmd.Flags().StringVar(&updateDomain, "domain", "", "domain name")
	testCmd.Flags().StringVar(&updateData, "data", "", "JSON data")

	// Simulate setting flags
	testCmd.Flags().Set("name", "Flag Updated Corp")
	testCmd.Flags().Set("domain", "https://flagupdated.com")
	updateData = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(testCmd, []string{"flag-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "flag-update") {
		t.Errorf("output missing 'flag-update': %s", output)
	}
}

func TestRunUpdate_AllFlags(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 250
	revenue := 10000000

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UpdateCompanyResponse{}
		resp.Data.UpdateCompany = types.Company{
			ID:            "all-flags",
			Name:          "All Flags Corp",
			Employees:     &employees,
			AnnualRevenue: &revenue,
			IdealCustomer: true,
			DomainName: types.Link{
				PrimaryLinkUrl: "https://allflags.com",
			},
			LinkedinLink: types.Link{
				PrimaryLinkUrl: "https://linkedin.com/company/allflags",
			},
			XLink: types.Link{
				PrimaryLinkUrl: "https://x.com/allflags",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origName := updateName
	origDomain := updateDomain
	origAddress := updateAddress
	origEmployees := updateEmployees
	origLinkedin := updateLinkedinLink
	origXLink := updateXLink
	origRevenue := updateAnnualRevenue
	origIdeal := updateIdealCustomer
	origData := updateData
	defer func() {
		updateName = origName
		updateDomain = origDomain
		updateAddress = origAddress
		updateEmployees = origEmployees
		updateLinkedinLink = origLinkedin
		updateXLink = origXLink
		updateAnnualRevenue = origRevenue
		updateIdealCustomer = origIdeal
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Create a new command to test flag changes
	testCmd := &cobra.Command{
		Use:  "update <id>",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}
	testCmd.Flags().StringVar(&updateName, "name", "", "company name")
	testCmd.Flags().StringVar(&updateDomain, "domain", "", "domain name")
	testCmd.Flags().StringVar(&updateAddress, "address", "", "address")
	testCmd.Flags().IntVar(&updateEmployees, "employees", 0, "number of employees")
	testCmd.Flags().StringVar(&updateLinkedinLink, "linkedin", "", "LinkedIn URL")
	testCmd.Flags().StringVar(&updateXLink, "x-link", "", "X URL")
	testCmd.Flags().IntVar(&updateAnnualRevenue, "revenue", 0, "annual revenue")
	testCmd.Flags().BoolVar(&updateIdealCustomer, "ideal-customer", false, "ideal customer")
	testCmd.Flags().StringVar(&updateData, "data", "", "JSON data")

	// Set all flags
	testCmd.Flags().Set("name", "All Flags Corp")
	testCmd.Flags().Set("domain", "https://allflags.com")
	testCmd.Flags().Set("address", "789 All Flags Blvd")
	testCmd.Flags().Set("employees", "250")
	testCmd.Flags().Set("linkedin", "https://linkedin.com/company/allflags")
	testCmd.Flags().Set("x-link", "https://x.com/allflags")
	testCmd.Flags().Set("revenue", "10000000")
	testCmd.Flags().Set("ideal-customer", "true")
	updateData = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(testCmd, []string{"all-flags"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "all-flags") {
		t.Errorf("output missing 'all-flags': %s", output)
	}
}

func TestUpdateCompanyInput_WithLinks(t *testing.T) {
	input := rest.UpdateCompanyInput{
		DomainName: &rest.LinkInput{
			PrimaryLinkUrl: "https://newdomain.com",
		},
		LinkedinLink: &rest.LinkInput{
			PrimaryLinkUrl: "https://linkedin.com/company/new",
		},
		XLink: &rest.LinkInput{
			PrimaryLinkUrl: "https://x.com/new",
		},
	}

	if input.DomainName == nil || input.DomainName.PrimaryLinkUrl != "https://newdomain.com" {
		t.Errorf("DomainName = %v, want https://newdomain.com", input.DomainName)
	}
	if input.LinkedinLink == nil || input.LinkedinLink.PrimaryLinkUrl != "https://linkedin.com/company/new" {
		t.Errorf("LinkedinLink = %v, want https://linkedin.com/company/new", input.LinkedinLink)
	}
	if input.XLink == nil || input.XLink.PrimaryLinkUrl != "https://x.com/new" {
		t.Errorf("XLink = %v, want https://x.com/new", input.XLink)
	}
}
