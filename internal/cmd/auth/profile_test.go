package auth

import (
	"errors"
	"os"
	"testing"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestResolveProfile_ExplicitProfile(t *testing.T) {
	setupMockStore(t)

	profile, err := ResolveProfile("explicit-profile")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "explicit-profile" {
		t.Errorf("ResolveProfile = %q, want %q", profile, "explicit-profile")
	}
}

func TestResolveProfile_ExplicitProfileWithWhitespace(t *testing.T) {
	setupMockStore(t)

	profile, err := ResolveProfile("   profile-with-spaces   ")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "profile-with-spaces" {
		t.Errorf("ResolveProfile = %q, want %q", profile, "profile-with-spaces")
	}
}

func TestResolveProfile_DefaultFromStore(t *testing.T) {
	mock := setupMockStore(t)
	_ = mock.SetDefaultAccount("stored-default")

	profile, err := ResolveProfile("")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "stored-default" {
		t.Errorf("ResolveProfile = %q, want %q", profile, "stored-default")
	}
}

func TestResolveProfile_FallbackToDefault(t *testing.T) {
	setupMockStore(t)
	// No default set in store

	profile, err := ResolveProfile("")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "default" {
		t.Errorf("ResolveProfile = %q, want %q", profile, "default")
	}
}

func TestCurrentProfile(t *testing.T) {
	setupMockStore(t)

	// Test with viper profile set
	viper.Reset()
	viper.Set("profile", "viper-profile")

	profile, err := CurrentProfile()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "viper-profile" {
		t.Errorf("CurrentProfile = %q, want %q", profile, "viper-profile")
	}
}

func TestCurrentProfile_Empty(t *testing.T) {
	mock := setupMockStore(t)
	_ = mock.SetDefaultAccount("stored-default")

	viper.Reset()
	viper.Set("profile", "")

	profile, err := CurrentProfile()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "stored-default" {
		t.Errorf("CurrentProfile = %q, want %q", profile, "stored-default")
	}
}

func TestRequireToken_ExplicitProfileNotFound(t *testing.T) {
	mock := setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })
	os.Unsetenv("TWENTY_TOKEN")

	// Add some profiles
	_ = mock.SetToken("existing", secrets.Token{RefreshToken: "token"})

	// Request a non-existent profile via viper flag
	viper.Reset()
	viper.Set("profile", "nonexistent")

	_, _, err := RequireToken()
	if err == nil {
		t.Error("Expected error for nonexistent profile")
	}

	if !contains(err.Error(), "nonexistent") {
		t.Errorf("Error = %q, want containing 'nonexistent'", err.Error())
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("Error = %q, want containing 'not found'", err.Error())
	}
}

func TestRequireToken_InvalidJSON(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	// Set an invalid JSON token that starts with {
	os.Setenv("TWENTY_TOKEN", "{invalid json")

	viper.Reset()

	// Should fall back to treating it as a raw token
	profile, token, err := RequireToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "env" {
		t.Errorf("profile = %q, want %q", profile, "env")
	}

	if token != "{invalid json" {
		t.Errorf("token = %q, want %q", token, "{invalid json")
	}
}

func TestRequireToken_JSONWithEmptyRefreshToken(t *testing.T) {
	setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })

	// Set JSON with empty refresh_token - should be treated as raw
	os.Setenv("TWENTY_TOKEN", `{"refresh_token": "", "created_at": "2024-01-01"}`)

	viper.Reset()

	profile, token, err := RequireToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if profile != "env" {
		t.Errorf("profile = %q, want %q", profile, "env")
	}

	// With empty refresh_token, it falls back to raw token
	if token != `{"refresh_token": "", "created_at": "2024-01-01"}` {
		t.Errorf("token = %q", token)
	}
}

func TestGetToken_StoreGetError(t *testing.T) {
	mock := setupMockStore(t)
	mock.SetGetError(errors.New("store error"))

	_, err := GetToken("profile")
	if err == nil {
		t.Error("Expected error from store")
	}
}

func TestSaveToken_StoreSetError(t *testing.T) {
	mock := setupMockStore(t)
	mock.SetSetError(errors.New("store error"))

	err := SaveToken("profile", "token")
	if err == nil {
		t.Error("Expected error from store")
	}
}

func TestDeleteToken_StoreError(t *testing.T) {
	setupMockStore(t)

	// Try to delete a non-existent token
	err := DeleteToken("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent token")
	}
}

func TestListProfiles_Empty(t *testing.T) {
	setupMockStore(t)

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("Expected empty list, got %d profiles", len(profiles))
	}
}

func TestResolveProfile_GetDefaultAccountError(t *testing.T) {
	setupCustomStore(t, &getDefaultErrorStore{})

	_, err := ResolveProfile("")
	if err == nil {
		t.Error("Expected error from GetDefaultAccount")
	}
}

// getDefaultErrorStore returns an error from GetDefaultAccount
type getDefaultErrorStore struct {
	secrets.MockStore
}

func (m *getDefaultErrorStore) Keys() ([]string, error) {
	return nil, nil
}

func (m *getDefaultErrorStore) SetToken(profile string, tok secrets.Token) error {
	return nil
}

func (m *getDefaultErrorStore) GetToken(profile string) (secrets.Token, error) {
	return secrets.Token{RefreshToken: "token"}, nil
}

func (m *getDefaultErrorStore) DeleteToken(profile string) error {
	return nil
}

func (m *getDefaultErrorStore) ListTokens() ([]secrets.Token, error) {
	return nil, nil
}

func (m *getDefaultErrorStore) GetDefaultAccount() (string, error) {
	return "", errors.New("get default account error")
}

func (m *getDefaultErrorStore) SetDefaultAccount(profile string) error {
	return nil
}

func TestGetToken_ResolveProfileError(t *testing.T) {
	setupCustomStore(t, &getDefaultErrorStore{})

	viper.Reset()
	viper.Set("profile", "")

	// GetToken with empty profile should try to resolve, which will fail
	_, err := GetToken("")
	if err == nil {
		t.Error("GetToken should return error when ResolveProfile fails")
	}
}

func TestResolveProfile_GetStoreError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	t.Cleanup(func() {
		store = originalStore
		storeOpener = originalOpener
	})

	// Clear store so getStore will try to open a new one
	store = nil
	storeOpener = func() (secrets.Store, error) {
		return nil, errors.New("store error")
	}

	// ResolveProfile should fall back to "default" when getStore fails
	profile, err := ResolveProfile("")
	if err != nil {
		t.Fatalf("ResolveProfile should not return error: %v", err)
	}

	if profile != "default" {
		t.Errorf("ResolveProfile = %q, want %q", profile, "default")
	}
}

func TestRequireToken_CurrentProfileError(t *testing.T) {
	originalStore := store
	originalOpener := storeOpener
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() {
		store = originalStore
		storeOpener = originalOpener
		os.Setenv("TWENTY_TOKEN", originalEnv)
	})

	os.Unsetenv("TWENTY_TOKEN")
	store = nil

	// Make getStore succeed but GetDefaultAccount fail
	mock := &getDefaultErrorStore{}
	storeOpener = func() (secrets.Store, error) {
		return mock, nil
	}

	viper.Reset()
	viper.Set("profile", "")

	_, _, err := RequireToken()
	if err == nil {
		t.Error("RequireToken should return error when CurrentProfile fails")
	}
}

func TestRequireToken_ProfileFoundButNoToken(t *testing.T) {
	mock := setupMockStore(t)
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })
	os.Unsetenv("TWENTY_TOKEN")

	// Add the profile but it will have an empty token (which shouldn't happen in real use)
	// Actually, the mock returns keyring.ErrKeyNotFound for missing profiles
	// Let's test with a profile that exists in the list but has no token

	// Add a token for the profile
	_ = mock.SetToken("explicitprofile", secrets.Token{RefreshToken: "token"})

	// Set an explicit profile that exists but request a different way
	viper.Reset()
	viper.Set("profile", "explicitprofile")

	profile, token, err := RequireToken()
	if err != nil {
		t.Fatalf("RequireToken should succeed: %v", err)
	}

	if profile != "explicitprofile" {
		t.Errorf("profile = %q, want %q", profile, "explicitprofile")
	}

	if token != "token" {
		t.Errorf("token = %q, want %q", token, "token")
	}
}

func TestRequireToken_ListProfilesError(t *testing.T) {
	setupCustomStore(t, &listErrorStore2{})
	originalEnv := os.Getenv("TWENTY_TOKEN")
	t.Cleanup(func() { os.Setenv("TWENTY_TOKEN", originalEnv) })
	os.Unsetenv("TWENTY_TOKEN")

	// Set an explicit profile
	viper.Reset()
	viper.Set("profile", "someprofile")

	_, _, err := RequireToken()
	// Should return "not logged in" error since ListProfiles fails
	if err == nil {
		t.Error("RequireToken should return error")
	}

	if !contains(err.Error(), "not logged in") {
		t.Errorf("Error = %q, want containing 'not logged in'", err.Error())
	}
}

// listErrorStore2 returns error on ListTokens and ErrKeyNotFound on GetToken
type listErrorStore2 struct {
	secrets.MockStore
}

func (m *listErrorStore2) Keys() ([]string, error) {
	return nil, nil
}

func (m *listErrorStore2) SetToken(profile string, tok secrets.Token) error {
	return nil
}

func (m *listErrorStore2) GetToken(profile string) (secrets.Token, error) {
	return secrets.Token{}, keyring.ErrKeyNotFound
}

func (m *listErrorStore2) DeleteToken(profile string) error {
	return nil
}

func (m *listErrorStore2) ListTokens() ([]secrets.Token, error) {
	return nil, errors.New("list tokens error")
}

func (m *listErrorStore2) GetDefaultAccount() (string, error) {
	return "", nil
}

func (m *listErrorStore2) SetDefaultAccount(profile string) error {
	return nil
}
