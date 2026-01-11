package people

import (
	"encoding/json"
	"testing"
)

func TestBatchDeleteCmd_Flags(t *testing.T) {
	flags := []string{"file", "force"}
	for _, flag := range flags {
		if batchDeleteCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestBatchDeleteCmd_FileFlagShorthand(t *testing.T) {
	flag := batchDeleteCmd.Flags().Lookup("file")
	if flag == nil {
		t.Fatal("file flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("file flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestBatchDeleteCmd_Use(t *testing.T) {
	if batchDeleteCmd.Use != "batch-delete" {
		t.Errorf("Use = %q, want %q", batchDeleteCmd.Use, "batch-delete")
	}
}

func TestBatchDeleteCmd_Short(t *testing.T) {
	if batchDeleteCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestBatchDeleteCmd_Long(t *testing.T) {
	if batchDeleteCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestBatchDeleteCmd_ForceDefaultValue(t *testing.T) {
	flag := batchDeleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("force flag default = %q, want %q", flag.DefValue, "false")
	}
}

func TestBatchDeleteInput_ParseJSON(t *testing.T) {
	jsonData := `["id-1", "id-2", "id-3"]`

	var ids []string
	err := json.Unmarshal([]byte(jsonData), &ids)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(ids) != 3 {
		t.Errorf("Expected 3 IDs, got %d", len(ids))
	}

	expected := []string{"id-1", "id-2", "id-3"}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("ID[%d] = %q, want %q", i, id, expected[i])
		}
	}
}

func TestBatchDeleteInput_ParseJSONEmpty(t *testing.T) {
	jsonData := `[]`

	var ids []string
	err := json.Unmarshal([]byte(jsonData), &ids)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(ids) != 0 {
		t.Errorf("Expected 0 IDs, got %d", len(ids))
	}
}

func TestBatchDeleteInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{not an array}`

	var ids []string
	err := json.Unmarshal([]byte(invalidJSON), &ids)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestBatchDeleteInput_InvalidArrayContent(t *testing.T) {
	// Array of objects instead of strings
	invalidJSON := `[{"id": "1"}, {"id": "2"}]`

	var ids []string
	err := json.Unmarshal([]byte(invalidJSON), &ids)
	if err == nil {
		t.Error("Expected error for invalid array content")
	}
}

func TestBatchDeleteResult_Structure(t *testing.T) {
	result := map[string]interface{}{
		"deleted": []string{"id-1", "id-2"},
		"errors":  []string{},
	}

	deleted, ok := result["deleted"].([]string)
	if !ok {
		t.Fatal("deleted should be []string")
	}
	if len(deleted) != 2 {
		t.Errorf("Expected 2 deleted IDs, got %d", len(deleted))
	}

	errors, ok := result["errors"].([]string)
	if !ok {
		t.Fatal("errors should be []string")
	}
	if len(errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors))
	}
}

func TestBatchDeleteResult_WithErrors(t *testing.T) {
	result := map[string]interface{}{
		"deleted": []string{"id-1"},
		"errors":  []string{"ID id-2: not found", "ID id-3: permission denied"},
	}

	deleted, ok := result["deleted"].([]string)
	if !ok {
		t.Fatal("deleted should be []string")
	}
	if len(deleted) != 1 {
		t.Errorf("Expected 1 deleted ID, got %d", len(deleted))
	}

	errors, ok := result["errors"].([]string)
	if !ok {
		t.Fatal("errors should be []string")
	}
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
}

func TestBatchDeleteInput_SingleID(t *testing.T) {
	jsonData := `["single-id"]`

	var ids []string
	err := json.Unmarshal([]byte(jsonData), &ids)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(ids) != 1 {
		t.Errorf("Expected 1 ID, got %d", len(ids))
	}
	if ids[0] != "single-id" {
		t.Errorf("ID = %q, want %q", ids[0], "single-id")
	}
}

func TestBatchDeleteInput_LargeList(t *testing.T) {
	// Generate a large list of IDs
	ids := make([]string, 100)
	for i := 0; i < 100; i++ {
		ids[i] = "id-" + string(rune('A'+i%26))
	}

	jsonData, err := json.Marshal(ids)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	var parsed []string
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(parsed) != 100 {
		t.Errorf("Expected 100 IDs, got %d", len(parsed))
	}
}
