package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestOutputGet_JSON(t *testing.T) {
	data := types.Person{
		ID:   "123",
		Name: types.Name{FirstName: "John", LastName: "Doe"},
	}

	headers := []string{"FIELD", "VALUE"}
	rowsFunc := func(p types.Person) [][]string {
		return [][]string{
			{"ID", p.ID},
			{"Name", p.Name.FirstName + " " + p.Name.LastName},
		}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputGet(data, "json", "", headers, rowsFunc, nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputGet() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"id"`) {
		t.Errorf("JSON output missing 'id': %s", output)
	}
	if !strings.Contains(output, `"123"`) {
		t.Errorf("JSON output missing '123': %s", output)
	}
}

func TestOutputGet_YAML(t *testing.T) {
	data := types.Person{
		ID:   "456",
		Name: types.Name{FirstName: "Jane"},
	}

	headers := []string{"FIELD", "VALUE"}
	rowsFunc := func(p types.Person) [][]string {
		return [][]string{
			{"ID", p.ID},
		}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputGet(data, "yaml", "", headers, rowsFunc, nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputGet() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "id:") {
		t.Errorf("YAML output missing 'id:': %s", output)
	}
}

func TestOutputGet_CSV(t *testing.T) {
	data := types.Person{
		ID:   "789",
		Name: types.Name{FirstName: "Bob", LastName: "Wilson"},
	}

	headers := []string{"FIELD", "VALUE"}
	rowsFunc := func(p types.Person) [][]string {
		return [][]string{
			{"ID", p.ID},
			{"Name", p.Name.FirstName + " " + p.Name.LastName},
		}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputGet(data, "csv", "", headers, rowsFunc, nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputGet() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "FIELD,VALUE") {
		t.Errorf("CSV output missing headers: %s", output)
	}
	if !strings.Contains(output, "ID,789") {
		t.Errorf("CSV output missing ID row: %s", output)
	}
}

func TestOutputGet_Table(t *testing.T) {
	data := types.Person{
		ID:   "111",
		Name: types.Name{FirstName: "Alice"},
	}

	headers := []string{"FIELD", "VALUE"}
	rowsFunc := func(p types.Person) [][]string {
		return [][]string{
			{"ID", p.ID},
			{"Name", p.Name.FirstName},
		}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputGet(data, "text", "", headers, rowsFunc, nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputGet() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "ID") {
		t.Errorf("table output missing 'ID': %s", output)
	}
	if !strings.Contains(output, "111") {
		t.Errorf("table output missing '111': %s", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("table output missing 'Alice': %s", output)
	}
}

func TestOutputGet_WithQuery(t *testing.T) {
	data := types.Person{
		ID:   "999",
		Name: types.Name{FirstName: "Query", LastName: "Test"},
	}

	headers := []string{"FIELD", "VALUE"}
	rowsFunc := func(p types.Person) [][]string {
		return [][]string{{"ID", p.ID}}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputGet(data, "json", ".id", headers, rowsFunc, nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputGet() with query error = %v", err)
	}

	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"999"` {
		t.Errorf("query result = %q, want %q", output, `"999"`)
	}
}

func TestRunGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if !strings.HasSuffix(r.URL.Path, "/rest/people/test-id") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Twenty API returns {"data":{"person":{...}}} format
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"person": map[string]interface{}{
					"id": "test-id",
					"name": map[string]interface{}{
						"firstName": "John",
						"lastName":  "Doe",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := GetConfig[types.Person]{
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			return client.GetPerson(ctx, id, nil)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
				{"Name", p.Name.FirstName + " " + p.Name.LastName},
			}
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(cfg, "test-id")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test-id") {
		t.Errorf("output missing 'test-id': %s", output)
	}
}

func TestRunGet_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := GetConfig[types.Person]{
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			return client.GetPerson(ctx, id, nil)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{{"ID", p.ID}}
		},
	}

	err := runGet(cfg, "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunGet_NoToken(t *testing.T) {
	// Clear any environment token and use a nonexistent profile
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	cfg := GetConfig[types.Person]{
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			// This should never be called if auth fails
			t.Error("GetFunc should not be called when auth fails")
			return &types.Person{ID: id}, nil
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{{"ID", p.ID}}
		},
	}

	err := runGet(cfg, "test-id")
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunGet_TableOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"person": map[string]interface{}{
					"id": "table-test-id",
					"name": map[string]interface{}{
						"firstName": "Table",
						"lastName":  "Test",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := GetConfig[types.Person]{
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			return client.GetPerson(ctx, id, nil)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
				{"First Name", p.Name.FirstName},
				{"Last Name", p.Name.LastName},
			}
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(cfg, "table-test-id")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "ID") || !strings.Contains(output, "table-test-id") {
		t.Errorf("table output missing ID row: %s", output)
	}
	if !strings.Contains(output, "First Name") || !strings.Contains(output, "Table") {
		t.Errorf("table output missing First Name row: %s", output)
	}
}

func TestRunGet_CSVOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"person": map[string]interface{}{
					"id": "csv-test-id",
					"name": map[string]interface{}{
						"firstName": "CSV",
						"lastName":  "Test",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := GetConfig[types.Person]{
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			return client.GetPerson(ctx, id, nil)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
				{"Name", p.Name.FirstName + " " + p.Name.LastName},
			}
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(cfg, "csv-test-id")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "FIELD,VALUE") {
		t.Errorf("CSV output missing headers: %s", output)
	}
	if !strings.Contains(output, "ID,csv-test-id") {
		t.Errorf("CSV output missing ID row: %s", output)
	}
}

func TestRunGet_YAMLOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"person": map[string]interface{}{
					"id": "yaml-test-id",
					"name": map[string]interface{}{
						"firstName": "YAML",
						"lastName":  "Test",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "yaml")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := GetConfig[types.Person]{
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			return client.GetPerson(ctx, id, nil)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{{"ID", p.ID}}
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(cfg, "yaml-test-id")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "id:") || !strings.Contains(output, "yaml-test-id") {
		t.Errorf("YAML output missing expected content: %s", output)
	}
}
