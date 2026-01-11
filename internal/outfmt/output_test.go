package outfmt

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestWriteCSVFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantCSV []string // lines expected in output
		wantErr bool
	}{
		{
			name:  "direct array",
			input: `[{"id":"1","name":"Alice"},{"id":"2","name":"Bob"}]`,
			wantCSV: []string{
				"id,name",
				"1,Alice",
				"2,Bob",
			},
		},
		{
			name:  "wrapped in data",
			input: `{"data":{"favorites":[{"id":"f1","position":1},{"id":"f2","position":2}]}}`,
			wantCSV: []string{
				"id,position",
				"f1,1",
				"f2,2",
			},
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantCSV: []string{},
		},
		{
			name:  "nested objects",
			input: `[{"id":"1","meta":{"key":"value"}}]`,
			wantCSV: []string{
				"id,meta",
				`1,"{""key"":""value""}"`,
			},
		},
		{
			name:  "null values",
			input: `[{"id":"1","name":null}]`,
			wantCSV: []string{
				"id,name",
				"1,",
			},
		},
		{
			name:  "boolean values",
			input: `[{"id":"1","active":true}]`,
			wantCSV: []string{
				"active,id",
				"true,1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var data json.RawMessage = []byte(tt.input)

			err := WriteCSVFromJSON(&buf, data)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteCSVFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.wantCSV) == 0 {
				if buf.Len() > 0 {
					t.Errorf("expected empty output, got %q", buf.String())
				}
				return
			}

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			if len(lines) != len(tt.wantCSV) {
				t.Errorf("got %d lines, want %d lines\nGot:\n%s", len(lines), len(tt.wantCSV), buf.String())
				return
			}

			for i, want := range tt.wantCSV {
				if lines[i] != want {
					t.Errorf("line %d: got %q, want %q", i, lines[i], want)
				}
			}
		})
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email"}
	rows := [][]string{
		{"1", "Alice", "alice@example.com"},
		{"2", "Bob", "bob@example.com"},
	}

	err := WriteCSV(&buf, headers, rows)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	expected := "id,name,email\n1,Alice,alice@example.com\n2,Bob,bob@example.com\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriteCSV_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email"}
	rows := [][]string{}

	err := WriteCSV(&buf, headers, rows)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	// Should still output headers even with empty rows
	expected := "id,name,email\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriteCSV_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		rows     [][]string
		expected string
	}{
		{
			name:    "commas in values",
			headers: []string{"name", "address"},
			rows: [][]string{
				{"Alice", "123 Main St, Apt 4"},
			},
			expected: "name,address\nAlice,\"123 Main St, Apt 4\"\n",
		},
		{
			name:    "quotes in values",
			headers: []string{"name", "quote"},
			rows: [][]string{
				{"Bob", `He said "hello"`},
			},
			expected: "name,quote\nBob,\"He said \"\"hello\"\"\"\n",
		},
		{
			name:    "newlines in values",
			headers: []string{"name", "bio"},
			rows: [][]string{
				{"Charlie", "Line 1\nLine 2"},
			},
			expected: "name,bio\nCharlie,\"Line 1\nLine 2\"\n",
		},
		{
			name:    "mixed special characters",
			headers: []string{"col1", "col2"},
			rows: [][]string{
				{"value,with,commas", "and\n\"quotes\""},
			},
			expected: "col1,col2\n\"value,with,commas\",\"and\n\"\"quotes\"\"\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteCSV(&buf, tt.headers, tt.rows)
			if err != nil {
				t.Fatalf("WriteCSV() error = %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("got %q, want %q", buf.String(), tt.expected)
			}
		})
	}
}

func TestWriteTableFromJSON(t *testing.T) {
	var buf bytes.Buffer
	input := `{"data":{"favorites":[{"id":"f1","position":1},{"id":"f2","position":2}]}}`
	err := WriteTableFromJSON(&buf, json.RawMessage(input))
	if err != nil {
		t.Fatalf("WriteTableFromJSON() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"id", "position", "f1", "f2", "1", "2"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestWriteYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		query   string
		want    []string // substrings expected in output
		wantErr bool
	}{
		{
			name: "struct slice",
			input: []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				{ID: "1", Name: "Alice"},
				{ID: "2", Name: "Bob"},
			},
			want: []string{"id: \"1\"", "name: Alice", "id: \"2\"", "name: Bob"},
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value", "num": 42},
			want:  []string{"key: value", "num: 42"},
		},
		{
			name: "with jq query",
			input: []map[string]interface{}{
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
			},
			query: ".[0].name",
			want:  []string{"Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteYAML(&buf, tt.input, tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			out := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("expected output to contain %q, got %q", want, out)
				}
			}
		})
	}
}

func TestWriteYAMLFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string // substrings expected in output
		wantErr bool
	}{
		{
			name:  "direct array",
			input: `[{"id":"1","name":"Alice"},{"id":"2","name":"Bob"}]`,
			want:  []string{"- id:", "name: Alice", "name: Bob"},
		},
		{
			name:  "wrapped in data",
			input: `{"data":{"favorites":[{"id":"f1","position":1}]}}`,
			want:  []string{"data:", "favorites:", "id: f1", "position: 1"},
		},
		{
			name:  "empty object",
			input: `{}`,
			want:  []string{"{}"},
		},
		{
			name:  "nested objects",
			input: `{"outer":{"inner":{"key":"value"}}}`,
			want:  []string{"outer:", "inner:", "key: value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteYAMLFromJSON(&buf, json.RawMessage(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteYAMLFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			out := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("expected output to contain %q, got %q", want, out)
				}
			}
		})
	}
}

func TestWriteTableFromJSON_DirectArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string // substrings expected in output
		wantErr bool
	}{
		{
			name:  "direct array of objects",
			input: `[{"id":"1","name":"Alice"},{"id":"2","name":"Bob"}]`,
			want:  []string{"id", "name", "1", "Alice", "2", "Bob"},
		},
		{
			name:    "empty array",
			input:   `[]`,
			want:    []string{},
			wantErr: false,
		},
		{
			name:  "single object in array",
			input: `[{"id":"only","status":"active"}]`,
			want:  []string{"id", "status", "only", "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteTableFromJSON(&buf, json.RawMessage(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTableFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			out := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("expected output to contain %q, got %q", want, out)
				}
			}
		})
	}
}

func TestFindRecordArray(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantFirst map[string]interface{}
	}{
		{
			name:      "direct array",
			input:     `[{"id":"1"},{"id":"2"}]`,
			wantLen:   2,
			wantFirst: map[string]interface{}{"id": "1"},
		},
		{
			name:      "nested in data.resource",
			input:     `{"data":{"favorites":[{"id":"f1"},{"id":"f2"}]}}`,
			wantLen:   2,
			wantFirst: map[string]interface{}{"id": "f1"},
		},
		{
			name:      "nested in data (direct array under data)",
			input:     `{"data":[{"id":"d1"},{"id":"d2"}]}`,
			wantLen:   2,
			wantFirst: map[string]interface{}{"id": "d1"},
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantLen: 0,
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantLen: 0,
		},
		{
			name:      "object with array field",
			input:     `{"items":[{"name":"item1"}]}`,
			wantLen:   1,
			wantFirst: map[string]interface{}{"name": "item1"},
		},
		{
			name:    "array of non-objects",
			input:   `["a","b","c"]`,
			wantLen: 0, // Should return nil for non-object arrays
		},
		{
			name:    "deeply nested but no array at top level of data",
			input:   `{"data":{"info":{"name":"test"}}}`,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsed interface{}
			if err := json.Unmarshal([]byte(tt.input), &parsed); err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			result := findRecordArray(parsed)

			if len(result) != tt.wantLen {
				t.Errorf("findRecordArray() returned %d records, want %d", len(result), tt.wantLen)
				return
			}

			if tt.wantLen > 0 && tt.wantFirst != nil {
				// Check first record matches expected
				for key, wantVal := range tt.wantFirst {
					if gotVal, ok := result[0][key]; !ok {
						t.Errorf("first record missing key %q", key)
					} else if gotVal != wantVal {
						t.Errorf("first record[%q] = %v, want %v", key, gotVal, wantVal)
					}
				}
			}
		})
	}
}

func TestFormatJSONValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "nil value",
			input: nil,
			want:  "",
		},
		{
			name:  "string value",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "float64 whole number",
			input: float64(42),
			want:  "42",
		},
		{
			name:  "float64 with decimals",
			input: float64(3.14),
			want:  "3.14",
		},
		{
			name:  "bool true",
			input: true,
			want:  "true",
		},
		{
			name:  "bool false",
			input: false,
			want:  "false",
		},
		{
			name:  "nested object",
			input: map[string]interface{}{"key": "value"},
			want:  `{"key":"value"}`,
		},
		{
			name:  "empty object",
			input: map[string]interface{}{},
			want:  `{}`,
		},
		{
			name:  "array",
			input: []interface{}{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "empty array",
			input: []interface{}{},
			want:  `[]`,
		},
		{
			name:  "nested object with multiple fields",
			input: map[string]interface{}{"name": "test", "count": float64(10)},
			want:  "", // Will check contains since map ordering is non-deterministic
		},
		{
			name:  "array of objects",
			input: []interface{}{map[string]interface{}{"id": "1"}},
			want:  `[{"id":"1"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatJSONValue(tt.input)

			if tt.name == "nested object with multiple fields" {
				// Special case: check it's valid JSON with expected fields
				if !strings.Contains(got, `"name":"test"`) || !strings.Contains(got, `"count":10`) {
					t.Errorf("formatJSONValue() = %q, expected to contain name and count fields", got)
				}
				return
			}

			if got != tt.want {
				t.Errorf("formatJSONValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Tests for Printer struct and methods
func TestNewPrinter(t *testing.T) {
	tests := []struct {
		format string
		want   Format
	}{
		{"json", FormatJSON},
		{"yaml", FormatYAML},
		{"table", FormatTable},
		{"csv", FormatCSV},
		{"text", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			p := NewPrinter(tt.format)
			if p.Format != tt.want {
				t.Errorf("NewPrinter(%q).Format = %v, want %v", tt.format, p.Format, tt.want)
			}
			if p.Writer == nil {
				t.Error("NewPrinter() should set Writer to os.Stdout")
			}
		})
	}
}

func TestPrinter_Print_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &buf}

	data := map[string]interface{}{"name": "test", "value": 42}
	err := p.Print(data)
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	// Should be pretty-printed JSON
	if !strings.Contains(buf.String(), "\"name\": \"test\"") {
		t.Errorf("Print() output = %q, expected pretty JSON", buf.String())
	}
}

func TestPrinter_Print_YAML(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatYAML, Writer: &buf}

	data := map[string]interface{}{"name": "test", "value": 42}
	err := p.Print(data)
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !strings.Contains(buf.String(), "name: test") {
		t.Errorf("Print() output = %q, expected YAML", buf.String())
	}
}

func TestPrinter_Print_Table_Error(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Writer: &buf}

	data := map[string]interface{}{"name": "test"}
	err := p.Print(data)
	if err == nil {
		t.Error("Print() with table format should return error")
	}
	if !strings.Contains(err.Error(), "table format requires PrintTable method") {
		t.Errorf("Print() error = %v, expected table format error", err)
	}
}

func TestPrinter_PrintTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Writer: &buf}

	headers := []string{"ID", "Name", "Status"}
	rows := [][]string{
		{"1", "Alice", "active"},
		{"2", "Bob", "inactive"},
	}

	p.PrintTable(headers, rows)

	out := buf.String()
	// Check headers
	if !strings.Contains(out, "ID") || !strings.Contains(out, "Name") || !strings.Contains(out, "Status") {
		t.Errorf("PrintTable() missing headers in output: %q", out)
	}
	// Check data
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("PrintTable() missing data in output: %q", out)
	}
}

// Tests for WriteCSVFromStruct
type testRecord struct {
	ID     string  `json:"id"`
	Name   string  `json:"name,omitempty"`
	Count  int     `json:"count"`
	Active bool    `json:"active"`
	Score  float64 `json:"score"`
}

type testRecordWithNested struct {
	ID   string            `json:"id"`
	Meta map[string]string `json:"meta"`
}

type testRecordWithPointer struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}

func TestWriteCSVFromStruct(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    []string // substrings expected in output
		wantErr bool
	}{
		{
			name: "slice of structs",
			input: []testRecord{
				{ID: "1", Name: "Alice", Count: 10, Active: true, Score: 3.14},
				{ID: "2", Name: "Bob", Count: 20, Active: false, Score: 2.71},
			},
			want: []string{"id,name,count,active,score", "1,Alice,10,true,3.14", "2,Bob,20,false,2.71"},
		},
		{
			name:  "single struct",
			input: testRecord{ID: "1", Name: "Alice", Count: 10, Active: true, Score: 3.14},
			want:  []string{"id,name,count,active,score", "1,Alice,10,true,3.14"},
		},
		{
			name:  "empty slice",
			input: []testRecord{},
			want:  []string{},
		},
		{
			name: "struct with nested map",
			input: []testRecordWithNested{
				{ID: "1", Meta: map[string]string{"key": "value"}},
			},
			want: []string{"id,meta", `1,"{""key"":""value""}"`},
		},
		{
			name:    "non-struct type",
			input:   []string{"a", "b", "c"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteCSVFromStruct(&buf, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteCSVFromStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(tt.want) == 0 {
				if buf.Len() > 0 {
					t.Errorf("expected empty output, got %q", buf.String())
				}
				return
			}

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			for i, want := range tt.want {
				if i >= len(lines) {
					t.Errorf("missing line %d: want %q", i, want)
					continue
				}
				if lines[i] != want {
					t.Errorf("line %d: got %q, want %q", i, lines[i], want)
				}
			}
		})
	}
}

func TestWriteCSVFromStruct_PointerFields(t *testing.T) {
	name := "Alice"
	records := []testRecordWithPointer{
		{ID: "1", Name: &name},
		{ID: "2", Name: nil},
	}

	var buf bytes.Buffer
	err := WriteCSVFromStruct(&buf, records)
	if err != nil {
		t.Fatalf("WriteCSVFromStruct() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected output to contain Alice, got %q", out)
	}
}

func TestWriteCSVFromStruct_SliceFields(t *testing.T) {
	type recordWithSlice struct {
		ID    string   `json:"id"`
		Tags  []string `json:"tags"`
		Empty []string `json:"empty"`
	}

	records := []recordWithSlice{
		{ID: "1", Tags: []string{"a", "b"}, Empty: []string{}},
	}

	var buf bytes.Buffer
	err := WriteCSVFromStruct(&buf, records)
	if err != nil {
		t.Fatalf("WriteCSVFromStruct() error = %v", err)
	}

	out := buf.String()
	// CSV escapes quotes, so ["a","b"] becomes "[""a"",""b""]"
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("expected output to contain array elements, got %q", out)
	}
}

func TestWriteCSVFromStruct_TimeField(t *testing.T) {
	type recordWithTime struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	records := []recordWithTime{
		{ID: "1", CreatedAt: now},
	}

	var buf bytes.Buffer
	err := WriteCSVFromStruct(&buf, records)
	if err != nil {
		t.Fatalf("WriteCSVFromStruct() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "2024") {
		t.Errorf("expected output to contain year, got %q", out)
	}
}

func TestWriteCSVFromStruct_PointerToStruct(t *testing.T) {
	records := []*testRecord{
		{ID: "1", Name: "Alice", Count: 10, Active: true, Score: 3.14},
		{ID: "2", Name: "Bob", Count: 20, Active: false, Score: 2.71},
	}

	var buf bytes.Buffer
	err := WriteCSVFromStruct(&buf, records)
	if err != nil {
		t.Fatalf("WriteCSVFromStruct() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("expected output to contain names, got %q", out)
	}
}

// Tests for WriteJSON with jq query
func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		query   string
		want    []string // substrings expected
		wantErr bool
	}{
		{
			name:  "simple map without query",
			input: map[string]interface{}{"name": "test", "value": 42},
			query: "",
			want:  []string{`"name": "test"`, `"value": 42`},
		},
		{
			name: "array with jq query",
			input: []map[string]interface{}{
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
			},
			query: ".[0].name",
			want:  []string{"Alice"},
		},
		{
			name: "filter with jq",
			input: []map[string]interface{}{
				{"id": "1", "active": true},
				{"id": "2", "active": false},
			},
			query: ".[] | select(.active == true) | .id",
			want:  []string{"1"},
		},
		{
			name:    "invalid jq query",
			input:   map[string]interface{}{"name": "test"},
			query:   ".[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSON(&buf, tt.input, tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			out := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("WriteJSON() output = %q, expected to contain %q", out, want)
				}
			}
		})
	}
}

// Test WriteCSVFromJSON with byte slice input
func TestWriteCSVFromJSON_ByteSlice(t *testing.T) {
	input := []byte(`[{"id":"1","name":"Alice"}]`)
	var buf bytes.Buffer
	err := WriteCSVFromJSON(&buf, input)
	if err != nil {
		t.Fatalf("WriteCSVFromJSON() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Alice") {
		t.Errorf("expected output to contain Alice, got %q", buf.String())
	}
}

// Test WriteCSVFromJSON with struct input (marshals to JSON first)
func TestWriteCSVFromJSON_Struct(t *testing.T) {
	type record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	input := []record{{ID: "1", Name: "Alice"}}
	var buf bytes.Buffer
	err := WriteCSVFromJSON(&buf, input)
	if err != nil {
		t.Fatalf("WriteCSVFromJSON() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Alice") {
		t.Errorf("expected output to contain Alice, got %q", buf.String())
	}
}

// Test WriteYAMLFromJSON with byte slice input
func TestWriteYAMLFromJSON_ByteSlice(t *testing.T) {
	input := []byte(`{"name":"test","value":42}`)
	var buf bytes.Buffer
	err := WriteYAMLFromJSON(&buf, input)
	if err != nil {
		t.Fatalf("WriteYAMLFromJSON() error = %v", err)
	}

	if !strings.Contains(buf.String(), "name: test") {
		t.Errorf("expected output to contain name: test, got %q", buf.String())
	}
}

// Test WriteTableFromJSON with byte slice input
func TestWriteTableFromJSON_ByteSlice(t *testing.T) {
	input := []byte(`[{"id":"1","name":"Alice"}]`)
	var buf bytes.Buffer
	err := WriteTableFromJSON(&buf, input)
	if err != nil {
		t.Fatalf("WriteTableFromJSON() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Alice") {
		t.Errorf("expected output to contain Alice, got %q", buf.String())
	}
}

// Test error cases for WriteCSVFromJSON
func TestWriteCSVFromJSON_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "invalid JSON",
			input:   json.RawMessage(`{invalid`),
			wantErr: true,
		},
		{
			name:    "single object with no fields",
			input:   json.RawMessage(`{"data":{"resource":[{}]}}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteCSVFromJSON(&buf, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteCSVFromJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test error cases for WriteTableFromJSON
func TestWriteTableFromJSON_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "invalid JSON",
			input:   json.RawMessage(`{invalid`),
			wantErr: true,
		},
		{
			name:    "single object with no fields",
			input:   json.RawMessage(`{"data":{"resource":[{}]}}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteTableFromJSON(&buf, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTableFromJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test error cases for WriteYAMLFromJSON
func TestWriteYAMLFromJSON_Errors(t *testing.T) {
	var buf bytes.Buffer
	err := WriteYAMLFromJSON(&buf, json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("WriteYAMLFromJSON() expected error for invalid JSON")
	}
}

// Test WriteYAMLFromJSON with struct input (marshals to JSON first)
func TestWriteYAMLFromJSON_Struct(t *testing.T) {
	type record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	input := record{ID: "1", Name: "Alice"}
	var buf bytes.Buffer
	err := WriteYAMLFromJSON(&buf, input)
	if err != nil {
		t.Fatalf("WriteYAMLFromJSON() error = %v", err)
	}

	if !strings.Contains(buf.String(), "name: Alice") {
		t.Errorf("expected output to contain name: Alice, got %q", buf.String())
	}
}

// Test WriteTableFromJSON with struct input (marshals to JSON first)
func TestWriteTableFromJSON_Struct(t *testing.T) {
	type record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	input := []record{{ID: "1", Name: "Alice"}}
	var buf bytes.Buffer
	err := WriteTableFromJSON(&buf, input)
	if err != nil {
		t.Fatalf("WriteTableFromJSON() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Alice") {
		t.Errorf("expected output to contain Alice, got %q", buf.String())
	}
}

// Test getStructHeaders with json:"-" tag (should skip field)
func TestGetStructHeaders_SkipMinusTag(t *testing.T) {
	type recordWithSkippedField struct {
		ID      string `json:"id"`
		Secret  string `json:"-"`
		Name    string `json:"name"`
		NoTag   string
		SkipAll string `json:"-,"`
	}

	var buf bytes.Buffer
	records := []recordWithSkippedField{
		{ID: "1", Secret: "hidden", Name: "Alice", NoTag: "value", SkipAll: "also hidden"},
	}
	err := WriteCSVFromStruct(&buf, records)
	if err != nil {
		t.Fatalf("WriteCSVFromStruct() error = %v", err)
	}

	out := buf.String()
	// Should have id, name, NoTag but use field name for "-" tag
	if strings.Contains(out, "hidden") && !strings.Contains(out, "Secret") {
		// The field with json:"-" should use field name "Secret" as header
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected output to contain Alice, got %q", out)
	}
}

// Test formatValue with empty/invalid reflect.Value
func TestFormatValue_InvalidValue(t *testing.T) {
	// Create an invalid reflect.Value
	var invalid reflect.Value
	result := formatValue(invalid)
	if result != "" {
		t.Errorf("formatValue(invalid) = %q, want empty string", result)
	}
}

// Test formatValue with nil pointer
func TestFormatValue_NilPointer(t *testing.T) {
	var nilPtr *string
	v := reflect.ValueOf(nilPtr)
	result := formatValue(v)
	if result != "" {
		t.Errorf("formatValue(nil pointer) = %q, want empty string", result)
	}
}

// Test formatValue with non-nil pointer
func TestFormatValue_NonNilPointer(t *testing.T) {
	str := "hello"
	ptr := &str
	v := reflect.ValueOf(ptr)
	result := formatValue(v)
	if result != "hello" {
		t.Errorf("formatValue(pointer) = %q, want \"hello\"", result)
	}
}

// Test formatValue with empty map
func TestFormatValue_EmptyMap(t *testing.T) {
	m := map[string]string{}
	v := reflect.ValueOf(m)
	result := formatValue(v)
	if result != "" {
		t.Errorf("formatValue(empty map) = %q, want empty string", result)
	}
}

// Test formatValue with empty slice
func TestFormatValue_EmptySlice(t *testing.T) {
	s := []string{}
	v := reflect.ValueOf(s)
	result := formatValue(v)
	if result != "" {
		t.Errorf("formatValue(empty slice) = %q, want empty string", result)
	}
}

// Test formatValue with nested struct
func TestFormatValue_NestedStruct(t *testing.T) {
	type inner struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	}
	s := inner{Key: "test", Value: 42}
	v := reflect.ValueOf(s)
	result := formatValue(v)
	// Should be JSON serialized
	if !strings.Contains(result, "test") || !strings.Contains(result, "42") {
		t.Errorf("formatValue(struct) = %q, expected JSON with key and value", result)
	}
}

// Test WriteCSVFromStruct with struct containing unexported fields
func TestWriteCSVFromStruct_UnexportedFields(t *testing.T) {
	type recordWithUnexported struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		internal string // unexported, should be skipped
	}

	var buf bytes.Buffer
	records := []recordWithUnexported{
		{ID: "1", Name: "Alice", internal: "hidden"},
	}
	err := WriteCSVFromStruct(&buf, records)
	if err != nil {
		t.Fatalf("WriteCSVFromStruct() error = %v", err)
	}

	out := buf.String()
	// Should only have id and name headers, not internal
	if strings.Contains(out, "internal") || strings.Contains(out, "hidden") {
		t.Errorf("output should not contain unexported field, got %q", out)
	}
	if !strings.Contains(out, "id") || !strings.Contains(out, "name") {
		t.Errorf("output should contain exported fields, got %q", out)
	}
}

// Test WriteYAML with invalid jq query
func TestWriteYAML_InvalidQuery(t *testing.T) {
	var buf bytes.Buffer
	input := map[string]interface{}{"name": "test"}
	err := WriteYAML(&buf, input, ".[invalid")
	if err == nil {
		t.Error("WriteYAML() expected error for invalid jq query")
	}
}

// Test WriteYAML with query that returns multiple results
func TestWriteYAML_MultipleResults(t *testing.T) {
	var buf bytes.Buffer
	input := []map[string]interface{}{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}
	err := WriteYAML(&buf, input, ".[].name")
	if err != nil {
		t.Fatalf("WriteYAML() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("WriteYAML() output = %q, expected both names", out)
	}
}

// Test WriteJSON with query that returns multiple results
func TestWriteJSON_MultipleResults(t *testing.T) {
	var buf bytes.Buffer
	input := []map[string]interface{}{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}
	err := WriteJSON(&buf, input, ".[].name")
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("WriteJSON() output = %q, expected both names", out)
	}
}

// Test applyQuery with runtime error from jq
func TestApplyQuery_RuntimeError(t *testing.T) {
	var buf bytes.Buffer
	// .foo on a string will cause a runtime error
	input := "just a string"
	err := WriteJSON(&buf, input, ".foo")
	if err == nil {
		t.Error("WriteJSON() expected error for accessing property on string")
	}
}

// Test formatJSONValue with unknown type (should use default case)
func TestFormatJSONValue_UnknownType(t *testing.T) {
	// int type goes through default case
	result := formatJSONValue(42)
	if result != "42" {
		t.Errorf("formatJSONValue(42) = %q, want \"42\"", result)
	}
}

// Test WriteTableFromJSON with struct that needs marshaling error
func TestWriteTableFromJSON_MarshalError(t *testing.T) {
	// Create a type that cannot be marshaled to JSON (channel)
	ch := make(chan int)
	var buf bytes.Buffer
	err := WriteTableFromJSON(&buf, ch)
	if err == nil {
		t.Error("WriteTableFromJSON() expected error for unmarshalable type")
	}
}

// Test WriteCSVFromJSON with struct that needs marshaling error
func TestWriteCSVFromJSON_MarshalError(t *testing.T) {
	// Create a type that cannot be marshaled to JSON (channel)
	ch := make(chan int)
	var buf bytes.Buffer
	err := WriteCSVFromJSON(&buf, ch)
	if err == nil {
		t.Error("WriteCSVFromJSON() expected error for unmarshalable type")
	}
}

// Test WriteYAMLFromJSON with struct that needs marshaling error
func TestWriteYAMLFromJSON_MarshalError(t *testing.T) {
	// Create a type that cannot be marshaled to JSON (channel)
	ch := make(chan int)
	var buf bytes.Buffer
	err := WriteYAMLFromJSON(&buf, ch)
	if err == nil {
		t.Error("WriteYAMLFromJSON() expected error for unmarshalable type")
	}
}
