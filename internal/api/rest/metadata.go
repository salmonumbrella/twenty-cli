package rest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

// ListObjects returns all object metadata from the Twenty metadata API
func (c *Client) ListObjects(ctx context.Context) ([]types.ObjectMetadata, error) {
	path := "/rest/metadata/objects"

	// Twenty API returns {"data":{"objects":[...]}} format
	var apiResp types.ObjectsListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return apiResp.Data.Objects, nil
}

// GetObject returns metadata for a specific object type by its singular name or ID.
// If nameOrID looks like a UUID, it queries directly; otherwise it lists all objects
// and finds the one matching by nameSingular or namePlural.
func (c *Client) GetObject(ctx context.Context, nameOrID string) (*types.ObjectMetadata, error) {
	// Check if it looks like a UUID (simple heuristic: contains dashes and is 36 chars)
	if len(nameOrID) == 36 && nameOrID[8] == '-' && nameOrID[13] == '-' {
		path := fmt.Sprintf("/rest/metadata/objects/%s", nameOrID)
		var apiResp types.ObjectResponse
		if err := c.Get(ctx, path, &apiResp); err != nil {
			return nil, err
		}
		return &apiResp.Data.Object, nil
	}

	// Otherwise, look up by name from the list
	objects, err := c.ListObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects for lookup: %w", err)
	}

	for i := range objects {
		if objects[i].NameSingular == nameOrID || objects[i].NamePlural == nameOrID {
			// Found the object by name, now get full details by ID
			path := fmt.Sprintf("/rest/metadata/objects/%s", objects[i].ID)
			var apiResp types.ObjectResponse
			if err := c.Get(ctx, path, &apiResp); err != nil {
				return nil, err
			}
			return &apiResp.Data.Object, nil
		}
	}

	return nil, fmt.Errorf("object not found: %s", nameOrID)
}

// ListFields returns all field metadata from the Twenty metadata API
func (c *Client) ListFields(ctx context.Context) ([]types.FieldMetadata, error) {
	path := "/rest/metadata/fields"

	// Twenty API returns {"data":{"fields":[...]}} format
	var apiResp types.FieldsListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return apiResp.Data.Fields, nil
}

// GetField returns metadata for a specific field by its ID
func (c *Client) GetField(ctx context.Context, id string) (*types.FieldMetadata, error) {
	path := fmt.Sprintf("/rest/metadata/fields/%s", id)

	// Twenty API returns {"data":{"field":{...}}} format
	var apiResp types.FieldResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.Field, nil
}

// CreateObject creates a new custom object.
func (c *Client) CreateObject(ctx context.Context, input map[string]interface{}) (json.RawMessage, error) {
	return c.DoRaw(ctx, "POST", "/rest/metadata/objects", input)
}

// UpdateObject updates an object by ID or name.
func (c *Client) UpdateObject(ctx context.Context, id string, input map[string]interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/rest/metadata/objects/%s", id)
	return c.DoRaw(ctx, "PATCH", path, input)
}

// DeleteObject deletes an object by ID or name.
func (c *Client) DeleteObject(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/metadata/objects/%s", id)
	return c.Delete(ctx, path)
}

// CreateField creates a new field definition.
func (c *Client) CreateField(ctx context.Context, input map[string]interface{}) (json.RawMessage, error) {
	return c.DoRaw(ctx, "POST", "/rest/metadata/fields", input)
}

// UpdateField updates a field by ID.
func (c *Client) UpdateField(ctx context.Context, id string, input map[string]interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/rest/metadata/fields/%s", id)
	return c.DoRaw(ctx, "PATCH", path, input)
}

// DeleteField deletes a field by ID.
func (c *Client) DeleteField(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/metadata/fields/%s", id)
	return c.Delete(ctx, path)
}
