package graphql

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestQueryCmd_Use(t *testing.T) {
	if queryCmd.Use != "query" {
		t.Errorf("Use = %q, want %q", queryCmd.Use, "query")
	}
}

func TestQueryCmd_Short(t *testing.T) {
	if queryCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestMutateCmd_Use(t *testing.T) {
	if mutateCmd.Use != "mutate" {
		t.Errorf("Use = %q, want %q", mutateCmd.Use, "mutate")
	}
}

func TestMutateCmd_Short(t *testing.T) {
	if mutateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestQueryCmd_Flags(t *testing.T) {
	flags := []string{"query", "file", "variables", "variables-file", "operation", "endpoint"}
	for _, flag := range flags {
		if queryCmd.Flags().Lookup(flag) == nil {
			t.Errorf("query command missing flag %q", flag)
		}
	}
}

func TestMutateCmd_Flags(t *testing.T) {
	flags := []string{"query", "file", "variables", "variables-file", "operation", "endpoint"}
	for _, flag := range flags {
		if mutateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("mutate command missing flag %q", flag)
		}
	}
}

func TestNormalizeEndpoint_Empty(t *testing.T) {
	result := normalizeEndpoint("")
	if result != "/graphql" {
		t.Errorf("normalizeEndpoint(\"\") = %q, want %q", result, "/graphql")
	}
}

func TestNormalizeEndpoint_Whitespace(t *testing.T) {
	result := normalizeEndpoint("   ")
	if result != "/graphql" {
		t.Errorf("normalizeEndpoint(\"   \") = %q, want %q", result, "/graphql")
	}
}

func TestNormalizeEndpoint_NoSlash(t *testing.T) {
	result := normalizeEndpoint("metadata")
	if result != "/metadata" {
		t.Errorf("normalizeEndpoint(\"metadata\") = %q, want %q", result, "/metadata")
	}
}

func TestNormalizeEndpoint_WithSlash(t *testing.T) {
	result := normalizeEndpoint("/graphql")
	if result != "/graphql" {
		t.Errorf("normalizeEndpoint(\"/graphql\") = %q, want %q", result, "/graphql")
	}
}

func TestNormalizeEndpoint_GraphQL(t *testing.T) {
	result := normalizeEndpoint("graphql")
	if result != "/graphql" {
		t.Errorf("normalizeEndpoint(\"graphql\") = %q, want %q", result, "/graphql")
	}
}

func TestNormalizeEndpoint_WithWhitespace(t *testing.T) {
	result := normalizeEndpoint("  metadata  ")
	if result != "/metadata" {
		t.Errorf("normalizeEndpoint(\"  metadata  \") = %q, want %q", result, "/metadata")
	}
}

func TestReadQuery_NoQueryOrFile(t *testing.T) {
	// Save and restore global state
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	gqlQuery = ""
	gqlFile = ""

	_, err := readQuery()
	if err == nil {
		t.Fatal("expected error when no query or file specified")
	}
	if !strings.Contains(err.Error(), "missing GraphQL query") {
		t.Errorf("error = %q, want error containing 'missing GraphQL query'", err.Error())
	}
}

func TestReadQuery_WithQuery(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	gqlQuery = "{ users { id name } }"
	gqlFile = ""

	result, err := readQuery()
	if err != nil {
		t.Fatalf("readQuery() error = %v", err)
	}
	if result != "{ users { id name } }" {
		t.Errorf("readQuery() = %q, want %q", result, "{ users { id name } }")
	}
}

func TestReadQuery_FromFile(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	// Create temp file with query
	tmpDir := t.TempDir()
	queryFile := filepath.Join(tmpDir, "query.graphql")
	if err := os.WriteFile(queryFile, []byte("  { test { id } }  \n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	gqlQuery = ""
	gqlFile = queryFile

	result, err := readQuery()
	if err != nil {
		t.Fatalf("readQuery() error = %v", err)
	}
	// Should be trimmed
	if result != "{ test { id } }" {
		t.Errorf("readQuery() = %q, want %q", result, "{ test { id } }")
	}
}

func TestReadQuery_FileNotFound(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	gqlQuery = ""
	gqlFile = "/nonexistent/path/query.graphql"

	_, err := readQuery()
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "read query file") {
		t.Errorf("error = %q, want error containing 'read query file'", err.Error())
	}
}

func TestReadQuery_FilePrecedenceOverQuery(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	// Create temp file
	tmpDir := t.TempDir()
	queryFile := filepath.Join(tmpDir, "query.graphql")
	if err := os.WriteFile(queryFile, []byte("{ fromFile }"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	gqlQuery = "{ fromQuery }"
	gqlFile = queryFile

	result, err := readQuery()
	if err != nil {
		t.Fatalf("readQuery() error = %v", err)
	}
	// File should take precedence
	if result != "{ fromFile }" {
		t.Errorf("readQuery() = %q, want %q", result, "{ fromFile }")
	}
}

func TestReadVariables_Empty(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	gqlVars = ""
	gqlVarsFile = ""

	result, err := readVariables()
	if err != nil {
		t.Fatalf("readVariables() error = %v", err)
	}
	if result != nil {
		t.Errorf("readVariables() = %v, want nil", result)
	}
}

func TestReadVariables_FromString(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	gqlVars = `{"id": "123", "name": "test"}`
	gqlVarsFile = ""

	result, err := readVariables()
	if err != nil {
		t.Fatalf("readVariables() error = %v", err)
	}
	if result == nil {
		t.Fatal("readVariables() = nil, want non-nil")
	}
	if result["id"] != "123" {
		t.Errorf("result[\"id\"] = %v, want %q", result["id"], "123")
	}
	if result["name"] != "test" {
		t.Errorf("result[\"name\"] = %v, want %q", result["name"], "test")
	}
}

func TestReadVariables_FromFile(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	// Create temp file
	tmpDir := t.TempDir()
	varsFile := filepath.Join(tmpDir, "vars.json")
	if err := os.WriteFile(varsFile, []byte(`{"count": 42, "enabled": true}`), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	gqlVars = ""
	gqlVarsFile = varsFile

	result, err := readVariables()
	if err != nil {
		t.Fatalf("readVariables() error = %v", err)
	}
	if result == nil {
		t.Fatal("readVariables() = nil, want non-nil")
	}
	// JSON numbers are float64
	if result["count"] != float64(42) {
		t.Errorf("result[\"count\"] = %v, want %v", result["count"], float64(42))
	}
	if result["enabled"] != true {
		t.Errorf("result[\"enabled\"] = %v, want true", result["enabled"])
	}
}

func TestReadVariables_FileNotFound(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	gqlVars = ""
	gqlVarsFile = "/nonexistent/path/vars.json"

	_, err := readVariables()
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "read variables file") {
		t.Errorf("error = %q, want error containing 'read variables file'", err.Error())
	}
}

func TestReadVariables_InvalidJSON(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	gqlVars = "not valid json"
	gqlVarsFile = ""

	_, err := readVariables()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON variables") {
		t.Errorf("error = %q, want error containing 'invalid JSON variables'", err.Error())
	}
}

func TestReadVariables_InvalidJSONFromFile(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	// Create temp file with invalid JSON
	tmpDir := t.TempDir()
	varsFile := filepath.Join(tmpDir, "vars.json")
	if err := os.WriteFile(varsFile, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	gqlVars = ""
	gqlVarsFile = varsFile

	_, err := readVariables()
	if err == nil {
		t.Fatal("expected error for invalid JSON in file")
	}
	if !strings.Contains(err.Error(), "invalid JSON variables") {
		t.Errorf("error = %q, want error containing 'invalid JSON variables'", err.Error())
	}
}

func TestReadVariables_FilePrecedenceOverString(t *testing.T) {
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	// Create temp file
	tmpDir := t.TempDir()
	varsFile := filepath.Join(tmpDir, "vars.json")
	if err := os.WriteFile(varsFile, []byte(`{"source": "file"}`), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	gqlVars = `{"source": "string"}`
	gqlVarsFile = varsFile

	result, err := readVariables()
	if err != nil {
		t.Fatalf("readVariables() error = %v", err)
	}
	// File should take precedence
	if result["source"] != "file" {
		t.Errorf("result[\"source\"] = %v, want %q", result["source"], "file")
	}
}

func TestRunGraphQL_Success(t *testing.T) {
	// Save and restore global state
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/graphql") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if body["query"] != "{ users { id } }" {
			t.Errorf("query = %q, want %q", body["query"], "{ users { id } }")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"users": [{"id": "1"}]}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ users { id } }"
	gqlFile = ""
	gqlVars = ""
	gqlVarsFile = ""
	gqlOperation = ""
	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraphQL()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGraphQL() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "users") {
		t.Errorf("output missing 'users': %s", output)
	}
}

func TestRunGraphQL_WithVariables(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		// Verify variables are included
		vars, ok := body["variables"].(map[string]interface{})
		if !ok {
			t.Error("variables not included in request")
		}
		if vars["id"] != "123" {
			t.Errorf("variables[\"id\"] = %v, want %q", vars["id"], "123")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"user": {"id": "123"}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "query GetUser($id: ID!) { user(id: $id) { id } }"
	gqlFile = ""
	gqlVars = `{"id": "123"}`
	gqlVarsFile = ""
	gqlOperation = ""
	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraphQL()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGraphQL() error = %v", err)
	}
}

func TestRunGraphQL_WithOperationName(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		// Verify operationName is included
		if body["operationName"] != "GetUsers" {
			t.Errorf("operationName = %v, want %q", body["operationName"], "GetUsers")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"users": []}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "query GetUsers { users { id } }"
	gqlFile = ""
	gqlVars = ""
	gqlVarsFile = ""
	gqlOperation = "GetUsers"
	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraphQL()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGraphQL() error = %v", err)
	}
}

func TestRunGraphQL_CustomEndpoint(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/metadata") {
			t.Errorf("unexpected path: %s, want suffix /metadata", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ __schema { types { name } } }"
	gqlFile = ""
	gqlVars = ""
	gqlVarsFile = ""
	gqlOperation = ""
	gqlEndpoint = "metadata"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraphQL()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGraphQL() error = %v", err)
	}
}

func TestRunGraphQL_NoToken(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ test }"
	gqlFile = ""

	err := runGraphQL()
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunGraphQL_NoQuery(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	t.Cleanup(viper.Reset)

	gqlQuery = ""
	gqlFile = ""

	err := runGraphQL()
	if err == nil {
		t.Fatal("expected error when no query is provided")
	}
	if !strings.Contains(err.Error(), "missing GraphQL query") {
		t.Errorf("error = %q, want error containing 'missing GraphQL query'", err.Error())
	}
}

func TestRunGraphQL_APIError(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors": [{"message": "syntax error"}]}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ invalid syntax"
	gqlFile = ""
	gqlVars = ""
	gqlVarsFile = ""
	gqlOperation = ""
	gqlEndpoint = "graphql"

	err := runGraphQL()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunGraphQL_TextOutput(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"test": "value"}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ test }"
	gqlFile = ""
	gqlVars = ""
	gqlVarsFile = ""
	gqlOperation = ""
	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraphQL()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGraphQL() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Text output should still be JSON for raw GraphQL responses
	if !strings.Contains(output, "test") {
		t.Errorf("output missing 'test': %s", output)
	}
}

func TestRunGraphQL_EmptyOutput(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	origOperation := gqlOperation
	origEndpoint := gqlEndpoint
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
		gqlOperation = origOperation
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": null}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ test }"
	gqlFile = ""
	gqlVars = ""
	gqlVarsFile = ""
	gqlOperation = ""
	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraphQL()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGraphQL() error = %v", err)
	}
}

func TestRunGraphQL_InvalidVariables(t *testing.T) {
	origQuery := gqlQuery
	origFile := gqlFile
	origVars := gqlVars
	origVarsFile := gqlVarsFile
	defer func() {
		gqlQuery = origQuery
		gqlFile = origFile
		gqlVars = origVars
		gqlVarsFile = origVarsFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	t.Cleanup(viper.Reset)

	gqlQuery = "{ test }"
	gqlFile = ""
	gqlVars = "invalid json"
	gqlVarsFile = ""

	err := runGraphQL()
	if err == nil {
		t.Fatal("expected error for invalid variables JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON variables") {
		t.Errorf("error = %q, want error containing 'invalid JSON variables'", err.Error())
	}
}
