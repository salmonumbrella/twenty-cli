package rest

import (
	"net/url"
	"strings"
	"testing"
)

func TestApplyListParams(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		opts           *ListOptions
		expectedPath   string
		expectedParams map[string]string
	}{
		{
			name:         "nil options",
			path:         "/rest/people",
			opts:         nil,
			expectedPath: "/rest/people",
		},
		{
			name:         "empty options",
			path:         "/rest/people",
			opts:         &ListOptions{},
			expectedPath: "/rest/people",
		},
		{
			name: "with limit",
			path: "/rest/people",
			opts: &ListOptions{Limit: 25},
			expectedParams: map[string]string{
				"limit": "25",
			},
		},
		{
			name: "with cursor",
			path: "/rest/people",
			opts: &ListOptions{Cursor: "abc123"},
			expectedParams: map[string]string{
				"starting_after": "abc123",
			},
		},
		{
			name: "with sort",
			path: "/rest/people",
			opts: &ListOptions{Sort: "createdAt"},
			expectedParams: map[string]string{
				"order_by": "createdAt",
			},
		},
		{
			name: "with order",
			path: "/rest/people",
			opts: &ListOptions{Order: "desc"},
			expectedParams: map[string]string{
				"order_by_direction": "desc",
			},
		},
		{
			name: "with fields",
			path: "/rest/people",
			opts: &ListOptions{Fields: "id,name,email"},
			expectedParams: map[string]string{
				"fields": "id%2Cname%2Cemail",
			},
		},
		{
			name: "with include (sets depth)",
			path: "/rest/people",
			opts: &ListOptions{Include: []string{"company"}},
			expectedParams: map[string]string{
				"depth": "1",
			},
		},
		{
			name: "with explicit depth",
			path: "/rest/people",
			opts: &ListOptions{Depth: 2},
			expectedParams: map[string]string{
				"depth": "2",
			},
		},
		{
			name: "include takes precedence over depth",
			path: "/rest/people",
			opts: &ListOptions{Include: []string{"company"}, Depth: 5},
			expectedParams: map[string]string{
				"depth": "1",
			},
		},
		{
			name: "with custom params",
			path: "/rest/people",
			opts: &ListOptions{
				Params: url.Values{
					"custom":  []string{"value"},
					"another": []string{"val1", "val2"},
				},
			},
			expectedParams: map[string]string{
				"custom": "value",
			},
		},
		{
			name: "all options combined",
			path: "/rest/people",
			opts: &ListOptions{
				Limit:   50,
				Cursor:  "xyz789",
				Sort:    "name",
				Order:   "asc",
				Fields:  "id,name",
				Include: []string{"company"},
			},
			expectedParams: map[string]string{
				"limit":              "50",
				"starting_after":     "xyz789",
				"order_by":           "name",
				"order_by_direction": "asc",
				"fields":             "id%2Cname",
				"depth":              "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applyListParams(tt.path, tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectedPath != "" && result != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, result)
			}

			if tt.expectedParams != nil {
				for param, expected := range tt.expectedParams {
					if !strings.Contains(result, param+"="+expected) {
						t.Errorf("expected query to contain %s=%s, got %s", param, expected, result)
					}
				}
			}
		})
	}
}

func TestBuildFilterString(t *testing.T) {
	tests := []struct {
		name     string
		filter   map[string]interface{}
		expected string
	}{
		{
			name: "simple string filter with implicit eq",
			filter: map[string]interface{}{
				"name": "John",
			},
			expected: `name[eq]:"John"`,
		},
		{
			name: "nested field filter",
			filter: map[string]interface{}{
				"emails": map[string]interface{}{
					"primaryEmail": map[string]interface{}{
						"ilike": "%test%",
					},
				},
			},
			expected: `emails.primaryEmail[ilike]:"%test%"`,
		},
		{
			name: "map of string filter",
			filter: map[string]interface{}{
				"status": map[string]string{
					"eq": "active",
				},
			},
			expected: `status[eq]:"active"`,
		},
		{
			name: "numeric filter",
			filter: map[string]interface{}{
				"age": 25,
			},
			expected: `age[eq]:25`,
		},
		{
			name: "boolean filter",
			filter: map[string]interface{}{
				"active": true,
			},
			expected: `active[eq]:true`,
		},
		{
			name: "deeply nested filter",
			filter: map[string]interface{}{
				"address": map[string]interface{}{
					"city": map[string]interface{}{
						"name": map[string]string{
							"eq": "San Francisco",
						},
					},
				},
			},
			expected: `address.city.name[eq]:"San Francisco"`,
		},
		{
			name: "OR filter with []interface{}",
			filter: map[string]interface{}{
				"or": []interface{}{
					map[string]interface{}{
						"name": map[string]interface{}{
							"firstName": map[string]interface{}{
								"ilike": "%dickinson%",
							},
						},
					},
					map[string]interface{}{
						"name": map[string]interface{}{
							"lastName": map[string]interface{}{
								"ilike": "%dickinson%",
							},
						},
					},
				},
			},
			expected: `or(name.firstName[ilike]:"%dickinson%",name.lastName[ilike]:"%dickinson%")`,
		},
		{
			name: "OR filter with []map[string]interface{}",
			filter: map[string]interface{}{
				"or": []map[string]interface{}{
					{
						"name": map[string]interface{}{
							"firstName": map[string]string{
								"ilike": "%smith%",
							},
						},
					},
					{
						"emails": map[string]interface{}{
							"primaryEmail": map[string]string{
								"ilike": "%smith%",
							},
						},
					},
				},
			},
			expected: `or(name.firstName[ilike]:"%smith%",emails.primaryEmail[ilike]:"%smith%")`,
		},
		{
			name: "NOT filter",
			filter: map[string]interface{}{
				"not": map[string]interface{}{
					"id": map[string]interface{}{
						"is": "NULL",
					},
				},
			},
			expected: `not(id[is]:"NULL")`,
		},
		{
			name: "simple operator filter",
			filter: map[string]interface{}{
				"createdAt": map[string]interface{}{
					"gte": "2023-01-01",
				},
			},
			expected: `createdAt[gte]:"2023-01-01"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFilterString(tt.filter)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestApplyListParams_WithFilter(t *testing.T) {
	opts := &ListOptions{
		Limit: 10,
		Filter: map[string]interface{}{
			"emails": map[string]interface{}{
				"primaryEmail": map[string]interface{}{
					"ilike": "%test%",
				},
			},
		},
	}

	result, err := applyListParams("/rest/people", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "limit=10") {
		t.Errorf("expected limit=10 in result, got %s", result)
	}

	// The filter should be a single query param with the string value
	// URL-encoded: filter=emails.primaryEmail%5Bilike%5D%3A%22%25test%25%22
	if !strings.Contains(result, "filter=") {
		t.Errorf("expected filter= in result, got %s", result)
	}

	// Decode and check the filter value
	parsedURL, err := url.Parse(result)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	filterValue := parsedURL.Query().Get("filter")
	expectedFilter := `emails.primaryEmail[ilike]:"%test%"`
	if filterValue != expectedFilter {
		t.Errorf("expected filter value %q, got %q", expectedFilter, filterValue)
	}
}

func TestIsOperator(t *testing.T) {
	operators := []string{
		"eq", "neq", "ne", "gt", "gte", "lt", "lte",
		"like", "ilike", "in", "is",
		"startsWith", "endsWith", "contains", "containsAny",
	}

	for _, op := range operators {
		if !isOperator(op) {
			t.Errorf("expected %q to be an operator", op)
		}
	}

	nonOperators := []string{"firstName", "lastName", "email", "name", "address"}
	for _, field := range nonOperators {
		if isOperator(field) {
			t.Errorf("expected %q to NOT be an operator", field)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", `"hello"`},
		{"%test%", `"%test%"`},
		{123, "123"},
		{true, "true"},
		{false, "false"},
		{3.14, "3.14"},
	}

	for _, tt := range tests {
		result := formatValue(tt.input)
		if result != tt.expected {
			t.Errorf("formatValue(%v): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}
