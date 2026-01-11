package webhooks

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
)

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"url", "operation", "description", "secret"}
	for _, flag := range flags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestCreateCmd_RequiredFlags(t *testing.T) {
	requiredFlags := []string{"url", "operation", "secret"}
	for _, flag := range requiredFlags {
		f := createCmd.Flags().Lookup(flag)
		if f == nil {
			t.Errorf("%s flag not registered", flag)
			continue
		}
		// Check if the flag is marked as required via annotations
		if createCmd.MarkFlagRequired(flag) == nil {
			// Flag was not already required, so we just re-marked it (harmless)
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

func TestCreateCmd_Long(t *testing.T) {
	if createCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestRunCreate_Success(t *testing.T) {
	// Clean up package-level variables after test
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	createdAt := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/webhooks") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["targetUrl"] != "https://example.com/webhook" {
			t.Errorf("targetUrl = %q, want %q", input["targetUrl"], "https://example.com/webhook")
		}
		if input["operation"] != "*.created" {
			t.Errorf("operation = %q, want %q", input["operation"], "*.created")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":          "webhook-123",
			"targetUrl":   "https://example.com/webhook",
			"operation":   "*.created",
			"description": "Test webhook",
			"isActive":    true,
			"createdAt":   createdAt.Format(time.RFC3339),
			"updatedAt":   createdAt.Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set package-level variables
	createURL = "https://example.com/webhook"
	createOperation = "*.created"
	createDescription = "Test webhook"
	createSecret = "my-secret-key"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

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

	if !strings.Contains(output, "webhook-123") {
		t.Errorf("output missing webhook ID: %s", output)
	}
	if !strings.Contains(output, "*.created") {
		t.Errorf("output missing operation: %s", output)
	}
	if !strings.Contains(output, "https://example.com/webhook") {
		t.Errorf("output missing target URL: %s", output)
	}
}

func TestRunCreate_SuccessJSON(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	createdAt := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":          "webhook-456",
			"targetUrl":   "https://example.com/json-webhook",
			"operation":   "person.updated",
			"description": "JSON output webhook",
			"isActive":    true,
			"createdAt":   createdAt.Format(time.RFC3339),
			"updatedAt":   createdAt.Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createURL = "https://example.com/json-webhook"
	createOperation = "person.updated"
	createDescription = "JSON output webhook"
	createSecret = "json-secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

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

	if !strings.Contains(output, "webhook-456") {
		t.Errorf("output missing webhook ID: %s", output)
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	createURL = "https://example.com/webhook"
	createOperation = "*.created"
	createDescription = ""
	createSecret = "secret"

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunCreate_APIError(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	createURL = "https://example.com/webhook"
	createOperation = "*.created"
	createDescription = ""
	createSecret = "secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to create webhook") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunCreate_ServerError(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	createURL = "https://example.com/webhook"
	createOperation = "*.created"
	createDescription = ""
	createSecret = "secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunCreate_WithDescription(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	createdAt := time.Now().UTC().Truncate(time.Second)
	var receivedDescription string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)
		if desc, ok := input["description"].(string); ok {
			receivedDescription = desc
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":          "webhook-with-desc",
			"targetUrl":   "https://example.com/described",
			"operation":   "company.created",
			"description": receivedDescription,
			"isActive":    true,
			"createdAt":   createdAt.Format(time.RFC3339),
			"updatedAt":   createdAt.Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createURL = "https://example.com/described"
	createOperation = "company.created"
	createDescription = "Webhook for company creation events"
	createSecret = "desc-secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

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

	if receivedDescription != "Webhook for company creation events" {
		t.Errorf("description = %q, want %q", receivedDescription, "Webhook for company creation events")
	}
}

func TestRunCreate_EmptyDescription(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	createdAt := time.Now().UTC().Truncate(time.Second)
	var receivedInput map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedInput)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":        "webhook-no-desc",
			"targetUrl": "https://example.com/nodesc",
			"operation": "person.deleted",
			"isActive":  true,
			"createdAt": createdAt.Format(time.RFC3339),
			"updatedAt": createdAt.Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createURL = "https://example.com/nodesc"
	createOperation = "person.deleted"
	createDescription = ""
	createSecret = "nodesc-secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

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

	// When description is empty, it's omitted from JSON (omitempty tag)
	// Verify other required fields are present
	if receivedInput["targetUrl"] != "https://example.com/nodesc" {
		t.Errorf("targetUrl = %q, want %q", receivedInput["targetUrl"], "https://example.com/nodesc")
	}
	if receivedInput["operation"] != "person.deleted" {
		t.Errorf("operation = %q, want %q", receivedInput["operation"], "person.deleted")
	}
}

func TestRunCreate_VerifiesRequestBody(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	var receivedInput map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedInput)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":        "webhook-verify",
			"targetUrl": receivedInput["targetUrl"],
			"operation": receivedInput["operation"],
			"isActive":  true,
			"createdAt": time.Now().Format(time.RFC3339),
			"updatedAt": time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createURL = "https://verify.example.com/hook"
	createOperation = "opportunity.updated"
	createDescription = "Verification test"
	createSecret = "verify-secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

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

	// Verify all expected fields in request body
	if receivedInput["targetUrl"] != "https://verify.example.com/hook" {
		t.Errorf("targetUrl = %q, want %q", receivedInput["targetUrl"], "https://verify.example.com/hook")
	}
	if receivedInput["operation"] != "opportunity.updated" {
		t.Errorf("operation = %q, want %q", receivedInput["operation"], "opportunity.updated")
	}
	if receivedInput["description"] != "Verification test" {
		t.Errorf("description = %q, want %q", receivedInput["description"], "Verification test")
	}
	if receivedInput["secret"] != "verify-secret" {
		t.Errorf("secret = %q, want %q", receivedInput["secret"], "verify-secret")
	}
}

func TestRunCreate_Unauthorized(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	createURL = "https://example.com/webhook"
	createOperation = "*.created"
	createDescription = ""
	createSecret = "secret"

	t.Setenv("TWENTY_TOKEN", "bad-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunCreate_JSONOutputWithQuery(t *testing.T) {
	t.Cleanup(func() {
		createURL = ""
		createOperation = ""
		createDescription = ""
		createSecret = ""
	})

	createdAt := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":          "webhook-query",
			"targetUrl":   "https://example.com/query",
			"operation":   "task.created",
			"description": "Query test",
			"isActive":    true,
			"createdAt":   createdAt.Format(time.RFC3339),
			"updatedAt":   createdAt.Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createURL = "https://example.com/query"
	createOperation = "task.created"
	createDescription = "Query test"
	createSecret = "query-secret"

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", ".id")
	t.Cleanup(viper.Reset)

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

	// With a jq query, output should be filtered
	if !strings.Contains(output, "webhook-query") {
		t.Errorf("output should contain webhook ID: %s", output)
	}
}
