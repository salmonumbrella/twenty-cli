package rest

import (
	"fmt"
	"net/url"
	"strings"
)

// ListOptions configures list endpoints.
type ListOptions struct {
	Limit   int
	Cursor  string
	Filter  map[string]interface{}
	Sort    string
	Order   string
	Fields  string
	Include []string
	Depth   int // 0 = no relations, 1 = include relations
	Params  url.Values
}

func applyListParams(path string, opts *ListOptions) (string, error) {
	if opts == nil {
		return path, nil
	}

	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Cursor != "" {
		params.Set("starting_after", opts.Cursor)
	}
	if opts.Sort != "" {
		params.Set("order_by", opts.Sort)
	}
	if opts.Order != "" {
		params.Set("order_by_direction", opts.Order)
	}
	if opts.Fields != "" {
		params.Set("fields", opts.Fields)
	}
	// Twenty API uses 'depth' to fetch relations (0=none, 1=include)
	// When --include is specified, set depth=1 to fetch all relations
	if len(opts.Include) > 0 {
		params.Set("depth", "1")
	} else if opts.Depth > 0 {
		params.Set("depth", fmt.Sprintf("%d", opts.Depth))
	}
	// Twenty REST API uses string filter format:
	// filter=field[op]:"value" or filter=or(field1[op]:"value1",field2[op]:"value2")
	if len(opts.Filter) > 0 {
		filterStr := buildFilterString(opts.Filter)
		if filterStr != "" {
			params.Set("filter", filterStr)
		}
	}
	for key, vals := range opts.Params {
		for _, val := range vals {
			params.Add(key, val)
		}
	}

	if len(params) > 0 {
		return path + "?" + params.Encode(), nil
	}
	return path, nil
}

// buildFilterString converts a filter map to the Twenty API string format.
// Examples:
//   - Simple: {"createdAt": {"gte": "2023-01-01"}} -> createdAt[gte]:"2023-01-01"
//   - Nested: {"emails": {"primaryEmail": {"eq": "x"}}} -> emails.primaryEmail[eq]:"x"
//   - OR: {"or": [...]} -> or(filter1,filter2,...)
//   - NOT: {"not": {"id": {"is": "NULL"}}} -> not(id[is]:"NULL")
func buildFilterString(filter map[string]interface{}) string {
	var parts []string
	for key, value := range filter {
		part := buildFilterPart(key, value)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ",")
}

// buildFilterPart handles a single key-value pair in the filter.
func buildFilterPart(key string, value interface{}) string {
	switch key {
	case "or", "and", "not":
		return buildLogicalOperator(key, value)
	default:
		return buildFieldFilter(key, value)
	}
}

// buildLogicalOperator handles or(), and(), not() constructs.
func buildLogicalOperator(op string, value interface{}) string {
	var parts []string

	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				part := buildFilterString(m)
				if part != "" {
					parts = append(parts, part)
				}
			}
		}
	case []map[string]interface{}:
		for _, item := range v {
			part := buildFilterString(item)
			if part != "" {
				parts = append(parts, part)
			}
		}
	case map[string]interface{}:
		// For "not", value is typically a single map
		part := buildFilterString(v)
		if part != "" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("%s(%s)", op, strings.Join(parts, ","))
}

// buildFieldFilter handles field filters, possibly with nested fields.
// Returns formats like: field[op]:"value" or field.subfield[op]:"value"
func buildFieldFilter(fieldPath string, value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		// Check if this is an operator map (keys are operators like "eq", "ilike", etc.)
		// or a nested field map (keys are field names)
		for nestedKey, nestedValue := range v {
			if isOperator(nestedKey) {
				// This is an operator: field[op]:"value"
				return fmt.Sprintf("%s[%s]:%s", fieldPath, nestedKey, formatValue(nestedValue))
			}
			// This is a nested field: recurse with dot notation
			newPath := fieldPath + "." + nestedKey
			return buildFieldFilter(newPath, nestedValue)
		}
	case map[string]string:
		// Direct operator map
		for op, val := range v {
			return fmt.Sprintf("%s[%s]:%s", fieldPath, op, formatValue(val))
		}
	case string:
		// Direct value without operator (implicit eq)
		return fmt.Sprintf("%s[eq]:%s", fieldPath, formatValue(v))
	default:
		// Other types (numbers, booleans, etc.)
		return fmt.Sprintf("%s[eq]:%s", fieldPath, formatValue(v))
	}
	return ""
}

// isOperator returns true if the key is a filter operator.
func isOperator(key string) bool {
	operators := map[string]bool{
		"eq":          true,
		"neq":         true,
		"ne":          true,
		"gt":          true,
		"gte":         true,
		"lt":          true,
		"lte":         true,
		"like":        true,
		"ilike":       true,
		"in":          true,
		"is":          true,
		"startsWith":  true,
		"endsWith":    true,
		"contains":    true,
		"containsAny": true,
	}
	return operators[key]
}

// formatValue formats a value for the filter string.
// Strings are wrapped in quotes, other types are converted to string.
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
