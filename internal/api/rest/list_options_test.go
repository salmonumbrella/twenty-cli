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

func TestAddFilterParams(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		filter         map[string]interface{}
		expectedParams map[string]string
	}{
		{
			name:   "simple string filter",
			prefix: "filter",
			filter: map[string]interface{}{
				"name": "John",
			},
			expectedParams: map[string]string{
				"filter[name]": "John",
			},
		},
		{
			name:   "nested map filter",
			prefix: "filter",
			filter: map[string]interface{}{
				"emails": map[string]interface{}{
					"primaryEmail": map[string]interface{}{
						"ilike": "%test%",
					},
				},
			},
			expectedParams: map[string]string{
				"filter[emails][primaryEmail][ilike]": "%test%",
			},
		},
		{
			name:   "map of string filter",
			prefix: "filter",
			filter: map[string]interface{}{
				"status": map[string]string{
					"eq": "active",
				},
			},
			expectedParams: map[string]string{
				"filter[status][eq]": "active",
			},
		},
		{
			name:   "numeric filter",
			prefix: "filter",
			filter: map[string]interface{}{
				"age": 25,
			},
			expectedParams: map[string]string{
				"filter[age]": "25",
			},
		},
		{
			name:   "boolean filter",
			prefix: "filter",
			filter: map[string]interface{}{
				"active": true,
			},
			expectedParams: map[string]string{
				"filter[active]": "true",
			},
		},
		{
			name:   "multiple filters",
			prefix: "filter",
			filter: map[string]interface{}{
				"name":   "John",
				"status": "active",
			},
			expectedParams: map[string]string{
				"filter[name]":   "John",
				"filter[status]": "active",
			},
		},
		{
			name:   "deeply nested filter",
			prefix: "filter",
			filter: map[string]interface{}{
				"address": map[string]interface{}{
					"city": map[string]interface{}{
						"name": map[string]string{
							"eq": "San Francisco",
						},
					},
				},
			},
			expectedParams: map[string]string{
				"filter[address][city][name][eq]": "San Francisco",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := url.Values{}
			addFilterParams(params, tt.prefix, tt.filter)

			for expectedKey, expectedValue := range tt.expectedParams {
				actual := params.Get(expectedKey)
				if actual != expectedValue {
					t.Errorf("expected %s=%s, got %s=%s", expectedKey, expectedValue, expectedKey, actual)
				}
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

	// The filter should be URL encoded
	if !strings.Contains(result, "filter") {
		t.Errorf("expected filter in result, got %s", result)
	}
}
