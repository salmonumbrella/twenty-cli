package records

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

// Additional edge case tests to improve coverage

// ============== List Edge Cases ==============

func TestRunList_ParseQueryParamsError(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	listFilter = ""
	listFilterFile = ""
	listParams = []string{"invalid-no-equals"}
	noResolve = true
	listAll = false

	err := runList(listCmd, []string{"items"})
	if err == nil {
		t.Error("expected error for invalid param format")
	}

	listParams = nil
}

func TestRunList_ExtractListError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		// Return response without proper data structure
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result":"no data field"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	noResolve = true
	listAll = true
	listFilter = ""
	listParams = nil

	err := runList(listCmd, []string{"items"})
	if err == nil {
		t.Error("expected error when extractList fails")
	}

	listAll = false
}

// ============== Get Edge Cases ==============

func TestRunGet_ParseQueryParamsError(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	getFields = ""
	getInclude = ""
	getParams = []string{"invalid"}
	noResolve = true

	err := runGet(getCmd, []string{"people", "p1"})
	if err == nil {
		t.Error("expected error for invalid param format")
	}

	getParams = nil
}

// ============== Create Edge Cases ==============

func TestRunCreate_ResolveObjectError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			// Return valid object list
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"objects": []interface{}{
						map[string]interface{}{
							"namePlural":   "persons",
							"nameSingular": "person",
						},
					},
				},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"person":{"id":"new-id"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	createData = `{"name":"Test"}`
	createDataFile = ""
	createSet = nil
	noResolve = false // Test object resolution

	err := runCreate(createCmd, []string{"person"})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	createData = ""
	noResolve = true
}

// ============== Update Edge Cases ==============

func TestRunUpdate_ResolveObjectError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"objects": []interface{}{
						map[string]interface{}{
							"namePlural":   "persons",
							"nameSingular": "person",
						},
					},
				},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	updateData = `{"name":"Updated"}`
	updateDataFile = ""
	updateSet = nil
	noResolve = false

	err := runUpdate(updateCmd, []string{"person", "p1"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	updateData = ""
	noResolve = true
}

// ============== Delete Edge Cases ==============

func TestRunDelete_ResolveObject(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"objects": []interface{}{
						map[string]interface{}{
							"namePlural":   "persons",
							"nameSingular": "person",
						},
					},
				},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = true
	noResolve = false

	err := runDelete(deleteCmd, []string{"person", "p1"})
	if err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}

	noResolve = true
}

// ============== Import Edge Cases ==============

func TestRunImport_TextOutput(t *testing.T) {
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

	viper.Set("output", "text")
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

func TestRunImport_WithErrors(t *testing.T) {
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"batch failed"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	importData = `[{"name":"1"},{"name":"2"}]`
	importBatchSize = 1
	importDryRun = false
	importContinueOnErr = true
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport should not fail with continue-on-error, got %v", err)
	}

	importData = ""
	importBatchSize = 60
	importContinueOnErr = false
}

func TestRunImport_StopOnFirstError(t *testing.T) {
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"batch failed"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	importData = `[{"name":"1"},{"name":"2"}]`
	importBatchSize = 1
	importDryRun = false
	importContinueOnErr = false // Stop on first error
	noResolve = true

	err := runImport(importCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runImport should complete without error, got %v", err)
	}

	// Should only have 1 call since we stop on first error
	if callCount != 1 {
		t.Errorf("expected 1 call (stop on error), got %d", callCount)
	}

	importData = ""
	importBatchSize = 60
}

// ============== Export Edge Cases ==============

func TestRunExport_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
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
	noResolve = true

	err := runExport(exportCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}
}

func TestRunExport_ExtractListError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result":"no data field"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	exportFormat = "json"
	exportOutput = ""
	exportAll = false
	exportLimit = 100
	noResolve = true

	err := runExport(exportCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when extractList fails")
	}
}

// ============== Batch Operation Edge Cases ==============

func TestRunBatchCreate_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid data"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	batchCreateData = `[{"name":"John"}]`
	batchCreateDataFile = ""
	noResolve = true

	err := runBatchCreate(batchCreateCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	batchCreateData = ""
}

func TestRunBatchUpdate_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid data"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	batchUpdateData = `[{"id":"p1","name":"Updated"}]`
	batchUpdateDataFile = ""
	noResolve = true

	err := runBatchUpdate(batchUpdateCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	batchUpdateData = ""
}

func TestRunBatchDelete_NoPayload(t *testing.T) {
	originalForce := batchDeleteForce
	defer func() { batchDeleteForce = originalForce }()

	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	batchDeleteForce = true
	batchDeleteData = ""
	batchDeleteDataFile = ""
	batchDeleteIDs = ""
	noResolve = true

	err := runBatchDelete(batchDeleteCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

func TestRunBatchDestroy_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid data"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := batchDestroyForce
	defer func() { batchDestroyForce = originalForce }()

	batchDestroyForce = true
	batchDestroyIDs = "id1,id2"
	noResolve = true

	err := runBatchDestroy(batchDestroyCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	batchDestroyIDs = ""
}

func TestRunBatchRestore_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid data"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	batchRestoreIDs = "id1,id2"
	noResolve = true

	err := runBatchRestore(batchRestoreCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	batchRestoreIDs = ""
}

// ============== Merge/FindDuplicates/GroupBy Edge Cases ==============

func TestRunMerge_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"merge failed"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	mergeData = `{"primaryId":"p1","secondaryIds":["p2"]}`
	mergeDataFile = ""
	noResolve = true

	err := runMerge(mergeCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	mergeData = ""
}

func TestRunFindDuplicates_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"find duplicates failed"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	duplicatesData = `{"fields":["email"]}`
	duplicatesDataFile = ""
	noResolve = true

	err := runFindDuplicates(findDuplicatesCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	duplicatesData = ""
}

func TestRunGroupBy_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"group by failed"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	groupByData = `{"groupBy":"status"}`
	groupByDataFile = ""
	groupByFilter = ""
	groupByParams = nil
	noResolve = true

	err := runGroupBy(groupByCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	groupByData = ""
}

func TestRunGroupBy_InvalidFilter(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	groupByData = ""
	groupByFilter = "{invalid"
	groupByParams = nil
	noResolve = true

	err := runGroupBy(groupByCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for invalid filter JSON")
	}

	groupByFilter = ""
}
