package graphql

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSchemaCmd_Use(t *testing.T) {
	if schemaCmd.Use != "schema" {
		t.Errorf("Use = %q, want %q", schemaCmd.Use, "schema")
	}
}

func TestSchemaCmd_Short(t *testing.T) {
	if schemaCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestSchemaCmd_EndpointFlag(t *testing.T) {
	flag := schemaCmd.Flags().Lookup("endpoint")
	if flag == nil {
		t.Error("schema command missing endpoint flag")
	}
	if flag.DefValue != "graphql" {
		t.Errorf("endpoint default = %q, want %q", flag.DefValue, "graphql")
	}
}

func TestRunSchema_Success(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/graphql") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify introspection query is sent
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		query, ok := body["query"].(string)
		if !ok {
			t.Error("query not included in request")
		}
		if !strings.Contains(query, "__schema") {
			t.Error("introspection query should contain __schema")
		}
		if !strings.Contains(query, "IntrospectionQuery") {
			t.Error("introspection query should contain IntrospectionQuery")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"__schema": {"queryType": {"name": "Query"}, "types": []}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSchema() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "__schema") {
		t.Errorf("output missing '__schema': %s", output)
	}
}

func TestRunSchema_MetadataEndpoint(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/metadata") {
			t.Errorf("unexpected path: %s, want suffix /metadata", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"__schema": {"queryType": {"name": "Query"}}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "metadata"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSchema() error = %v", err)
	}
}

func TestRunSchema_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runSchema()
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunSchema_APIError(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	err := runSchema()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunSchema_TextOutput(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"__schema": {"queryType": {"name": "Query"}}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSchema() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Even with text output, schema returns JSON
	if !strings.Contains(output, "__schema") {
		t.Errorf("output missing '__schema': %s", output)
	}
}

func TestRunSchema_EmptyOutput(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"__schema": {"queryType": null}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSchema() error = %v", err)
	}
}

func TestRunSchema_FullIntrospectionResponse(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	// Mock a more complete introspection response
	fullResponse := `{
		"data": {
			"__schema": {
				"queryType": {"name": "Query"},
				"mutationType": {"name": "Mutation"},
				"subscriptionType": null,
				"types": [
					{
						"kind": "OBJECT",
						"name": "Query",
						"description": "Root query type",
						"fields": [
							{
								"name": "users",
								"description": "List all users",
								"args": [],
								"type": {"kind": "LIST", "name": null, "ofType": {"kind": "OBJECT", "name": "User"}},
								"isDeprecated": false,
								"deprecationReason": null
							}
						]
					}
				],
				"directives": [
					{
						"name": "deprecated",
						"description": "Marks field as deprecated",
						"locations": ["FIELD_DEFINITION"],
						"args": []
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fullResponse))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSchema() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify response contains expected schema elements
	if !strings.Contains(output, "Query") {
		t.Errorf("output missing 'Query': %s", output)
	}
	if !strings.Contains(output, "Mutation") {
		t.Errorf("output missing 'Mutation': %s", output)
	}
	if !strings.Contains(output, "users") {
		t.Errorf("output missing 'users': %s", output)
	}
}

func TestIntrospectionQuery_Format(t *testing.T) {
	// Verify introspection query contains expected elements
	if !strings.Contains(introspectionQuery, "__schema") {
		t.Error("introspectionQuery should contain __schema")
	}
	if !strings.Contains(introspectionQuery, "queryType") {
		t.Error("introspectionQuery should contain queryType")
	}
	if !strings.Contains(introspectionQuery, "mutationType") {
		t.Error("introspectionQuery should contain mutationType")
	}
	if !strings.Contains(introspectionQuery, "subscriptionType") {
		t.Error("introspectionQuery should contain subscriptionType")
	}
	if !strings.Contains(introspectionQuery, "types") {
		t.Error("introspectionQuery should contain types")
	}
	if !strings.Contains(introspectionQuery, "directives") {
		t.Error("introspectionQuery should contain directives")
	}
	if !strings.Contains(introspectionQuery, "FullType") {
		t.Error("introspectionQuery should contain FullType fragment")
	}
	if !strings.Contains(introspectionQuery, "InputValue") {
		t.Error("introspectionQuery should contain InputValue fragment")
	}
	if !strings.Contains(introspectionQuery, "TypeRef") {
		t.Error("introspectionQuery should contain TypeRef fragment")
	}
}

func TestRunSchema_ServerError(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

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

	gqlEndpoint = "graphql"

	err := runSchema()
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunSchema_GraphQLErrors(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	// GraphQL errors are returned with 200 status but contain errors array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": null, "errors": [{"message": "introspection disabled"}]}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// GraphQL errors with 200 status are not HTTP errors
	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	// No HTTP error, but output contains the error
	if err != nil {
		t.Fatalf("runSchema() error = %v (GraphQL errors with 200 are not HTTP errors)", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "errors") {
		t.Errorf("output should contain GraphQL errors: %s", output)
	}
}

func TestRunSchema_WithJQQuery(t *testing.T) {
	origEndpoint := gqlEndpoint
	defer func() {
		gqlEndpoint = origEndpoint
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"__schema": {"queryType": {"name": "Query"}}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", ".data.__schema.queryType.name")
	t.Cleanup(viper.Reset)

	gqlEndpoint = "graphql"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchema()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSchema() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// With jq query, output should be filtered
	if !strings.Contains(output, "Query") {
		t.Errorf("output should contain 'Query': %s", output)
	}
}
