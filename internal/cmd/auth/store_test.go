package auth

import (
	"errors"
	"os"
	"testing"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestSetStore(t *testing.T) {
	// Save original store to restore after test
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()

	// Set the mock store
	SetStore(mock)

	// Verify store was set by using it
	tok := secrets.Token{RefreshToken: "test-token"}
	err := mock.SetToken("test-profile", tok)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Use GetToken through the auth package to verify store is wired up
	token, err := GetToken("test-profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "test-token" {
		t.Errorf("GetToken = %q, want %q", token, "test-token")
	}
}

func TestSetStore_ReplacesPreviousStore(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	// Set first mock
	mock1 := secrets.NewMockStore()
	SetStore(mock1)
	_ = mock1.SetToken("profile1", secrets.Token{RefreshToken: "token1"})

	// Set second mock
	mock2 := secrets.NewMockStore()
	SetStore(mock2)
	_ = mock2.SetToken("profile2", secrets.Token{RefreshToken: "token2"})

	// Verify first mock's data is no longer accessible
	_, err := GetToken("profile1")
	if err == nil {
		t.Error("GetToken should fail for profile1 after store replacement")
	}

	// Verify second mock's data is accessible
	token, err := GetToken("profile2")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if token != "token2" {
		t.Errorf("GetToken = %q, want %q", token, "token2")
	}
}

func TestRequireToken_NoStore(t *testing.T) {
	// Save and restore original state
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	// Clear environment variable to force keychain lookup
	os.Unsetenv("TWENTY_TOKEN")

	// Use a mock store with no tokens
	mock := secrets.NewMockStore()
	SetStore(mock)

	// Clear any viper profile setting
	viper.Set("profile", "")

	// RequireToken should fail because no token is available
	_, _, err := RequireToken()
	if err == nil {
		t.Error("RequireToken should return error when no token is available")
	}

	// Error message should suggest login
	if err != nil && !contains(err.Error(), "not logged in") {
		t.Errorf("RequireToken error = %q, want error containing 'not logged in'", err.Error())
	}
}

func TestRequireToken_WithEnvToken(t *testing.T) {
	// Save and restore original state
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	// Set token via environment variable
	os.Setenv("TWENTY_TOKEN", "env-test-token")

	// Use a mock store (should not be consulted)
	mock := secrets.NewMockStore()
	SetStore(mock)

	profile, token, err := RequireToken()
	if err != nil {
		t.Fatalf("RequireToken failed: %v", err)
	}

	if profile != "env" {
		t.Errorf("RequireToken profile = %q, want %q", profile, "env")
	}

	if token != "env-test-token" {
		t.Errorf("RequireToken token = %q, want %q", token, "env-test-token")
	}
}

func TestRequireToken_WithEnvTokenJSON(t *testing.T) {
	// Save and restore original state
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	// Set token via environment variable as JSON
	os.Setenv("TWENTY_TOKEN", `{"refresh_token": "json-env-token", "created_at": "2024-01-01"}`)

	mock := secrets.NewMockStore()
	SetStore(mock)

	profile, token, err := RequireToken()
	if err != nil {
		t.Fatalf("RequireToken failed: %v", err)
	}

	if profile != "env" {
		t.Errorf("RequireToken profile = %q, want %q", profile, "env")
	}

	if token != "json-env-token" {
		t.Errorf("RequireToken token = %q, want %q", token, "json-env-token")
	}
}

func TestRequireToken_WithStoredToken(t *testing.T) {
	// Save and restore original state
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	// Clear environment variable
	os.Unsetenv("TWENTY_TOKEN")

	// Set up mock store with a token
	mock := secrets.NewMockStore()
	SetStore(mock)
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "stored-token"})

	// Clear viper profile setting to use default
	viper.Set("profile", "")

	profile, token, err := RequireToken()
	if err != nil {
		t.Fatalf("RequireToken failed: %v", err)
	}

	if profile != "default" {
		t.Errorf("RequireToken profile = %q, want %q", profile, "default")
	}

	if token != "stored-token" {
		t.Errorf("RequireToken token = %q, want %q", token, "stored-token")
	}
}

func TestRequireToken_ProfileNotFound(t *testing.T) {
	// Save and restore original state
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	// Clear environment variable
	os.Unsetenv("TWENTY_TOKEN")

	// Set up mock store without the requested profile
	mock := secrets.NewMockStore()
	SetStore(mock)
	_ = mock.SetToken("other-profile", secrets.Token{RefreshToken: "other-token"})

	// Request a non-existent profile
	viper.Set("profile", "missing-profile")

	_, _, err := RequireToken()
	if err == nil {
		t.Error("RequireToken should return error for missing profile")
	}

	// Error message should mention the profile
	if err != nil && !contains(err.Error(), "missing-profile") {
		t.Errorf("RequireToken error = %q, want error mentioning 'missing-profile'", err.Error())
	}
}

func TestSaveToken(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	err := SaveToken("test-profile", "test-api-token")
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify token was saved
	tok, err := mock.GetToken("test-profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if tok.RefreshToken != "test-api-token" {
		t.Errorf("Saved token = %q, want %q", tok.RefreshToken, "test-api-token")
	}
}

func TestDeleteToken(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Save a token first
	_ = mock.SetToken("test-profile", secrets.Token{RefreshToken: "test-token"})

	// Delete it
	err := DeleteToken("test-profile")
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify it's gone
	_, err = mock.GetToken("test-profile")
	if err == nil {
		t.Error("Token should be deleted")
	}
}

func TestListProfiles(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Add some tokens
	_ = mock.SetToken("profile-a", secrets.Token{RefreshToken: "token-a"})
	_ = mock.SetToken("profile-b", secrets.Token{RefreshToken: "token-b"})

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("ListProfiles returned %d profiles, want 2", len(profiles))
	}

	// Check both profiles are present
	found := make(map[string]bool)
	for _, p := range profiles {
		found[p] = true
	}

	if !found["profile-a"] {
		t.Error("ListProfiles missing 'profile-a'")
	}
	if !found["profile-b"] {
		t.Error("ListProfiles missing 'profile-b'")
	}
}

func TestGetToken_EmptyProfile(t *testing.T) {
	originalStore := store
	originalEnv := os.Getenv("TWENTY_TOKEN")
	defer func() {
		store = originalStore
		os.Setenv("TWENTY_TOKEN", originalEnv)
	}()

	os.Unsetenv("TWENTY_TOKEN")

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Store a token for the default profile
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "default-token"})

	// Clear viper to use default profile resolution
	viper.Set("profile", "")

	// GetToken with empty profile should resolve to default
	token, err := GetToken("")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "default-token" {
		t.Errorf("GetToken = %q, want %q", token, "default-token")
	}
}

func TestGetToken_StoreError(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Request a token that doesn't exist
	_, err := GetToken("nonexistent")
	if err == nil {
		t.Error("GetToken should return error for nonexistent profile")
	}

	// Should be a keyring not found error
	if err != nil && !contains(err.Error(), keyring.ErrKeyNotFound.Error()) {
		t.Errorf("GetToken error = %q, want keyring not found error", err.Error())
	}
}

func TestListProfiles_TokensError(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := &errorMockStore{listErr: errors.New("list error")}
	SetStore(mock)

	_, err := ListProfiles()
	if err == nil {
		t.Error("ListProfiles should return error when ListTokens fails")
	}
}

// errorMockStore is a mock that can return errors for specific operations
type errorMockStore struct {
	secrets.MockStore
	getStoreErr error
	listErr     error
}

func (m *errorMockStore) Keys() ([]string, error) {
	return nil, nil
}

func (m *errorMockStore) SetToken(profile string, tok secrets.Token) error {
	return nil
}

func (m *errorMockStore) GetToken(profile string) (secrets.Token, error) {
	return secrets.Token{}, keyring.ErrKeyNotFound
}

func (m *errorMockStore) DeleteToken(profile string) error {
	return nil
}

func (m *errorMockStore) ListTokens() ([]secrets.Token, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return nil, nil
}

func (m *errorMockStore) GetDefaultAccount() (string, error) {
	return "", nil
}

func (m *errorMockStore) SetDefaultAccount(profile string) error {
	return nil
}

// Tests for getStore nil case - these test behavior when store needs to be opened
// but we can't easily test secrets.OpenDefault() without integration tests

func TestGetStore_ReturnsExistingStore(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// getStore should return the existing store
	s, err := getStore()
	if err != nil {
		t.Fatalf("getStore failed: %v", err)
	}

	if s != mock {
		t.Error("getStore should return the set store")
	}
}

func TestSaveToken_EmptyProfile(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// SaveToken with empty profile should fail at the store level
	err := SaveToken("", "token")
	if err == nil {
		t.Error("SaveToken with empty profile should fail")
	}
}

func TestGetToken_ResolveProfileWithStore(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Set up a token for the default profile
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "test-token"})

	// Clear viper
	viper.Set("profile", "")

	// GetToken with empty profile should resolve to default
	token, err := GetToken("")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "test-token" {
		t.Errorf("GetToken = %q, want %q", token, "test-token")
	}
}

func TestDeleteToken_ExistingToken(t *testing.T) {
	originalStore := store
	defer func() { store = originalStore }()

	mock := secrets.NewMockStore()
	SetStore(mock)

	// Add a token
	_ = mock.SetToken("test-profile", secrets.Token{RefreshToken: "test-token"})

	// Delete it
	err := DeleteToken("test-profile")
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify it's gone
	_, err = mock.GetToken("test-profile")
	if err == nil {
		t.Error("Token should be deleted")
	}
}

// Tests for getStore with nil store - these test the storeOpener path

func TestGetStore_OpensNewStore(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalOpener
	}()

	// Clear store so getStore will try to open a new one
	store = nil

	mock := secrets.NewMockStore()
	storeOpener = func() (secrets.Store, error) {
		return mock, nil
	}

	s, err := getStore()
	if err != nil {
		t.Fatalf("getStore failed: %v", err)
	}

	if s != mock {
		t.Error("getStore should return the opened store")
	}

	// Verify store is now set
	if store != mock {
		t.Error("getStore should set the store variable")
	}
}

func TestGetStore_OpenerError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalOpener
	}()

	// Clear store so getStore will try to open a new one
	store = nil

	expectedErr := errors.New("failed to open store")
	storeOpener = func() (secrets.Store, error) {
		return nil, expectedErr
	}

	_, err := getStore()
	if err == nil {
		t.Error("getStore should return error when opener fails")
	}

	if err != expectedErr {
		t.Errorf("getStore error = %v, want %v", err, expectedErr)
	}
}

func TestSaveToken_GetStoreError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalOpener
	}()

	store = nil
	storeOpener = func() (secrets.Store, error) {
		return nil, errors.New("store error")
	}

	err := SaveToken("profile", "token")
	if err == nil {
		t.Error("SaveToken should return error when getStore fails")
	}
}

func TestGetToken_GetStoreError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalOpener
	}()

	store = nil
	storeOpener = func() (secrets.Store, error) {
		return nil, errors.New("store error")
	}

	_, err := GetToken("profile")
	if err == nil {
		t.Error("GetToken should return error when getStore fails")
	}
}

func TestDeleteToken_GetStoreError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalOpener
	}()

	store = nil
	storeOpener = func() (secrets.Store, error) {
		return nil, errors.New("store error")
	}

	err := DeleteToken("profile")
	if err == nil {
		t.Error("DeleteToken should return error when getStore fails")
	}
}

func TestListProfiles_GetStoreError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	defer func() {
		store = originalStore
		storeOpener = originalOpener
	}()

	store = nil
	storeOpener = func() (secrets.Store, error) {
		return nil, errors.New("store error")
	}

	_, err := ListProfiles()
	if err == nil {
		t.Error("ListProfiles should return error when getStore fails")
	}
}

func TestSetStoreOpener(t *testing.T) {
	originalOpener := storeOpener
	defer func() { storeOpener = originalOpener }()

	called := false
	newOpener := func() (secrets.Store, error) {
		called = true
		return nil, nil
	}

	SetStoreOpener(newOpener)

	// Verify it was set by checking if it gets called when store is nil
	originalStore := store
	defer func() { store = originalStore }()
	store = nil

	_, _ = getStore()

	if !called {
		t.Error("SetStoreOpener should set the opener function")
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
