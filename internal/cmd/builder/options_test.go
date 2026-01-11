package builder

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestListOptions_RegisterFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	opts := &ListOptions{}

	opts.RegisterFlags(cmd)

	// Verify flags are registered
	flags := []string{"limit", "cursor", "all", "filter", "filter-file", "param", "sort", "order", "fields", "include"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag %q not registered", flag)
		}
	}
}

func TestListOptions_ToRESTOptions(t *testing.T) {
	opts := &ListOptions{
		Limit:   50,
		Cursor:  "abc123",
		Sort:    "createdAt",
		Order:   "desc",
		Fields:  "id,name",
		Include: "company,notes",
	}

	restOpts, err := opts.ToRESTOptions()
	if err != nil {
		t.Fatalf("ToRESTOptions() error = %v", err)
	}

	if restOpts.Limit != 50 {
		t.Errorf("Limit = %d, want 50", restOpts.Limit)
	}
	if restOpts.Cursor != "abc123" {
		t.Errorf("Cursor = %q, want %q", restOpts.Cursor, "abc123")
	}
	if len(restOpts.Include) != 2 {
		t.Errorf("Include length = %d, want 2", len(restOpts.Include))
	}
}

func TestListOptions_ToRESTOptions_WithFilter(t *testing.T) {
	opts := &ListOptions{
		Limit:  20,
		Filter: `{"email":{"like":"%@example.com"}}`,
	}

	restOpts, err := opts.ToRESTOptions()
	if err != nil {
		t.Fatalf("ToRESTOptions() error = %v", err)
	}

	if restOpts.Filter == nil {
		t.Fatal("Filter should not be nil")
	}

	email, ok := restOpts.Filter["email"]
	if !ok {
		t.Fatal("Filter should contain 'email' key")
	}

	emailMap, ok := email.(map[string]interface{})
	if !ok {
		t.Fatal("email filter should be a map")
	}

	like, ok := emailMap["like"]
	if !ok {
		t.Fatal("email filter should contain 'like' key")
	}

	if like != "%@example.com" {
		t.Errorf("like = %q, want %q", like, "%@example.com")
	}
}

func TestListOptions_ToRESTOptions_InvalidFilter(t *testing.T) {
	opts := &ListOptions{
		Limit:  20,
		Filter: `{invalid json}`,
	}

	_, err := opts.ToRESTOptions()
	if err == nil {
		t.Fatal("expected error for invalid JSON filter")
	}
}

func TestListOptions_ToRESTOptions_Include(t *testing.T) {
	tests := []struct {
		name     string
		include  string
		expected []string
	}{
		{
			name:     "single include",
			include:  "company",
			expected: []string{"company"},
		},
		{
			name:     "multiple includes",
			include:  "company,notes,tasks",
			expected: []string{"company", "notes", "tasks"},
		},
		{
			name:     "empty include",
			include:  "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ListOptions{
				Limit:   20,
				Include: tt.include,
			}

			restOpts, err := opts.ToRESTOptions()
			if err != nil {
				t.Fatalf("ToRESTOptions() error = %v", err)
			}

			if len(restOpts.Include) != len(tt.expected) {
				t.Errorf("Include length = %d, want %d", len(restOpts.Include), len(tt.expected))
				return
			}

			for i, v := range tt.expected {
				if restOpts.Include[i] != v {
					t.Errorf("Include[%d] = %q, want %q", i, restOpts.Include[i], v)
				}
			}
		})
	}
}

func TestListOptions_ToRESTOptions_Params(t *testing.T) {
	opts := &ListOptions{
		Limit:  20,
		Params: []string{"depth=1", "custom_key=custom_value"},
	}

	restOpts, err := opts.ToRESTOptions()
	if err != nil {
		t.Fatalf("ToRESTOptions() error = %v", err)
	}

	if restOpts.Params == nil {
		t.Fatal("Params should not be nil")
	}

	if restOpts.Params.Get("depth") != "1" {
		t.Errorf("Params[depth] = %q, want %q", restOpts.Params.Get("depth"), "1")
	}

	if restOpts.Params.Get("custom_key") != "custom_value" {
		t.Errorf("Params[custom_key] = %q, want %q", restOpts.Params.Get("custom_key"), "custom_value")
	}
}

func TestListOptions_ToRESTOptions_InvalidParams(t *testing.T) {
	opts := &ListOptions{
		Limit:  20,
		Params: []string{"invalid_param_without_equals"},
	}

	_, err := opts.ToRESTOptions()
	if err == nil {
		t.Fatal("expected error for invalid param format")
	}
}
