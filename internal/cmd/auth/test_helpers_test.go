package auth

import (
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

// setupMockStore sets up a mock store for testing and returns the mock.
// It automatically restores the original store when the test completes.
func setupMockStore(t *testing.T) *secrets.MockStore {
	t.Helper()
	originalStore := store
	t.Cleanup(func() { store = originalStore })
	mock := secrets.NewMockStore()
	SetStore(mock)
	return mock
}

// setupCustomStore sets up a custom store for testing.
// It automatically restores the original store when the test completes.
func setupCustomStore(t *testing.T, s secrets.Store) {
	t.Helper()
	originalStore := store
	t.Cleanup(func() { store = originalStore })
	SetStore(s)
}
