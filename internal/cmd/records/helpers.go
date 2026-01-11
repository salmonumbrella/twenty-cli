package records

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func resolveObject(ctx context.Context, client *rest.Client, name string, skip bool) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || skip {
		return name, nil
	}

	objects, err := client.ListObjects(ctx)
	if err != nil {
		// Best effort: fall back to provided name.
		return name, nil
	}

	for _, obj := range objects {
		if strings.EqualFold(name, obj.NamePlural) || strings.EqualFold(name, obj.NameSingular) {
			return obj.NamePlural, nil
		}
	}

	return name, nil
}

func buildPath(plural, suffix string) string {
	if suffix != "" && !strings.HasPrefix(suffix, "/") {
		suffix = "/" + suffix
	}
	return "/rest/" + plural + suffix
}

func parseQueryParams(limit int, cursor, filter, filterFile, sort, order, fields, include string, params []string) (url.Values, error) {
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		values.Set("starting_after", cursor)
	}
	if sort != "" {
		values.Set("order_by", sort)
	}
	if order != "" {
		values.Set("order_by_direction", order)
	}
	if fields != "" {
		values.Set("fields", fields)
	}
	// Twenty API uses 'depth' to fetch relations (0=none, 1=include)
	// When --include is specified, set depth=1 to fetch all relations
	if include != "" {
		values.Set("depth", "1")
	}

	if filter != "" || filterFile != "" {
		raw, err := shared.ReadJSONInput(filter, filterFile)
		if err != nil {
			return nil, err
		}
		if raw != nil {
			encoded, err := json.Marshal(raw)
			if err != nil {
				return nil, fmt.Errorf("encode filter: %w", err)
			}
			values.Set("filter", string(encoded))
		}
	}

	for _, p := range params {
		key, val, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid param %q (expected key=value)", p)
		}
		values.Add(key, val)
	}

	return values, nil
}

func parseBody(data, file string, sets []string) (map[string]interface{}, error) {
	payload, err := shared.ReadJSONInput(data, file)
	if err != nil {
		return nil, err
	}

	var body map[string]interface{}
	if payload != nil {
		var ok bool
		body, ok = payload.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("payload must be a JSON object")
		}
	} else {
		body = map[string]interface{}{}
	}

	for _, set := range sets {
		if err := applySet(body, set); err != nil {
			return nil, err
		}
	}

	if payload == nil && len(sets) == 0 {
		return nil, fmt.Errorf("missing JSON payload; use --data, --file, or --set")
	}

	return body, nil
}

func applySet(target map[string]interface{}, expr string) error {
	path, value, ok := strings.Cut(expr, "=")
	if !ok {
		return fmt.Errorf("invalid set expression %q (expected key=value)", expr)
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("invalid set expression %q (empty key)", expr)
	}

	parts := strings.Split(path, ".")
	current := target
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid set expression %q (empty path segment)", expr)
		}
		if i == len(parts)-1 {
			current[part] = parseJSONValue(value)
			return nil
		}

		next, ok := current[part]
		if !ok {
			child := map[string]interface{}{}
			current[part] = child
			current = child
			continue
		}

		child, ok := next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("set path %q conflicts with non-object value", path)
		}
		current = child
	}

	return nil
}

func parseJSONValue(raw string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var out interface{}
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out
	}
	return raw
}

func extractList(raw json.RawMessage, plural string) ([]interface{}, *types.PageInfo, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, nil, err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("response missing data field")
	}

	var items []interface{}
	if val, ok := data[plural]; ok {
		if arr, ok := val.([]interface{}); ok {
			items = arr
		}
	}

	if items == nil {
		// Fallback: use first array in data.
		for _, v := range data {
			if arr, ok := v.([]interface{}); ok {
				items = arr
				break
			}
		}
	}

	if items == nil {
		return nil, nil, fmt.Errorf("response did not contain list data for %s", plural)
	}

	var pageInfo *types.PageInfo
	if pi, ok := resp["pageInfo"].(map[string]interface{}); ok {
		pageInfo = &types.PageInfo{}
		if v, ok := pi["hasNextPage"].(bool); ok {
			pageInfo.HasNextPage = v
		}
		if v, ok := pi["endCursor"].(string); ok {
			pageInfo.EndCursor = v
		}
	}

	return items, pageInfo, nil
}
