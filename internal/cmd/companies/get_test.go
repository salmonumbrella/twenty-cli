package companies

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestGetCmd_Flags(t *testing.T) {
	if getCmd.Flags().Lookup("include") == nil {
		t.Error("include flag not registered")
	}
}

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
	// Command should require exactly 1 argument
	if getCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = getCmd.Args(getCmd, []string{"company-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = getCmd.Args(getCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRunGet_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 100
	revenue := 1000000

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies/company-123" {
			t.Errorf("expected path /rest/companies/company-123, got %s", r.URL.Path)
		}

		resp := types.CompanyResponse{}
		resp.Data.Company = types.Company{
			ID:   "company-123",
			Name: "Test Corp",
			DomainName: types.Link{
				PrimaryLinkUrl: "https://testcorp.com",
			},
			Address: types.Address{
				AddressCity: "San Francisco",
			},
			Employees:     &employees,
			AnnualRevenue: &revenue,
			IdealCustomer: true,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origInclude := getInclude
	defer func() { getInclude = origInclude }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	getInclude = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"company-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "company-123") {
		t.Errorf("output missing 'company-123': %s", output)
	}
	if !strings.Contains(output, "Test Corp") {
		t.Errorf("output missing 'Test Corp': %s", output)
	}
}

func TestRunGet_JSONOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 50

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CompanyResponse{}
		resp.Data.Company = types.Company{
			ID:        "company-json",
			Name:      "JSON Corp",
			Employees: &employees,
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origInclude := getInclude
	defer func() { getInclude = origInclude }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	getInclude = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"company-json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify it's valid JSON
	var result types.Company
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if result.ID != "company-json" {
		t.Errorf("ID = %q, want %q", result.ID, "company-json")
	}
	if result.Name != "JSON Corp" {
		t.Errorf("Name = %q, want %q", result.Name, "JSON Corp")
	}
}

func TestRunGet_CSVOutput(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 75

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CompanyResponse{}
		resp.Data.Company = types.Company{
			ID:   "company-csv",
			Name: "CSV Corp",
			DomainName: types.Link{
				PrimaryLinkUrl: "https://csvcorp.com",
			},
			Address: types.Address{
				AddressCity: "New York",
			},
			Employees: &employees,
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origInclude := getInclude
	defer func() { getInclude = origInclude }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	getInclude = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"company-csv"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify it's valid CSV
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
	expectedHeaders := []string{"id", "name", "domain", "employees", "city", "createdAt", "updatedAt"}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check data
	if records[1][0] != "company-csv" {
		t.Errorf("ID = %q, want %q", records[1][0], "company-csv")
	}
	if records[1][1] != "CSV Corp" {
		t.Errorf("Name = %q, want %q", records[1][1], "CSV Corp")
	}
}

func TestRunGet_WithInclude(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The include parameter is passed as "include[]" by the client
		resp := types.CompanyResponse{}
		resp.Data.Company = types.Company{
			ID:        "company-include",
			Name:      "Include Corp",
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Save original flag values
	origInclude := getInclude
	defer func() { getInclude = origInclude }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	getInclude = "people,opportunities"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"company-include"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify it completed successfully and contains expected output
	if !strings.Contains(output, "company-include") {
		t.Errorf("output missing 'company-include': %s", output)
	}
}

func TestRunGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Company not found"}}`))
	}))
	defer server.Close()

	// Save original flag values
	origInclude := getInclude
	defer func() { getInclude = origInclude }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	getInclude = ""

	err := runGet(getCmd, []string{"non-existent"})
	if err == nil {
		t.Fatal("expected error for non-existent company")
	}

	if !strings.Contains(err.Error(), "failed to get company") {
		t.Errorf("error message should contain 'failed to get company', got: %v", err)
	}
}

func TestOutputCompany_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 100
	c := &types.Company{
		ID:        "test-company",
		Name:      "Test Corp",
		Employees: &employees,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := outputCompany(c, "json", "")
	if err != nil {
		t.Fatalf("outputCompany() error = %v", err)
	}
}

func TestOutputCompany_CSV(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 50
	c := &types.Company{
		ID:   "csv-company",
		Name: "CSV Corp",
		DomainName: types.Link{
			PrimaryLinkUrl: "https://csvcorp.com",
		},
		Address: types.Address{
			AddressCity: "Boston",
		},
		Employees: &employees,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputCompany(c, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputCompany() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify it's valid CSV
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	_, err = reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}
}

func TestOutputCompany_Text(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees := 75
	revenue := 500000
	c := &types.Company{
		ID:   "text-company",
		Name: "Text Corp",
		DomainName: types.Link{
			PrimaryLinkUrl: "https://textcorp.com",
		},
		Address: types.Address{
			AddressCity: "Chicago",
		},
		Employees:     &employees,
		AnnualRevenue: &revenue,
		IdealCustomer: true,
		LinkedinLink: types.Link{
			PrimaryLinkUrl: "https://linkedin.com/company/textcorp",
		},
		XLink: types.Link{
			PrimaryLinkUrl: "https://x.com/textcorp",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputCompany(c, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputCompany() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected fields
	if !strings.Contains(output, "text-company") {
		t.Errorf("output missing ID: %s", output)
	}
	if !strings.Contains(output, "Text Corp") {
		t.Errorf("output missing Name: %s", output)
	}
}

func TestOutputCompany_DefaultFormat(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	c := &types.Company{
		ID:        "default-company",
		Name:      "Default Corp",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Empty string should use default (text) format
	err := outputCompany(c, "", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputCompany() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "default-company") {
		t.Errorf("output missing ID: %s", output)
	}
}

func TestOutputCompany_NilEmployees(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	c := &types.Company{
		ID:        "nil-emp-company",
		Name:      "Nil Employees Corp",
		Employees: nil, // nil employees
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputCompany(c, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputCompany() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify it's valid CSV and handles nil employees
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Check employees field is "0" when nil
	if len(records) >= 2 && records[1][3] != "0" {
		t.Errorf("employees should be '0' for nil, got %q", records[1][3])
	}
}
