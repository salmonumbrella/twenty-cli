package records

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// ============== Batch Create Tests ==============

func TestBatchCreateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if batchCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestBatchCreateCmd_Use(t *testing.T) {
	if batchCreateCmd.Use != "batch-create <object>" {
		t.Errorf("Use = %q, want %q", batchCreateCmd.Use, "batch-create <object>")
	}
}

func TestBatchCreateCmd_Short(t *testing.T) {
	if batchCreateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestBatchCreateCmd_Args(t *testing.T) {
	err := batchCreateCmd.Args(batchCreateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	err = batchCreateCmd.Args(batchCreateCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestRunBatchCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/batch/people" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload []interface{}
		json.Unmarshal(body, &payload)

		if len(payload) != 2 {
			t.Errorf("expected 2 items, got %d", len(payload))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[{"id":"p1"},{"id":"p2"}]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	batchCreateData = `[{"name":"John"},{"name":"Jane"}]`
	batchCreateDataFile = ""
	noResolve = true

	err := runBatchCreate(batchCreateCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runBatchCreate failed: %v", err)
	}

	batchCreateData = ""
}

func TestRunBatchCreate_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	batchCreateData = ""
	batchCreateDataFile = ""
	noResolve = true

	err := runBatchCreate(batchCreateCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

func TestRunBatchCreate_NonArrayPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	batchCreateData = `{"name":"not an array"}`
	batchCreateDataFile = ""
	noResolve = true

	err := runBatchCreate(batchCreateCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for non-array payload")
	}

	batchCreateData = ""
}

// ============== Batch Update Tests ==============

func TestBatchUpdateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if batchUpdateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestBatchUpdateCmd_Use(t *testing.T) {
	if batchUpdateCmd.Use != "batch-update <object>" {
		t.Errorf("Use = %q, want %q", batchUpdateCmd.Use, "batch-update <object>")
	}
}

func TestBatchUpdateCmd_Short(t *testing.T) {
	if batchUpdateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunBatchUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/batch/people" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"people":[{"id":"p1"},{"id":"p2"}]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	batchUpdateData = `[{"id":"p1","name":"Updated1"},{"id":"p2","name":"Updated2"}]`
	batchUpdateDataFile = ""
	noResolve = true

	err := runBatchUpdate(batchUpdateCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runBatchUpdate failed: %v", err)
	}

	batchUpdateData = ""
}

func TestRunBatchUpdate_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	batchUpdateData = ""
	batchUpdateDataFile = ""
	noResolve = true

	err := runBatchUpdate(batchUpdateCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

func TestRunBatchUpdate_NonArrayPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	batchUpdateData = `{"id":"p1"}`
	batchUpdateDataFile = ""
	noResolve = true

	err := runBatchUpdate(batchUpdateCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for non-array payload")
	}

	batchUpdateData = ""
}

// ============== Batch Delete Tests ==============

func TestBatchDeleteCmd_Flags(t *testing.T) {
	flags := []string{"data", "file", "ids", "force"}
	for _, flag := range flags {
		if batchDeleteCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestBatchDeleteCmd_Use(t *testing.T) {
	if batchDeleteCmd.Use != "batch-delete <object>" {
		t.Errorf("Use = %q, want %q", batchDeleteCmd.Use, "batch-delete <object>")
	}
}

func TestBatchDeleteCmd_Short(t *testing.T) {
	if batchDeleteCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunBatchDelete_WithoutForce(t *testing.T) {
	originalForce := batchDeleteForce
	defer func() { batchDeleteForce = originalForce }()

	batchDeleteForce = false

	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called without --force")
	}))
	defer cleanup()

	noResolve = true

	err := runBatchDelete(batchDeleteCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error when --force is not set, got %v", err)
	}
}

func TestRunBatchDelete_WithIDs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if !strings.HasPrefix(r.URL.Path, "/rest/batch/people") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		filter := r.URL.Query().Get("filter")
		if !strings.Contains(filter, "id[in]") {
			t.Errorf("expected filter with id[in], got %q", filter)
		}

		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"deletedCount":2}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := batchDeleteForce
	defer func() { batchDeleteForce = originalForce }()

	batchDeleteForce = true
	batchDeleteData = ""
	batchDeleteDataFile = ""
	batchDeleteIDs = "id1,id2"
	noResolve = true

	err := runBatchDelete(batchDeleteCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runBatchDelete failed: %v", err)
	}

	batchDeleteIDs = ""
}

func TestRunBatchDelete_WithJSONData(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"deletedCount":2}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := batchDeleteForce
	defer func() { batchDeleteForce = originalForce }()

	batchDeleteForce = true
	batchDeleteData = `["id1","id2"]`
	batchDeleteDataFile = ""
	batchDeleteIDs = ""
	noResolve = true

	err := runBatchDelete(batchDeleteCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runBatchDelete failed: %v", err)
	}

	batchDeleteData = ""
}

// ============== Batch Destroy Tests ==============

func TestBatchDestroyCmd_Flags(t *testing.T) {
	flags := []string{"data", "file", "ids", "force"}
	for _, flag := range flags {
		if batchDestroyCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestBatchDestroyCmd_Use(t *testing.T) {
	if batchDestroyCmd.Use != "batch-destroy <object>" {
		t.Errorf("Use = %q, want %q", batchDestroyCmd.Use, "batch-destroy <object>")
	}
}

func TestBatchDestroyCmd_Short(t *testing.T) {
	if batchDestroyCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunBatchDestroy_WithoutForce(t *testing.T) {
	originalForce := batchDestroyForce
	defer func() { batchDestroyForce = originalForce }()

	batchDestroyForce = false

	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called without --force")
	}))
	defer cleanup()

	noResolve = true

	err := runBatchDestroy(batchDestroyCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error when --force is not set, got %v", err)
	}
}

func TestRunBatchDestroy_WithIDs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/batch/people/destroy" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"destroyedCount":2}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := batchDestroyForce
	defer func() { batchDestroyForce = originalForce }()

	batchDestroyForce = true
	batchDestroyData = ""
	batchDestroyDataFile = ""
	batchDestroyIDs = "id1,id2"
	noResolve = true

	err := runBatchDestroy(batchDestroyCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runBatchDestroy failed: %v", err)
	}

	batchDestroyIDs = ""
}

// ============== Batch Restore Tests ==============

func TestBatchRestoreCmd_Flags(t *testing.T) {
	flags := []string{"data", "file", "ids"}
	for _, flag := range flags {
		if batchRestoreCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestBatchRestoreCmd_Use(t *testing.T) {
	if batchRestoreCmd.Use != "batch-restore <object>" {
		t.Errorf("Use = %q, want %q", batchRestoreCmd.Use, "batch-restore <object>")
	}
}

func TestBatchRestoreCmd_Short(t *testing.T) {
	if batchRestoreCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunBatchRestore_WithIDs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/batch/people/restore" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"restoredCount":2}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	batchRestoreData = ""
	batchRestoreDataFile = ""
	batchRestoreIDs = "id1,id2"
	noResolve = true

	err := runBatchRestore(batchRestoreCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runBatchRestore failed: %v", err)
	}

	batchRestoreIDs = ""
}

func TestRunBatchRestore_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	batchRestoreData = ""
	batchRestoreDataFile = ""
	batchRestoreIDs = ""
	noResolve = true

	err := runBatchRestore(batchRestoreCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

// ============== Helper Function Tests ==============

func TestReadBatchIDs_WithIDs(t *testing.T) {
	result, err := readBatchIDs("", "", "id1, id2, id3")
	if err != nil {
		t.Fatalf("readBatchIDs failed: %v", err)
	}

	arr, ok := result.([]string)
	if !ok {
		t.Fatal("expected []string result")
	}

	if len(arr) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(arr))
	}
}

func TestReadBatchIDs_WithData(t *testing.T) {
	result, err := readBatchIDs(`["id1","id2"]`, "", "")
	if err != nil {
		t.Fatalf("readBatchIDs failed: %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatal("expected []interface{} result")
	}

	if len(arr) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(arr))
	}
}

func TestReadBatchIDs_NoInput(t *testing.T) {
	_, err := readBatchIDs("", "", "")
	if err == nil {
		t.Error("expected error when no input provided")
	}
}

func TestReadBatchIDs_EmptyIDs(t *testing.T) {
	_, err := readBatchIDs("", "", "  ,  ,  ")
	if err == nil {
		t.Error("expected error for empty IDs")
	}
}

func TestReadBatchIDs_NonArrayPayload(t *testing.T) {
	_, err := readBatchIDs(`{"id":"not-array"}`, "", "")
	if err == nil {
		t.Error("expected error for non-array payload")
	}
}

func TestReadBatchIDStrings_WithIDs(t *testing.T) {
	result, err := readBatchIDStrings("", "", "id1, id2, id3")
	if err != nil {
		t.Fatalf("readBatchIDStrings failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(result))
	}
	if result[0] != "id1" {
		t.Errorf("result[0] = %q, want %q", result[0], "id1")
	}
}

func TestReadBatchIDStrings_WithData(t *testing.T) {
	result, err := readBatchIDStrings(`["id1","id2"]`, "", "")
	if err != nil {
		t.Fatalf("readBatchIDStrings failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(result))
	}
}

func TestReadBatchIDStrings_NoInput(t *testing.T) {
	_, err := readBatchIDStrings("", "", "")
	if err == nil {
		t.Error("expected error when no input provided")
	}
}

func TestReadBatchIDStrings_NonStringArray(t *testing.T) {
	_, err := readBatchIDStrings(`[1,2,3]`, "", "")
	if err == nil {
		t.Error("expected error for non-string array")
	}
}
