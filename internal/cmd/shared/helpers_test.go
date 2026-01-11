package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestReadJSONMap_DirectJSON(t *testing.T) {
	input := `{"name": "test", "count": 42}`

	result, err := ReadJSONMap(input, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if name, ok := result["name"].(string); !ok || name != "test" {
		t.Errorf("expected name='test', got %v", result["name"])
	}

	if count, ok := result["count"].(float64); !ok || count != 42 {
		t.Errorf("expected count=42, got %v", result["count"])
	}
}

func TestReadJSONMap_EmptyString(t *testing.T) {
	result, err := ReadJSONMap("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for empty input, got %v", result)
	}
}

func TestReadJSONMap_InvalidJSON(t *testing.T) {
	input := `{invalid json}`

	_, err := ReadJSONMap(input, "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestReadJSONMap_FromFile(t *testing.T) {
	// Create a temp file with JSON content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	content := `{"key": "value", "number": 123}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := ReadJSONMap("", tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if key, ok := result["key"].(string); !ok || key != "value" {
		t.Errorf("expected key='value', got %v", result["key"])
	}

	if number, ok := result["number"].(float64); !ok || number != 123 {
		t.Errorf("expected number=123, got %v", result["number"])
	}
}

func TestReadJSONMap_NotAnObject(t *testing.T) {
	// JSON array instead of object
	input := `[1, 2, 3]`

	_, err := ReadJSONMap(input, "")
	if err == nil {
		t.Fatal("expected error for non-object JSON")
	}
}

func TestReadJSONMap_FileNotFound(t *testing.T) {
	_, err := ReadJSONMap("", "/nonexistent/path/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		cursor     string
		filter     string
		filterFile string
		params     []string
		wantKeys   map[string]string
		wantErr    bool
	}{
		{
			name:     "empty params",
			limit:    0,
			cursor:   "",
			filter:   "",
			params:   nil,
			wantKeys: map[string]string{},
		},
		{
			name:   "limit only",
			limit:  10,
			cursor: "",
			filter: "",
			params: nil,
			wantKeys: map[string]string{
				"limit": "10",
			},
		},
		{
			name:   "cursor only",
			limit:  0,
			cursor: "abc123",
			filter: "",
			params: nil,
			wantKeys: map[string]string{
				"starting_after": "abc123",
			},
		},
		{
			name:   "limit and cursor",
			limit:  25,
			cursor: "xyz789",
			filter: "",
			params: nil,
			wantKeys: map[string]string{
				"limit":          "25",
				"starting_after": "xyz789",
			},
		},
		{
			name:   "with filter",
			limit:  0,
			cursor: "",
			filter: `{"status": "active"}`,
			params: nil,
			wantKeys: map[string]string{
				"filter": `{"status":"active"}`,
			},
		},
		{
			name:   "with custom params",
			limit:  0,
			cursor: "",
			filter: "",
			params: []string{"orderBy=name", "direction=asc"},
			wantKeys: map[string]string{
				"orderBy":   "name",
				"direction": "asc",
			},
		},
		{
			name:   "all options combined",
			limit:  50,
			cursor: "cursor123",
			filter: `{"type": "test"}`,
			params: []string{"extra=value"},
			wantKeys: map[string]string{
				"limit":          "50",
				"starting_after": "cursor123",
				"filter":         `{"type":"test"}`,
				"extra":          "value",
			},
		},
		{
			name:   "param with empty value",
			limit:  0,
			cursor: "",
			filter: "",
			params: []string{"key="},
			wantKeys: map[string]string{
				"key": "",
			},
		},
		{
			name:   "param with equals in value",
			limit:  0,
			cursor: "",
			filter: "",
			params: []string{"equation=a=b"},
			wantKeys: map[string]string{
				"equation": "a=b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildQueryParams(tt.limit, tt.cursor, tt.filter, tt.filterFile, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildQueryParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			for key, want := range tt.wantKeys {
				got := result.Get(key)
				if got != want {
					t.Errorf("key %q: got %q, want %q", key, got, want)
				}
			}

			// Check no extra keys
			for key := range result {
				if _, ok := tt.wantKeys[key]; !ok {
					t.Errorf("unexpected key %q in result", key)
				}
			}
		})
	}
}

func TestBuildQueryParams_InvalidParam(t *testing.T) {
	// Param without = should return error
	_, err := BuildQueryParams(0, "", "", "", []string{"invalidparam"})
	if err == nil {
		t.Fatal("expected error for param without =")
	}
}

func TestBuildQueryParams_InvalidFilter(t *testing.T) {
	// Invalid JSON filter should return error
	_, err := BuildQueryParams(0, "", "{invalid}", "", nil)
	if err == nil {
		t.Fatal("expected error for invalid filter JSON")
	}
}

func TestBuildQueryParams_FilterFromFile(t *testing.T) {
	// Create a temp file with filter JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "filter.json")

	content := `{"status": "pending"}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := BuildQueryParams(0, "", "", tmpFile, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := result.Get("filter")
	want := `{"status":"pending"}`
	if got != want {
		t.Errorf("filter: got %q, want %q", got, want)
	}
}

func TestNewRESTClient_Default(t *testing.T) {
	// Reset viper to ensure clean state
	viper.Reset()

	client := NewRESTClient("https://api.example.com", "test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewRESTClient_WithDebug(t *testing.T) {
	viper.Reset()
	viper.Set("debug", true)
	t.Cleanup(viper.Reset)

	client := NewRESTClient("https://api.example.com", "test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewRESTClient_WithNoRetry(t *testing.T) {
	viper.Reset()
	viper.Set("no_retry", true)
	t.Cleanup(viper.Reset)

	client := NewRESTClient("https://api.example.com", "test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewRESTClient_WithAllFlags(t *testing.T) {
	viper.Reset()
	viper.Set("debug", true)
	viper.Set("no_retry", true)
	t.Cleanup(viper.Reset)

	client := NewRESTClient("https://api.example.com", "test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestReadJSONInput_FromStdin(t *testing.T) {
	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write JSON to the pipe
	jsonData := `{"from": "stdin", "value": 99}`
	go func() {
		w.WriteString(jsonData)
		w.Close()
	}()

	result, err := ReadJSONInput("", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if obj["from"] != "stdin" {
		t.Errorf("expected from='stdin', got %v", obj["from"])
	}
	if obj["value"] != float64(99) {
		t.Errorf("expected value=99, got %v", obj["value"])
	}
}

func TestReadJSONMap_FromStdin(t *testing.T) {
	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write JSON to the pipe
	jsonData := `{"key": "stdin_value"}`
	go func() {
		w.WriteString(jsonData)
		w.Close()
	}()

	result, err := ReadJSONMap("", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["key"] != "stdin_value" {
		t.Errorf("expected key='stdin_value', got %v", result["key"])
	}
}
