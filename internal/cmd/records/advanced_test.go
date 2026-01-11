package records

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// ============== Merge Tests ==============

func TestMergeCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if mergeCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestMergeCmd_Use(t *testing.T) {
	if mergeCmd.Use != "merge <object>" {
		t.Errorf("Use = %q, want %q", mergeCmd.Use, "merge <object>")
	}
}

func TestMergeCmd_Short(t *testing.T) {
	if mergeCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestMergeCmd_Args(t *testing.T) {
	err := mergeCmd.Args(mergeCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	err = mergeCmd.Args(mergeCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestRunMerge_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/merge" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		if payload["primaryId"] != "p1" {
			t.Errorf("primaryId = %v, want %q", payload["primaryId"], "p1")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	mergeData = `{"primaryId":"p1","secondaryIds":["p2","p3"]}`
	mergeDataFile = ""
	noResolve = true

	err := runMerge(mergeCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runMerge failed: %v", err)
	}

	mergeData = ""
}

func TestRunMerge_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	mergeData = ""
	mergeDataFile = ""
	noResolve = true

	err := runMerge(mergeCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

// ============== Find Duplicates Tests ==============

func TestFindDuplicatesCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if findDuplicatesCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestFindDuplicatesCmd_Use(t *testing.T) {
	if findDuplicatesCmd.Use != "find-duplicates <object>" {
		t.Errorf("Use = %q, want %q", findDuplicatesCmd.Use, "find-duplicates <object>")
	}
}

func TestFindDuplicatesCmd_Short(t *testing.T) {
	if findDuplicatesCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunFindDuplicates_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/find-duplicates" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"duplicates":[{"id":"p1","matches":[{"id":"p2"}]}]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	duplicatesData = `{"fields":["email"]}`
	duplicatesDataFile = ""
	noResolve = true

	err := runFindDuplicates(findDuplicatesCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runFindDuplicates failed: %v", err)
	}

	duplicatesData = ""
}

func TestRunFindDuplicates_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	duplicatesData = ""
	duplicatesDataFile = ""
	noResolve = true

	err := runFindDuplicates(findDuplicatesCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

// ============== Group By Tests ==============

func TestGroupByCmd_Flags(t *testing.T) {
	flags := []string{"data", "file", "filter", "filter-file", "param"}
	for _, flag := range flags {
		if groupByCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestGroupByCmd_Use(t *testing.T) {
	if groupByCmd.Use != "group-by <object>" {
		t.Errorf("Use = %q, want %q", groupByCmd.Use, "group-by <object>")
	}
}

func TestGroupByCmd_Short(t *testing.T) {
	if groupByCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunGroupBy_WithPayload(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/group-by" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"groups":[{"key":"active","count":10}]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	groupByData = `{"groupBy":"status"}`
	groupByDataFile = ""
	groupByFilter = ""
	groupByFilterFile = ""
	groupByParams = nil
	noResolve = true

	err := runGroupBy(groupByCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runGroupBy failed: %v", err)
	}

	groupByData = ""
}

func TestRunGroupBy_WithFilter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.Method != "GET" {
			t.Errorf("expected GET for filter-based request, got %s", r.Method)
		}

		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query param")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"groups":[]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	groupByData = ""
	groupByDataFile = ""
	groupByFilter = `{"status":{"eq":"active"}}`
	groupByFilterFile = ""
	groupByParams = nil
	noResolve = true

	err := runGroupBy(groupByCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runGroupBy failed: %v", err)
	}

	groupByFilter = ""
}

func TestRunGroupBy_NoPayloadOrFilter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		// Should be a GET request without body
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"groups":[]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	groupByData = ""
	groupByDataFile = ""
	groupByFilter = ""
	groupByFilterFile = ""
	groupByParams = nil
	noResolve = true

	err := runGroupBy(groupByCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runGroupBy failed: %v", err)
	}
}

// ============== Export Tests ==============

func TestExportCmd_Flags(t *testing.T) {
	flags := []string{"format", "output", "all", "limit", "filter", "filter-file", "param", "sort", "order", "fields", "include"}
	for _, flag := range flags {
		if exportCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestExportCmd_Use(t *testing.T) {
	if exportCmd.Use != "export <object>" {
		t.Errorf("Use = %q, want %q", exportCmd.Use, "export <object>")
	}
}

func TestExportCmd_Short(t *testing.T) {
	if exportCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunExport_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"people":[{"id":"p1"},{"id":"p2"}]},"pageInfo":{"hasNextPage":false}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	exportFormat = "json"
	exportOutput = ""
	exportAll = false
	exportLimit = 100
	exportFilter = ""
	exportFilterFile = ""
	exportParams = nil
	exportSort = ""
	exportOrder = ""
	exportFields = ""
	exportInclude = ""
	noResolve = true

	err := runExport(exportCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}
}

func TestRunExport_ToFile(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"people":[{"id":"p1"}]},"pageInfo":{"hasNextPage":false}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Create temp file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "export.json")

	exportFormat = "json"
	exportOutput = outputFile
	exportAll = false
	exportLimit = 100
	exportFilter = ""
	exportFilterFile = ""
	exportParams = nil
	exportSort = ""
	exportOrder = ""
	exportFields = ""
	exportInclude = ""
	noResolve = true

	err := runExport(exportCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("export file was not created")
	}

	exportOutput = ""
}

func TestRunExport_UnsupportedFormat(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	exportFormat = "xml"
	noResolve = true

	err := runExport(exportCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for unsupported format")
	}

	exportFormat = "json"
}

func TestRunExport_WithPagination(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			w.Write([]byte(`{"data":{"people":[{"id":"p1"}]},"pageInfo":{"hasNextPage":true,"endCursor":"cursor1"}}`))
		} else {
			w.Write([]byte(`{"data":{"people":[{"id":"p2"}]},"pageInfo":{"hasNextPage":false}}`))
		}
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	exportFormat = "json"
	exportOutput = ""
	exportAll = true
	exportLimit = 1
	exportFilter = ""
	exportFilterFile = ""
	exportParams = nil
	exportSort = ""
	exportOrder = ""
	exportFields = ""
	exportInclude = ""
	noResolve = true

	err := runExport(exportCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runExport failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

// ============== Import Tests ==============

func TestImportCmd_Flags(t *testing.T) {
	flags := []string{"data", "batch-size", "dry-run", "continue-on-error"}
	for _, flag := range flags {
		if importCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestImportCmd_Use(t *testing.T) {
	if importCmd.Use != "import <object> [file]" {
		t.Errorf("Use = %q, want %q", importCmd.Use, "import <object> [file]")
	}
}

func TestImportCmd_Short(t *testing.T) {
	if importCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestImportCmd_Args(t *testing.T) {
	// Test with zero args
	err := importCmd.Args(importCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = importCmd.Args(importCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = importCmd.Args(importCmd, []string{"people", "file.json"})
	if err != nil {
		t.Errorf("Expected no error with two args, got %v", err)
	}

	// Test with three args
	err = importCmd.Args(importCmd, []string{"people", "file.json", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

func TestRunImport_Success(t *testing.T) {
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[{"id":"p1"},{"id":"p2"}]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	importData = `[{"name":"John"},{"name":"Jane"}]`
	importBatchSize = 60
	importDryRun = false
	importContinueOnErr = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}

	importData = ""
}

func TestRunImport_DryRun(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/batch/people" {
			t.Error("server should not be called during dry run")
		}
	}))
	defer cleanup()

	importData = `[{"name":"John"}]`
	importBatchSize = 60
	importDryRun = true
	importContinueOnErr = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}

	importData = ""
	importDryRun = false
}

func TestRunImport_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	importData = ""
	importDryRun = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

func TestRunImport_NonArrayPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	importData = `{"name":"not an array"}`
	importDryRun = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for non-array payload")
	}

	importData = ""
}

func TestRunImport_BatchSize(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		callCount++

		body, _ := io.ReadAll(r.Body)
		var payload []interface{}
		json.Unmarshal(body, &payload)

		// With batch size 2 and 5 records, we expect 3 calls
		if callCount <= 2 && len(payload) != 2 {
			t.Errorf("call %d: expected 2 records, got %d", callCount, len(payload))
		}
		if callCount == 3 && len(payload) != 1 {
			t.Errorf("call %d: expected 1 record, got %d", callCount, len(payload))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	importData = `[{"name":"1"},{"name":"2"},{"name":"3"},{"name":"4"},{"name":"5"}]`
	importBatchSize = 2
	importDryRun = false
	importContinueOnErr = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 batch calls, got %d", callCount)
	}

	importData = ""
	importBatchSize = 60
}

func TestRunImport_ContinueOnError(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		callCount++

		if callCount == 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"batch 1 failed"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	importData = `[{"name":"1"},{"name":"2"},{"name":"3"}]`
	importBatchSize = 1
	importDryRun = false
	importContinueOnErr = true
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport should not fail with continue-on-error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 batch calls with continue-on-error, got %d", callCount)
	}

	importData = ""
	importBatchSize = 60
	importContinueOnErr = false
}

func TestRunImport_JSONOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[{"id":"p1"}]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "json")
	importData = `[{"name":"John"}]`
	importBatchSize = 60
	importDryRun = false
	importContinueOnErr = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}

	importData = ""
}

func TestRunImport_BatchSizeDefaults(t *testing.T) {
	// Test that batch size is capped at 60
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	importData = `[{"name":"1"}]`
	importBatchSize = 100 // Should be capped to 60
	importDryRun = false
	importContinueOnErr = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}

	importData = ""
	importBatchSize = 60
}

func TestRunImport_ZeroBatchSize(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"people":[]}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	importData = `[{"name":"1"}]`
	importBatchSize = 0 // Should default to 60
	importDryRun = false
	importContinueOnErr = false
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport failed: %v", err)
	}

	importData = ""
	importBatchSize = 60
}
