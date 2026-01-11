package records

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "records" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "records")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCmd_Long(t *testing.T) {
	if Cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestCmd_HasSubcommands(t *testing.T) {
	subcommands := []string{
		"list", "get", "create", "update", "delete", "destroy", "restore",
		"batch-create", "batch-update", "batch-delete", "batch-destroy", "batch-restore",
		"merge", "find-duplicates", "group-by", "export", "import",
	}
	for _, name := range subcommands {
		found := false
		for _, cmd := range Cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

func TestCmd_SubcommandCount(t *testing.T) {
	expected := 17 // list, get, create, update, delete, destroy, restore, batch-create, batch-update, batch-delete, batch-destroy, batch-restore, merge, find-duplicates, group-by, export, import
	if len(Cmd.Commands()) != expected {
		t.Errorf("expected %d subcommands, got %d", expected, len(Cmd.Commands()))
	}
}

func TestCmd_NoResolveFlag(t *testing.T) {
	flag := Cmd.PersistentFlags().Lookup("no-resolve")
	if flag == nil {
		t.Error("no-resolve flag not registered")
	}
}

func TestCmd_ListSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "list" {
			if cmd.Short == "" {
				t.Error("list subcommand should have short description")
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestCmd_GetSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "get" {
			if cmd.Short == "" {
				t.Error("get subcommand should have short description")
			}
			return
		}
	}
	t.Error("get subcommand not found")
}

func TestCmd_CreateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "create" {
			if cmd.Short == "" {
				t.Error("create subcommand should have short description")
			}
			return
		}
	}
	t.Error("create subcommand not found")
}

func TestCmd_UpdateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "update" {
			if cmd.Short == "" {
				t.Error("update subcommand should have short description")
			}
			return
		}
	}
	t.Error("update subcommand not found")
}

func TestCmd_DeleteSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "delete" {
			if cmd.Short == "" {
				t.Error("delete subcommand should have short description")
			}
			return
		}
	}
	t.Error("delete subcommand not found")
}

func TestCmd_DestroySubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "destroy" {
			if cmd.Short == "" {
				t.Error("destroy subcommand should have short description")
			}
			return
		}
	}
	t.Error("destroy subcommand not found")
}

func TestCmd_RestoreSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "restore" {
			if cmd.Short == "" {
				t.Error("restore subcommand should have short description")
			}
			return
		}
	}
	t.Error("restore subcommand not found")
}

func TestCmd_BatchCreateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "batch-create" {
			if cmd.Short == "" {
				t.Error("batch-create subcommand should have short description")
			}
			return
		}
	}
	t.Error("batch-create subcommand not found")
}

func TestCmd_BatchUpdateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "batch-update" {
			if cmd.Short == "" {
				t.Error("batch-update subcommand should have short description")
			}
			return
		}
	}
	t.Error("batch-update subcommand not found")
}

func TestCmd_BatchDeleteSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "batch-delete" {
			if cmd.Short == "" {
				t.Error("batch-delete subcommand should have short description")
			}
			return
		}
	}
	t.Error("batch-delete subcommand not found")
}

func TestCmd_BatchDestroySubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "batch-destroy" {
			if cmd.Short == "" {
				t.Error("batch-destroy subcommand should have short description")
			}
			return
		}
	}
	t.Error("batch-destroy subcommand not found")
}

func TestCmd_BatchRestoreSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "batch-restore" {
			if cmd.Short == "" {
				t.Error("batch-restore subcommand should have short description")
			}
			return
		}
	}
	t.Error("batch-restore subcommand not found")
}

func TestCmd_MergeSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "merge" {
			if cmd.Short == "" {
				t.Error("merge subcommand should have short description")
			}
			return
		}
	}
	t.Error("merge subcommand not found")
}

func TestCmd_FindDuplicatesSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "find-duplicates" {
			if cmd.Short == "" {
				t.Error("find-duplicates subcommand should have short description")
			}
			return
		}
	}
	t.Error("find-duplicates subcommand not found")
}

func TestCmd_GroupBySubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "group-by" {
			if cmd.Short == "" {
				t.Error("group-by subcommand should have short description")
			}
			return
		}
	}
	t.Error("group-by subcommand not found")
}

func TestCmd_ExportSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "export" {
			if cmd.Short == "" {
				t.Error("export subcommand should have short description")
			}
			return
		}
	}
	t.Error("export subcommand not found")
}

func TestCmd_ImportSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "import" {
			if cmd.Short == "" {
				t.Error("import subcommand should have short description")
			}
			return
		}
	}
	t.Error("import subcommand not found")
}
