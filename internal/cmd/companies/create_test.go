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

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"name", "domain", "address", "employees", "linkedin", "x-link", "revenue", "ideal-customer", "data"}
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

func TestCreateCompanyInput_FromFlags(t *testing.T) {
	input := rest.CreateCompanyInput{
		Name:          "Test Corp",
		Address:       "123 Main St",
		Employees:     50,
		AnnualRevenue: 1000000,
		IdealCustomer: true,
	}

	if input.Name != "Test Corp" {
		t.Errorf("Name = %q, want %q", input.Name, "Test Corp")
	}
	if input.Address != "123 Main St" {
		t.Errorf("Address = %q, want %q", input.Address, "123 Main St")
	}
	if input.Employees != 50 {
		t.Errorf("Employees = %d, want %d", input.Employees, 50)
	}
	if input.AnnualRevenue != 1000000 {
		t.Errorf("AnnualRevenue = %d, want %d", input.AnnualRevenue, 1000000)
	}
	if !input.IdealCustomer {
		t.Error("IdealCustomer should be true")
	}
}

func TestCreateCompanyInput_FromJSON(t *testing.T) {
	jsonData := `{
		"name": "JSON Corp",
		"domainName": {"primaryLinkUrl": "https://jsoncorp.com"},
		"employees": 100,
		"idealCustomerProfile": true
	}`

	var input rest.CreateCompanyInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name != "JSON Corp" {
		t.Errorf("Name = %q, want %q", input.Name, "JSON Corp")
	}
	if input.DomainName == nil {
		t.Fatal("DomainName should not be nil")
	}
	if input.DomainName.PrimaryLinkUrl != "https://jsoncorp.com" {
		t.Errorf("DomainName.PrimaryLinkUrl = %q, want %q", input.DomainName.PrimaryLinkUrl, "https://jsoncorp.com")
	}
}

func TestCreateCompanyInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input rest.CreateCompanyInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCreateCompanyInput_EmptyFields(t *testing.T) {
	input := rest.CreateCompanyInput{}

	if input.Name != "" {
		t.Errorf("Name should be empty, got %q", input.Name)
	}
	if input.DomainName != nil {
		t.Errorf("DomainName should be nil, got %v", input.DomainName)
	}
}

func TestCreateCompanyInput_WithLinks(t *testing.T) {
	input := rest.CreateCompanyInput{
		Name: "Link Corp",
		DomainName: &rest.LinkInput{
			PrimaryLinkUrl: "https://linkcorp.com",
		},
		LinkedinLink: &rest.LinkInput{
			PrimaryLinkUrl: "https://linkedin.com/company/linkcorp",
		},
		XLink: &rest.LinkInput{
			PrimaryLinkUrl: "https://x.com/linkcorp",
		},
	}

	if input.DomainName == nil || input.DomainName.PrimaryLinkUrl != "https://linkcorp.com" {
		t.Errorf("DomainName = %v, want https://linkcorp.com", input.DomainName)
	}
	if input.LinkedinLink == nil || input.LinkedinLink.PrimaryLinkUrl != "https://linkedin.com/company/linkcorp" {
		t.Errorf("LinkedinLink = %v, want https://linkedin.com/company/linkcorp", input.LinkedinLink)
	}
	if input.XLink == nil || input.XLink.PrimaryLinkUrl != "https://x.com/linkcorp" {
		t.Errorf("XLink = %v, want https://x.com/linkcorp", input.XLink)
	}
}

func TestRunCreate_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 50

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies" {
			t.Errorf("expected path /rest/companies, got %s", r.URL.Path)
		}

		resp := types.CreateCompanyResponse{}
		resp.Data.CreateCompany = types.Company{
			ID:        "new-company-id",
			Name:      "Test Corp",
			Employees: &employees,
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origName := createName
	origData := createData
	defer func() {
		createName = origName
		createData = origData
	}()

	// Set up environment
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Set flag values
	createName = "Test Corp"
	createData = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Created company") {
		t.Errorf("output missing 'Created company': %s", output)
	}
	if !strings.Contains(output, "new-company-id") {
		t.Errorf("output missing 'new-company-id': %s", output)
	}
}

func TestRunCreate_WithJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 100

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreateCompanyResponse{}
		resp.Data.CreateCompany = types.Company{
			ID:        "json-company-id",
			Name:      "JSON Corp",
			Employees: &employees,
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origData := createData
	defer func() { createData = origData }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "JSON Corp", "employees": 100}`

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify it's valid JSON
	var result types.Company
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if result.ID != "json-company-id" {
		t.Errorf("ID = %q, want %q", result.ID, "json-company-id")
	}
}

func TestRunCreate_InvalidJSON(t *testing.T) {
	// Save original flag values
	origData := createData
	defer func() { createData = origData }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{invalid json}`

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "invalid JSON data") {
		t.Errorf("error message should contain 'invalid JSON data', got: %v", err)
	}
}

func TestRunCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Name is required"}}`))
	}))
	defer server.Close()

	// Save original flag values
	origName := createName
	defer func() { createName = origName }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createName = "Test Corp"

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	if !strings.Contains(err.Error(), "failed to create company") {
		t.Errorf("error message should contain 'failed to create company', got: %v", err)
	}
}

func TestRunCreate_WithAllFlags(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 75
	revenue := 5000000

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CreateCompanyResponse{}
		resp.Data.CreateCompany = types.Company{
			ID:            "full-company-id",
			Name:          "Full Corp",
			Employees:     &employees,
			AnnualRevenue: &revenue,
			IdealCustomer: true,
			DomainName: types.Link{
				PrimaryLinkUrl: "https://fullcorp.com",
			},
			LinkedinLink: types.Link{
				PrimaryLinkUrl: "https://linkedin.com/company/fullcorp",
			},
			XLink: types.Link{
				PrimaryLinkUrl: "https://x.com/fullcorp",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origName := createName
	origDomain := createDomain
	origAddress := createAddress
	origEmployees := createEmployees
	origLinkedin := createLinkedinLink
	origXLink := createXLink
	origRevenue := createAnnualRevenue
	origIdeal := createIdealCustomer
	origData := createData
	defer func() {
		createName = origName
		createDomain = origDomain
		createAddress = origAddress
		createEmployees = origEmployees
		createLinkedinLink = origLinkedin
		createXLink = origXLink
		createAnnualRevenue = origRevenue
		createIdealCustomer = origIdeal
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createName = "Full Corp"
	createDomain = "https://fullcorp.com"
	createAddress = "456 Full Ave"
	createEmployees = 75
	createLinkedinLink = "https://linkedin.com/company/fullcorp"
	createXLink = "https://x.com/fullcorp"
	createAnnualRevenue = 5000000
	createIdealCustomer = true
	createData = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "full-company-id") {
		t.Errorf("output missing 'full-company-id': %s", output)
	}
}
