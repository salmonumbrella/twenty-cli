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

func TestOutputList_JSON(t *testing.T) {
	data := []types.Person{
		{ID: "1", Name: types.Name{FirstName: "John", LastName: "Doe"}},
		{ID: "2", Name: types.Name{FirstName: "Jane", LastName: "Smith"}},
	}

	headers := []string{"ID", "NAME"}
	rowFunc := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName + " " + p.Name.LastName}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "json", "", headers, rowFunc, headers, rowFunc)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"id"`) || !strings.Contains(output, `"1"`) {
		t.Errorf("JSON output missing expected content: %s", output)
	}
}

func TestOutputList_YAML(t *testing.T) {
	data := []types.Person{
		{ID: "1", Name: types.Name{FirstName: "John"}},
	}

	headers := []string{"ID", "NAME"}
	rowFunc := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "yaml", "", headers, rowFunc, headers, rowFunc)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "id:") {
		t.Errorf("YAML output missing 'id:': %s", output)
	}
}

func TestOutputList_CSV(t *testing.T) {
	data := []types.Person{
		{ID: "1", Name: types.Name{FirstName: "John", LastName: "Doe"}},
		{ID: "2", Name: types.Name{FirstName: "Jane", LastName: "Smith"}},
	}

	headers := []string{"ID", "NAME"}
	rowFunc := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName + " " + p.Name.LastName}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "csv", "", headers, rowFunc, headers, rowFunc)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "ID,NAME") {
		t.Errorf("CSV output missing headers: %s", output)
	}
	if !strings.Contains(output, "1,John Doe") {
		t.Errorf("CSV output missing data row: %s", output)
	}
}

func TestOutputList_Table(t *testing.T) {
	data := []types.Person{
		{ID: "1", Name: types.Name{FirstName: "John"}},
	}

	headers := []string{"ID", "NAME"}
	rowFunc := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "text", "", headers, rowFunc, headers, rowFunc)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") {
		t.Errorf("table output missing headers: %s", output)
	}
	if !strings.Contains(output, "1") || !strings.Contains(output, "John") {
		t.Errorf("table output missing data: %s", output)
	}
}

func TestOutputList_EmptyData(t *testing.T) {
	var data []types.Person

	headers := []string{"ID", "NAME"}
	rowFunc := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "json", "", headers, rowFunc, headers, rowFunc)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() with empty data error = %v", err)
	}

	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	// Empty slice should output "[]" or "null"
	if output != "[]" && output != "null" {
		t.Errorf("expected empty array output, got: %s", output)
	}
}

func TestOutputList_WithQuery(t *testing.T) {
	data := []types.Person{
		{ID: "1", Name: types.Name{FirstName: "John"}},
		{ID: "2", Name: types.Name{FirstName: "Jane"}},
	}

	headers := []string{"ID", "NAME"}
	rowFunc := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "json", ".[0].id", headers, rowFunc, headers, rowFunc)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() with query error = %v", err)
	}

	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"1"` {
		t.Errorf("query result = %q, want %q", output, `"1"`)
	}
}

func TestOutputList_CustomCSV(t *testing.T) {
	data := []types.Person{
		{ID: "1", Name: types.Name{FirstName: "John", LastName: "Doe"}, Email: types.Email{PrimaryEmail: "john@example.com"}},
	}

	// Table shows ID and NAME
	tableHeaders := []string{"ID", "NAME"}
	tableRow := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName}
	}

	// CSV shows ID, NAME, and EMAIL
	csvHeaders := []string{"ID", "NAME", "EMAIL"}
	csvRow := func(p types.Person) []string {
		return []string{p.ID, p.Name.FirstName + " " + p.Name.LastName, p.Email.PrimaryEmail}
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputList(data, "csv", "", tableHeaders, tableRow, csvHeaders, csvRow)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputList() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "ID,NAME,EMAIL") {
		t.Errorf("CSV output missing custom headers: %s", output)
	}
	if !strings.Contains(output, "john@example.com") {
		t.Errorf("CSV output missing email: %s", output)
	}
}

func TestRunList_Success(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Twenty API returns {"data":{"people":[...]}} format
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"people": []map[string]interface{}{
					{"id": "1", "name": map[string]interface{}{"firstName": "John", "lastName": "Doe"}},
					{"id": "2", "name": map[string]interface{}{"firstName": "Jane", "lastName": "Smith"}},
				},
			},
			"pageInfo": map[string]interface{}{
				"hasNextPage": false,
			},
			"totalCount": 2,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set up environment and viper
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := ListConfig[types.Person]{
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			return client.ListPeople(ctx, opts)
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	opts := &ListOptions{
		Limit: 20,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(cfg, opts)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Output should contain the data
	if !strings.Contains(output, "John") && !strings.Contains(output, "1") {
		t.Errorf("output missing expected data: %s", output)
	}
}

func TestRunList_WithPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var response map[string]interface{}
		if callCount == 1 {
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"people": []map[string]interface{}{
						{"id": "1", "name": map[string]interface{}{"firstName": "John"}},
					},
				},
				"pageInfo": map[string]interface{}{
					"hasNextPage": true,
					"endCursor":   "cursor-1",
				},
				"totalCount": 2,
			}
		} else {
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"people": []map[string]interface{}{
						{"id": "2", "name": map[string]interface{}{"firstName": "Jane"}},
					},
				},
				"pageInfo": map[string]interface{}{
					"hasNextPage": false,
				},
				"totalCount": 2,
			}
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

	cfg := ListConfig[types.Person]{
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			return client.ListPeople(ctx, opts)
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	opts := &ListOptions{
		Limit: 20,
		All:   true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(cfg, opts)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Both pages should be present
	if !strings.Contains(output, "John") || !strings.Contains(output, "Jane") {
		t.Errorf("output missing paginated data: %s", output)
	}
}

func TestRunList_WithBuildFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"people": []map[string]interface{}{},
			},
			"pageInfo":   map[string]interface{}{"hasNextPage": false},
			"totalCount": 0,
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

	cfg := ListConfig[types.Person]{
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			// Verify filter was applied
			if opts.Filter == nil {
				t.Error("expected filter to be set")
			} else if opts.Filter["email"] == nil {
				t.Error("expected email filter to be set")
			}
			return client.ListPeople(ctx, opts)
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
		BuildFilter: func() map[string]interface{} {
			return map[string]interface{}{
				"email": map[string]interface{}{
					"like": "%@example.com",
				},
			}
		},
	}

	opts := &ListOptions{
		Limit: 20,
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(cfg, opts)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}
}

func TestRunList_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := ListConfig[types.Person]{
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			return client.ListPeople(ctx, opts)
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	opts := &ListOptions{
		Limit: 20,
	}

	err := runList(cfg, opts)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunList_InvalidFilter(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	cfg := ListConfig[types.Person]{
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			return nil, nil
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	opts := &ListOptions{
		Limit:  20,
		Filter: `{invalid json}`,
	}

	err := runList(cfg, opts)
	if err == nil {
		t.Fatal("expected error for invalid filter")
	}
}

func TestRunList_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	cfg := ListConfig[types.Person]{
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			t.Error("ListFunc should not be called when auth fails")
			return nil, nil
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	opts := &ListOptions{Limit: 20}

	err := runList(cfg, opts)
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}
