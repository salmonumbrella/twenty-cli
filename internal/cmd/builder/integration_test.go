// internal/cmd/builder/integration_test.go
package builder

import (
	"bytes"
	"context"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestListCommand_Integration(t *testing.T) {
	// Skip if no API credentials
	t.Skip("integration test - requires API credentials")

	cfg := ListConfig[types.Person]{
		Use:   "list",
		Short: "List people",
		ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
			return client.ListPeople(ctx, opts)
		},
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	cmd := NewListCommand(cfg)

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--limit", "1", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

func TestGetCommand_Integration(t *testing.T) {
	// Skip if no API credentials
	t.Skip("integration test - requires API credentials")

	cfg := GetConfig[types.Person]{
		Use:   "get",
		Short: "Get a person",
		GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
			return client.GetPerson(ctx, id, nil)
		},
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
				{"First Name", p.Name.FirstName},
				{"Last Name", p.Name.LastName},
			}
		},
	}

	cmd := NewGetCommand(cfg)

	var out bytes.Buffer
	cmd.SetOut(&out)
	// Note: This would need a valid person ID
	cmd.SetArgs([]string{"test-id", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output, got empty")
	}
}
