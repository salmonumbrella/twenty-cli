package records

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
)

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		plural   string
		suffix   string
		expected string
	}{
		{
			name:     "empty suffix",
			plural:   "people",
			suffix:   "",
			expected: "/rest/people",
		},
		{
			name:     "suffix with slash",
			plural:   "companies",
			suffix:   "/abc-123",
			expected: "/rest/companies/abc-123",
		},
		{
			name:     "suffix without slash",
			plural:   "tasks",
			suffix:   "task-id",
			expected: "/rest/tasks/task-id",
		},
		{
			name:     "complex suffix",
			plural:   "notes",
			suffix:   "note-id/restore",
			expected: "/rest/notes/note-id/restore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPath(tt.plural, tt.suffix)
			if result != tt.expected {
				t.Errorf("buildPath(%q, %q) = %q, want %q", tt.plural, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		cursor     string
		filter     string
		filterFile string
		sort       string
		order      string
		fields     string
		include    string
		params     []string
		wantErr    bool
		checkKey   string
		checkValue string
	}{
		{
			name:       "limit only",
			limit:      10,
			checkKey:   "limit",
			checkValue: "10",
		},
		{
			name:       "cursor only",
			cursor:     "abc123",
			checkKey:   "starting_after",
			checkValue: "abc123",
		},
		{
			name:       "sort only",
			sort:       "createdAt",
			checkKey:   "order_by",
			checkValue: "createdAt",
		},
		{
			name:       "order only",
			order:      "desc",
			checkKey:   "order_by_direction",
			checkValue: "desc",
		},
		{
			name:       "fields only",
			fields:     "id,name,email",
			checkKey:   "fields",
			checkValue: "id,name,email",
		},
		{
			name:       "include sets depth",
			include:    "company",
			checkKey:   "depth",
			checkValue: "1",
		},
		{
			name:       "filter JSON",
			filter:     `{"name":{"eq":"test"}}`,
			checkKey:   "filter",
			checkValue: `{"name":{"eq":"test"}}`,
		},
		{
			name:    "invalid param format",
			params:  []string{"invalid-param"},
			wantErr: true,
		},
		{
			name:       "valid param",
			params:     []string{"custom=value"},
			checkKey:   "custom",
			checkValue: "value",
		},
		{
			name:       "zero limit is not set",
			limit:      0,
			cursor:     "cursor123",
			checkKey:   "starting_after",
			checkValue: "cursor123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQueryParams(tt.limit, tt.cursor, tt.filter, tt.filterFile, tt.sort, tt.order, tt.fields, tt.include, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkKey != "" {
				val := result.Get(tt.checkKey)
				if val != tt.checkValue {
					t.Errorf("param %q = %q, want %q", tt.checkKey, val, tt.checkValue)
				}
			}
		})
	}
}

func TestParseQueryParams_MultipleParams(t *testing.T) {
	result, err := parseQueryParams(20, "cursor1", "", "", "name", "asc", "id,name", "company", []string{"extra=val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Get("limit") != "20" {
		t.Errorf("limit = %q, want %q", result.Get("limit"), "20")
	}
	if result.Get("starting_after") != "cursor1" {
		t.Errorf("starting_after = %q, want %q", result.Get("starting_after"), "cursor1")
	}
	if result.Get("order_by") != "name" {
		t.Errorf("order_by = %q, want %q", result.Get("order_by"), "name")
	}
	if result.Get("order_by_direction") != "asc" {
		t.Errorf("order_by_direction = %q, want %q", result.Get("order_by_direction"), "asc")
	}
	if result.Get("fields") != "id,name" {
		t.Errorf("fields = %q, want %q", result.Get("fields"), "id,name")
	}
	if result.Get("depth") != "1" {
		t.Errorf("depth = %q, want %q", result.Get("depth"), "1")
	}
	if result.Get("extra") != "val" {
		t.Errorf("extra = %q, want %q", result.Get("extra"), "val")
	}
}

func TestParseBody(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		file    string
		sets    []string
		wantErr bool
		check   func(map[string]interface{}) bool
	}{
		{
			name: "JSON data",
			data: `{"name":"test","value":123}`,
			check: func(m map[string]interface{}) bool {
				return m["name"] == "test" && m["value"].(float64) == 123
			},
		},
		{
			name: "set only",
			sets: []string{"name=test"},
			check: func(m map[string]interface{}) bool {
				return m["name"] == "test"
			},
		},
		{
			name: "nested set",
			sets: []string{"person.name=John"},
			check: func(m map[string]interface{}) bool {
				person, ok := m["person"].(map[string]interface{})
				if !ok {
					return false
				}
				return person["name"] == "John"
			},
		},
		{
			name: "JSON data with set override",
			data: `{"name":"original"}`,
			sets: []string{"name=override"},
			check: func(m map[string]interface{}) bool {
				return m["name"] == "override"
			},
		},
		{
			name:    "no data or sets",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "invalid set format",
			sets:    []string{"no-equals-sign"},
			wantErr: true,
		},
		{
			name:    "empty set key",
			sets:    []string{"=value"},
			wantErr: true,
		},
		{
			name:    "set with empty path segment",
			sets:    []string{"a..b=value"},
			wantErr: true,
		},
		{
			name:    "JSON array instead of object",
			data:    `[1,2,3]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBody(tt.data, tt.file, tt.sets)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.check != nil && !tt.check(result) {
				t.Errorf("check failed for result: %v", result)
			}
		})
	}
}

func TestApplySet(t *testing.T) {
	tests := []struct {
		name    string
		target  map[string]interface{}
		expr    string
		wantErr bool
		check   func(map[string]interface{}) bool
	}{
		{
			name:   "simple key value",
			target: map[string]interface{}{},
			expr:   "name=John",
			check: func(m map[string]interface{}) bool {
				return m["name"] == "John"
			},
		},
		{
			name:   "nested key value",
			target: map[string]interface{}{},
			expr:   "a.b.c=value",
			check: func(m map[string]interface{}) bool {
				a := m["a"].(map[string]interface{})
				b := a["b"].(map[string]interface{})
				return b["c"] == "value"
			},
		},
		{
			name: "existing nested object",
			target: map[string]interface{}{
				"person": map[string]interface{}{
					"first": "John",
				},
			},
			expr: "person.last=Doe",
			check: func(m map[string]interface{}) bool {
				person := m["person"].(map[string]interface{})
				return person["first"] == "John" && person["last"] == "Doe"
			},
		},
		{
			name:   "JSON value - number",
			target: map[string]interface{}{},
			expr:   "count=42",
			check: func(m map[string]interface{}) bool {
				return m["count"].(float64) == 42
			},
		},
		{
			name:   "JSON value - boolean true",
			target: map[string]interface{}{},
			expr:   "active=true",
			check: func(m map[string]interface{}) bool {
				return m["active"].(bool) == true
			},
		},
		{
			name:   "JSON value - boolean false",
			target: map[string]interface{}{},
			expr:   "active=false",
			check: func(m map[string]interface{}) bool {
				return m["active"].(bool) == false
			},
		},
		{
			name:   "JSON value - null",
			target: map[string]interface{}{},
			expr:   "value=null",
			check: func(m map[string]interface{}) bool {
				return m["value"] == nil
			},
		},
		{
			name:   "JSON value - array",
			target: map[string]interface{}{},
			expr:   `tags=["a","b"]`,
			check: func(m map[string]interface{}) bool {
				tags, ok := m["tags"].([]interface{})
				return ok && len(tags) == 2
			},
		},
		{
			name:    "no equals sign",
			target:  map[string]interface{}{},
			expr:    "invalid",
			wantErr: true,
		},
		{
			name:    "empty key",
			target:  map[string]interface{}{},
			expr:    "=value",
			wantErr: true,
		},
		{
			name:    "empty path segment",
			target:  map[string]interface{}{},
			expr:    "a..c=value",
			wantErr: true,
		},
		{
			name: "path conflicts with non-object",
			target: map[string]interface{}{
				"a": "string value",
			},
			expr:    "a.b=value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applySet(tt.target, tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.check != nil && !tt.check(tt.target) {
				t.Errorf("check failed for result: %v", tt.target)
			}
		})
	}
}

func TestParseJSONValue(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected interface{}
	}{
		{
			name:     "string",
			raw:      "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			raw:      "",
			expected: "",
		},
		{
			name:     "whitespace",
			raw:      "   ",
			expected: "",
		},
		{
			name:     "number",
			raw:      "42",
			expected: float64(42),
		},
		{
			name:     "float",
			raw:      "3.14",
			expected: 3.14,
		},
		{
			name:     "true",
			raw:      "true",
			expected: true,
		},
		{
			name:     "false",
			raw:      "false",
			expected: false,
		},
		{
			name:     "null",
			raw:      "null",
			expected: nil,
		},
		{
			name: "quoted string",
			raw:  `"quoted"`,
			// JSON unmarshals "quoted" -> quoted (string)
			expected: "quoted",
		},
		{
			name:     "invalid JSON treated as string",
			raw:      "not{valid}json",
			expected: "not{valid}json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseJSONValue(tt.raw)
			if result != tt.expected {
				t.Errorf("parseJSONValue(%q) = %v (%T), want %v (%T)", tt.raw, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestExtractList(t *testing.T) {
	tests := []struct {
		name        string
		raw         json.RawMessage
		plural      string
		wantErr     bool
		wantItems   int
		wantPageEnd string
	}{
		{
			name:        "valid response with exact plural",
			raw:         json.RawMessage(`{"data":{"people":[{"id":"1"},{"id":"2"}]},"pageInfo":{"hasNextPage":true,"endCursor":"cursor1"}}`),
			plural:      "people",
			wantItems:   2,
			wantPageEnd: "cursor1",
		},
		{
			name:      "valid response fallback to first array",
			raw:       json.RawMessage(`{"data":{"records":[{"id":"1"}]}}`),
			plural:    "other",
			wantItems: 1,
		},
		{
			name:    "missing data field",
			raw:     json.RawMessage(`{"result":{"items":[]}}`),
			plural:  "items",
			wantErr: true,
		},
		{
			name:    "no arrays in data",
			raw:     json.RawMessage(`{"data":{"scalar":"value"}}`),
			plural:  "items",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			raw:     json.RawMessage(`{invalid}`),
			plural:  "items",
			wantErr: true,
		},
		{
			name:      "no pageInfo",
			raw:       json.RawMessage(`{"data":{"items":[{"id":"1"}]}}`),
			plural:    "items",
			wantItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, pageInfo, err := extractList(tt.raw, tt.plural)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(items) != tt.wantItems {
				t.Errorf("got %d items, want %d", len(items), tt.wantItems)
			}

			if tt.wantPageEnd != "" {
				if pageInfo == nil {
					t.Error("expected pageInfo but got nil")
				} else if pageInfo.EndCursor != tt.wantPageEnd {
					t.Errorf("endCursor = %q, want %q", pageInfo.EndCursor, tt.wantPageEnd)
				}
			}
		})
	}
}

func TestResolveObject(t *testing.T) {
	tests := []struct {
		name     string
		objects  []struct{ NamePlural, NameSingular string }
		input    string
		skip     bool
		expected string
	}{
		{
			name: "resolves plural name",
			objects: []struct{ NamePlural, NameSingular string }{
				{"people", "person"},
				{"companies", "company"},
			},
			input:    "person",
			expected: "people",
		},
		{
			name: "resolves singular name",
			objects: []struct{ NamePlural, NameSingular string }{
				{"people", "person"},
				{"companies", "company"},
			},
			input:    "company",
			expected: "companies",
		},
		{
			name: "case insensitive",
			objects: []struct{ NamePlural, NameSingular string }{
				{"people", "person"},
			},
			input:    "PERSON",
			expected: "people",
		},
		{
			name: "no match returns input",
			objects: []struct{ NamePlural, NameSingular string }{
				{"people", "person"},
			},
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "skip resolution",
			input:    "myobject",
			skip:     true,
			expected: "myobject",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace trimmed",
			input:    "  people  ",
			skip:     true,
			expected: "people",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns object list
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := struct {
					Data struct {
						Objects []struct {
							NamePlural   string `json:"namePlural"`
							NameSingular string `json:"nameSingular"`
						} `json:"objects"`
					} `json:"data"`
				}{}
				for _, obj := range tt.objects {
					resp.Data.Objects = append(resp.Data.Objects, struct {
						NamePlural   string `json:"namePlural"`
						NameSingular string `json:"nameSingular"`
					}{obj.NamePlural, obj.NameSingular})
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			})

			server := httptest.NewServer(handler)
			defer server.Close()

			client := rest.NewClient(server.URL, "test-token", false, rest.WithNoRetry())
			result, err := resolveObject(context.Background(), client, tt.input, tt.skip)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("resolveObject(%q, %v) = %q, want %q", tt.input, tt.skip, result, tt.expected)
			}
		})
	}
}

func TestResolveObject_APIError(t *testing.T) {
	// When API errors, resolveObject should return the input name as fallback
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false, rest.WithNoRetry())
	result, err := resolveObject(context.Background(), client, "myobject", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "myobject" {
		t.Errorf("resolveObject should return input on API error, got %q", result)
	}
}
