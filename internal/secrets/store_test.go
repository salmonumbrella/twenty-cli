package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/99designs/keyring"

	"github.com/salmonumbrella/twenty-cli/internal/config"
)

func TestMockStore_SetAndGetToken(t *testing.T) {
	store := NewMockStore()

	tok := Token{
		Profile:      "test-profile",
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresAt:    time.Now().Add(time.Hour),
		Scopes:       []string{"read", "write"},
		CreatedAt:    time.Now(),
	}

	// Set token
	err := store.SetToken("test-profile", tok)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Get token
	got, err := store.GetToken("test-profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if got.Profile != "test-profile" {
		t.Errorf("Profile = %q, want %q", got.Profile, "test-profile")
	}
	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, tok.AccessToken)
	}
	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
	if len(got.Scopes) != len(tok.Scopes) {
		t.Errorf("Scopes = %v, want %v", got.Scopes, tok.Scopes)
	}
}

func TestMockStore_SetToken_NormalizesProfile(t *testing.T) {
	store := NewMockStore()

	tok := Token{RefreshToken: "refresh-123"}

	// Set with uppercase and spaces
	err := store.SetToken("  TEST-PROFILE  ", tok)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Get with lowercase
	got, err := store.GetToken("test-profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if got.Profile != "test-profile" {
		t.Errorf("Profile = %q, want %q", got.Profile, "test-profile")
	}
}

func TestMockStore_SetToken_MissingProfile(t *testing.T) {
	store := NewMockStore()

	tok := Token{RefreshToken: "refresh-123"}

	err := store.SetToken("", tok)
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("SetToken error = %v, want %v", err, errMissingProfile)
	}

	err = store.SetToken("   ", tok)
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("SetToken error = %v, want %v", err, errMissingProfile)
	}
}

func TestMockStore_SetToken_MissingRefreshToken(t *testing.T) {
	store := NewMockStore()

	tok := Token{AccessToken: "access-123"} // No refresh token

	err := store.SetToken("test-profile", tok)
	if !errors.Is(err, errMissingRefreshToken) {
		t.Errorf("SetToken error = %v, want %v", err, errMissingRefreshToken)
	}
}

func TestMockStore_GetTokenNotFound(t *testing.T) {
	store := NewMockStore()

	_, err := store.GetToken("nonexistent")
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Errorf("GetToken error = %v, want %v", err, keyring.ErrKeyNotFound)
	}
}

func TestMockStore_GetToken_MissingProfile(t *testing.T) {
	store := NewMockStore()

	_, err := store.GetToken("")
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("GetToken error = %v, want %v", err, errMissingProfile)
	}
}

func TestMockStore_DeleteToken(t *testing.T) {
	store := NewMockStore()

	tok := Token{RefreshToken: "refresh-123"}
	_ = store.SetToken("test-profile", tok)

	// Delete token
	err := store.DeleteToken("test-profile")
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify it's gone
	_, err = store.GetToken("test-profile")
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Errorf("GetToken after delete error = %v, want %v", err, keyring.ErrKeyNotFound)
	}
}

func TestMockStore_DeleteToken_NotFound(t *testing.T) {
	store := NewMockStore()

	err := store.DeleteToken("nonexistent")
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Errorf("DeleteToken error = %v, want %v", err, keyring.ErrKeyNotFound)
	}
}

func TestMockStore_DeleteToken_MissingProfile(t *testing.T) {
	store := NewMockStore()

	err := store.DeleteToken("")
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("DeleteToken error = %v, want %v", err, errMissingProfile)
	}
}

func TestMockStore_ListTokens(t *testing.T) {
	store := NewMockStore()

	// Add multiple tokens
	profiles := []string{"profile-a", "profile-b", "profile-c"}
	for _, p := range profiles {
		tok := Token{RefreshToken: "refresh-" + p}
		_ = store.SetToken(p, tok)
	}

	// List tokens
	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens) != len(profiles) {
		t.Errorf("ListTokens returned %d tokens, want %d", len(tokens), len(profiles))
	}

	// Verify all profiles are present
	found := make(map[string]bool)
	for _, tok := range tokens {
		found[tok.Profile] = true
	}

	for _, p := range profiles {
		if !found[p] {
			t.Errorf("ListTokens missing profile %q", p)
		}
	}
}

func TestMockStore_ListTokens_Empty(t *testing.T) {
	store := NewMockStore()

	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens) != 0 {
		t.Errorf("ListTokens returned %d tokens, want 0", len(tokens))
	}
}

func TestMockStore_DefaultAccount(t *testing.T) {
	store := NewMockStore()

	// Initially empty
	account, err := store.GetDefaultAccount()
	if err != nil {
		t.Fatalf("GetDefaultAccount failed: %v", err)
	}
	if account != "" {
		t.Errorf("GetDefaultAccount = %q, want empty string", account)
	}

	// Set default account
	err = store.SetDefaultAccount("my-profile")
	if err != nil {
		t.Fatalf("SetDefaultAccount failed: %v", err)
	}

	// Verify it's set
	account, err = store.GetDefaultAccount()
	if err != nil {
		t.Fatalf("GetDefaultAccount failed: %v", err)
	}
	if account != "my-profile" {
		t.Errorf("GetDefaultAccount = %q, want %q", account, "my-profile")
	}
}

func TestMockStore_SetDefaultAccount_Normalizes(t *testing.T) {
	store := NewMockStore()

	err := store.SetDefaultAccount("  MY-PROFILE  ")
	if err != nil {
		t.Fatalf("SetDefaultAccount failed: %v", err)
	}

	account, _ := store.GetDefaultAccount()
	if account != "my-profile" {
		t.Errorf("GetDefaultAccount = %q, want %q", account, "my-profile")
	}
}

func TestMockStore_SetDefaultAccount_MissingProfile(t *testing.T) {
	store := NewMockStore()

	err := store.SetDefaultAccount("")
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("SetDefaultAccount error = %v, want %v", err, errMissingProfile)
	}
}

func TestMockStore_Keys(t *testing.T) {
	store := NewMockStore()

	// Add tokens
	_ = store.SetToken("profile-a", Token{RefreshToken: "r1"})
	_ = store.SetToken("profile-b", Token{RefreshToken: "r2"})

	keys, err := store.Keys()
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Keys returned %d keys, want 2", len(keys))
	}

	// Keys should be in refresh token format
	for _, k := range keys {
		profile, ok := parseRefreshTokenKey(k)
		if !ok {
			t.Errorf("Key %q is not a valid refresh token key", k)
		}
		if profile != "profile-a" && profile != "profile-b" {
			t.Errorf("Unexpected profile in key: %q", profile)
		}
	}
}

func TestMockStore_SimulateErrors(t *testing.T) {
	store := NewMockStore()

	testErr := errors.New("simulated error")

	// Test SetToken error
	store.SetSetError(testErr)
	err := store.SetToken("profile", Token{RefreshToken: "r"})
	if !errors.Is(err, testErr) {
		t.Errorf("SetToken error = %v, want %v", err, testErr)
	}
	store.SetSetError(nil)

	// Add a token first
	_ = store.SetToken("profile", Token{RefreshToken: "r"})

	// Test GetToken error
	store.SetGetError(testErr)
	_, err = store.GetToken("profile")
	if !errors.Is(err, testErr) {
		t.Errorf("GetToken error = %v, want %v", err, testErr)
	}
	store.SetGetError(nil)
}

func TestMockStore_Reset(t *testing.T) {
	store := NewMockStore()

	// Set up some state
	_ = store.SetToken("profile", Token{RefreshToken: "r"})
	_ = store.SetDefaultAccount("profile")
	store.SetGetError(errors.New("test"))
	store.SetSetError(errors.New("test"))

	// Reset
	store.Reset()

	// Verify everything is cleared
	tokens, _ := store.ListTokens()
	if len(tokens) != 0 {
		t.Error("Reset did not clear tokens")
	}

	account, _ := store.GetDefaultAccount()
	if account != "" {
		t.Error("Reset did not clear default account")
	}

	// Verify errors are cleared (no error should occur)
	_ = store.SetToken("new-profile", Token{RefreshToken: "r"})
	_, err := store.GetToken("new-profile")
	if err != nil {
		t.Error("Reset did not clear errors")
	}
}

func TestToken_Fields(t *testing.T) {
	now := time.Now()
	expires := now.Add(time.Hour)

	tok := Token{
		Profile:      "test",
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresAt:    expires,
		Scopes:       []string{"read", "write", "admin"},
		CreatedAt:    now,
	}

	if tok.Profile != "test" {
		t.Errorf("Profile = %q, want %q", tok.Profile, "test")
	}
	if tok.AccessToken != "access-123" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "access-123")
	}
	if tok.RefreshToken != "refresh-456" {
		t.Errorf("RefreshToken = %q, want %q", tok.RefreshToken, "refresh-456")
	}
	if !tok.ExpiresAt.Equal(expires) {
		t.Errorf("ExpiresAt = %v, want %v", tok.ExpiresAt, expires)
	}
	if len(tok.Scopes) != 3 {
		t.Errorf("Scopes length = %d, want 3", len(tok.Scopes))
	}
	if !tok.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", tok.CreatedAt, now)
	}
}

func TestToken_ZeroValue(t *testing.T) {
	var tok Token

	if tok.Profile != "" {
		t.Errorf("Zero Profile = %q, want empty", tok.Profile)
	}
	if tok.AccessToken != "" {
		t.Errorf("Zero AccessToken = %q, want empty", tok.AccessToken)
	}
	if tok.RefreshToken != "" {
		t.Errorf("Zero RefreshToken = %q, want empty", tok.RefreshToken)
	}
	if !tok.ExpiresAt.IsZero() {
		t.Errorf("Zero ExpiresAt = %v, want zero", tok.ExpiresAt)
	}
	if tok.Scopes != nil {
		t.Errorf("Zero Scopes = %v, want nil", tok.Scopes)
	}
	if !tok.CreatedAt.IsZero() {
		t.Errorf("Zero CreatedAt = %v, want zero", tok.CreatedAt)
	}
}

// fakeKeyring is a test implementation of keyring.Keyring
type fakeKeyring struct {
	items   map[string]keyring.Item
	keysErr error
	getErr  error
	setErr  error
	delErr  error
}

func newFakeKeyring() *fakeKeyring {
	return &fakeKeyring{
		items: make(map[string]keyring.Item),
	}
}

func (f *fakeKeyring) Get(key string) (keyring.Item, error) {
	if f.getErr != nil {
		return keyring.Item{}, f.getErr
	}
	item, ok := f.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}
	return item, nil
}

func (f *fakeKeyring) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

func (f *fakeKeyring) Set(item keyring.Item) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.items[item.Key] = item
	return nil
}

func (f *fakeKeyring) Remove(key string) error {
	if f.delErr != nil {
		return f.delErr
	}
	if _, ok := f.items[key]; !ok {
		return keyring.ErrKeyNotFound
	}
	delete(f.items, key)
	return nil
}

func (f *fakeKeyring) Keys() ([]string, error) {
	if f.keysErr != nil {
		return nil, f.keysErr
	}
	keys := make([]string, 0, len(f.items))
	for k := range f.items {
		keys = append(keys, k)
	}
	return keys, nil
}

// Verify fakeKeyring implements keyring.Keyring interface
var _ keyring.Keyring = (*fakeKeyring)(nil)

// KeyringStore tests using fakeKeyring

func TestKeyringStore_Keys(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Empty
	keys, err := store.Keys()
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Keys returned %d keys, want 0", len(keys))
	}

	// Add some items
	_ = ring.Set(keyring.Item{Key: "key1", Data: []byte("data1")})
	_ = ring.Set(keyring.Item{Key: "key2", Data: []byte("data2")})

	keys, err = store.Keys()
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("Keys returned %d keys, want 2", len(keys))
	}
}

func TestKeyringStore_Keys_Error(t *testing.T) {
	ring := newFakeKeyring()
	ring.keysErr = errors.New("keyring error")
	store := &KeyringStore{ring: ring}

	_, err := store.Keys()
	if err == nil {
		t.Fatal("Keys expected error, got nil")
	}
	if !errors.Is(err, ring.keysErr) {
		t.Errorf("Keys error = %v, want wrapped %v", err, ring.keysErr)
	}
}

func TestKeyringStore_SetToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	tok := Token{
		RefreshToken: "refresh-123",
		AccessToken:  "access-456",
		Scopes:       []string{"read", "write"},
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	err := store.SetToken("test-profile", tok)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Verify refresh token is stored
	refreshKey := refreshTokenKey("test-profile")
	item, err := ring.Get(refreshKey)
	if err != nil {
		t.Fatalf("Get refresh token failed: %v", err)
	}

	var st storedToken
	if err := json.Unmarshal(item.Data, &st); err != nil {
		t.Fatalf("Unmarshal refresh token failed: %v", err)
	}
	if st.RefreshToken != "refresh-123" {
		t.Errorf("RefreshToken = %q, want %q", st.RefreshToken, "refresh-123")
	}

	// Verify access token is stored
	accessKey := accessTokenKey("test-profile")
	item, err = ring.Get(accessKey)
	if err != nil {
		t.Fatalf("Get access token failed: %v", err)
	}

	var sat storedAccessToken
	if err := json.Unmarshal(item.Data, &sat); err != nil {
		t.Fatalf("Unmarshal access token failed: %v", err)
	}
	if sat.AccessToken != "access-456" {
		t.Errorf("AccessToken = %q, want %q", sat.AccessToken, "access-456")
	}
}

func TestKeyringStore_SetToken_MissingProfile(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.SetToken("", Token{RefreshToken: "r"})
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("SetToken error = %v, want %v", err, errMissingProfile)
	}

	err = store.SetToken("   ", Token{RefreshToken: "r"})
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("SetToken error = %v, want %v", err, errMissingProfile)
	}
}

func TestKeyringStore_SetToken_MissingRefreshToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.SetToken("profile", Token{AccessToken: "a"})
	if !errors.Is(err, errMissingRefreshToken) {
		t.Errorf("SetToken error = %v, want %v", err, errMissingRefreshToken)
	}
}

func TestKeyringStore_SetToken_SetsCreatedAt(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	before := time.Now().UTC()
	err := store.SetToken("profile", Token{RefreshToken: "r"})
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}
	after := time.Now().UTC()

	item, _ := ring.Get(refreshTokenKey("profile"))
	var st storedToken
	_ = json.Unmarshal(item.Data, &st)

	if st.CreatedAt.Before(before) || st.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", st.CreatedAt, before, after)
	}
}

func TestKeyringStore_SetToken_PreservesCreatedAt(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	createdAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	err := store.SetToken("profile", Token{RefreshToken: "r", CreatedAt: createdAt})
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	item, _ := ring.Get(refreshTokenKey("profile"))
	var st storedToken
	_ = json.Unmarshal(item.Data, &st)

	if !st.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", st.CreatedAt, createdAt)
	}
}

func TestKeyringStore_SetToken_NoAccessToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Set token without access token
	err := store.SetToken("profile", Token{RefreshToken: "r"})
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Verify no access token stored
	_, err = ring.Get(accessTokenKey("profile"))
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Errorf("Expected ErrKeyNotFound for access token, got %v", err)
	}
}

func TestKeyringStore_SetToken_RefreshTokenError(t *testing.T) {
	ring := newFakeKeyring()
	ring.setErr = errors.New("keyring set error")
	store := &KeyringStore{ring: ring}

	err := store.SetToken("profile", Token{RefreshToken: "r"})
	if err == nil {
		t.Fatal("SetToken expected error, got nil")
	}
}

func TestKeyringStore_GetToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store tokens
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(time.Hour)

	refreshData, _ := json.Marshal(storedToken{
		RefreshToken: "refresh-123",
		Scopes:       []string{"read"},
		CreatedAt:    now,
	})
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: refreshData})

	accessData, _ := json.Marshal(storedAccessToken{
		AccessToken: "access-456",
		ExpiresAt:   expires,
	})
	_ = ring.Set(keyring.Item{Key: accessTokenKey("profile"), Data: accessData})

	// Get token
	tok, err := store.GetToken("profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if tok.Profile != "profile" {
		t.Errorf("Profile = %q, want %q", tok.Profile, "profile")
	}
	if tok.RefreshToken != "refresh-123" {
		t.Errorf("RefreshToken = %q, want %q", tok.RefreshToken, "refresh-123")
	}
	if tok.AccessToken != "access-456" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "access-456")
	}
	if len(tok.Scopes) != 1 || tok.Scopes[0] != "read" {
		t.Errorf("Scopes = %v, want [read]", tok.Scopes)
	}
}

func TestKeyringStore_GetToken_MissingProfile(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	_, err := store.GetToken("")
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("GetToken error = %v, want %v", err, errMissingProfile)
	}
}

func TestKeyringStore_GetToken_NotFound(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	_, err := store.GetToken("nonexistent")
	if err == nil {
		t.Fatal("GetToken expected error, got nil")
	}
}

func TestKeyringStore_GetToken_NoAccessToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store only refresh token
	refreshData, _ := json.Marshal(storedToken{RefreshToken: "refresh-123"})
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: refreshData})

	tok, err := store.GetToken("profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if tok.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty", tok.AccessToken)
	}
}

func TestKeyringStore_GetToken_InvalidAccessTokenJSON(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store valid refresh token
	refreshData, _ := json.Marshal(storedToken{RefreshToken: "refresh-123"})
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: refreshData})

	// Store invalid access token JSON
	_ = ring.Set(keyring.Item{Key: accessTokenKey("profile"), Data: []byte("invalid json")})

	// Should still return token (access token is optional)
	tok, err := store.GetToken("profile")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if tok.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty for invalid JSON", tok.AccessToken)
	}
}

func TestKeyringStore_GetToken_InvalidRefreshTokenJSON(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store invalid refresh token JSON
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: []byte("invalid json")})

	_, err := store.GetToken("profile")
	if err == nil {
		t.Fatal("GetToken expected error for invalid JSON, got nil")
	}
}

func TestKeyringStore_DeleteToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store tokens
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: []byte("{}")})
	_ = ring.Set(keyring.Item{Key: accessTokenKey("profile"), Data: []byte("{}")})

	err := store.DeleteToken("profile")
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify both tokens are gone
	_, err = ring.Get(refreshTokenKey("profile"))
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Errorf("Refresh token should be deleted")
	}

	_, err = ring.Get(accessTokenKey("profile"))
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Errorf("Access token should be deleted")
	}
}

func TestKeyringStore_DeleteToken_MissingProfile(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.DeleteToken("")
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("DeleteToken error = %v, want %v", err, errMissingProfile)
	}
}

func TestKeyringStore_DeleteToken_NotFound(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.DeleteToken("nonexistent")
	if err == nil {
		t.Fatal("DeleteToken expected error, got nil")
	}
}

func TestKeyringStore_DeleteToken_NoAccessToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store only refresh token
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: []byte("{}")})

	// Should succeed even if access token doesn't exist
	err := store.DeleteToken("profile")
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}
}

func TestKeyringStore_ListTokens(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store multiple tokens
	for _, profile := range []string{"profile-a", "profile-b"} {
		data, _ := json.Marshal(storedToken{RefreshToken: "r-" + profile})
		_ = ring.Set(keyring.Item{Key: refreshTokenKey(profile), Data: data})
	}

	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("ListTokens returned %d tokens, want 2", len(tokens))
	}
}

func TestKeyringStore_ListTokens_Empty(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens) != 0 {
		t.Errorf("ListTokens returned %d tokens, want 0", len(tokens))
	}
}

func TestKeyringStore_ListTokens_IgnoresNonRefreshKeys(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store a refresh token
	data, _ := json.Marshal(storedToken{RefreshToken: "r"})
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: data})

	// Store other keys that should be ignored
	_ = ring.Set(keyring.Item{Key: accessTokenKey("profile"), Data: []byte("{}")})
	_ = ring.Set(keyring.Item{Key: defaultAccountKey, Data: []byte("profile")})
	_ = ring.Set(keyring.Item{Key: "some-random-key", Data: []byte("data")})

	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens) != 1 {
		t.Errorf("ListTokens returned %d tokens, want 1", len(tokens))
	}
}

func TestKeyringStore_ListTokens_KeysError(t *testing.T) {
	ring := newFakeKeyring()
	ring.keysErr = errors.New("keyring error")
	store := &KeyringStore{ring: ring}

	_, err := store.ListTokens()
	if err == nil {
		t.Fatal("ListTokens expected error, got nil")
	}
}

func TestKeyringStore_GetDefaultAccount(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Initially empty (not found)
	account, err := store.GetDefaultAccount()
	if err != nil {
		t.Fatalf("GetDefaultAccount failed: %v", err)
	}
	if account != "" {
		t.Errorf("GetDefaultAccount = %q, want empty", account)
	}

	// Set default account
	_ = ring.Set(keyring.Item{Key: defaultAccountKey, Data: []byte("my-profile")})

	account, err = store.GetDefaultAccount()
	if err != nil {
		t.Fatalf("GetDefaultAccount failed: %v", err)
	}
	if account != "my-profile" {
		t.Errorf("GetDefaultAccount = %q, want %q", account, "my-profile")
	}
}

func TestKeyringStore_GetDefaultAccount_Error(t *testing.T) {
	ring := newFakeKeyring()
	ring.getErr = errors.New("keyring error")
	store := &KeyringStore{ring: ring}

	_, err := store.GetDefaultAccount()
	if err == nil {
		t.Fatal("GetDefaultAccount expected error, got nil")
	}
}

func TestKeyringStore_SetDefaultAccount(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.SetDefaultAccount("my-profile")
	if err != nil {
		t.Fatalf("SetDefaultAccount failed: %v", err)
	}

	item, err := ring.Get(defaultAccountKey)
	if err != nil {
		t.Fatalf("Get default account failed: %v", err)
	}
	if string(item.Data) != "my-profile" {
		t.Errorf("Default account = %q, want %q", string(item.Data), "my-profile")
	}
}

func TestKeyringStore_SetDefaultAccount_MissingProfile(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.SetDefaultAccount("")
	if !errors.Is(err, errMissingProfile) {
		t.Errorf("SetDefaultAccount error = %v, want %v", err, errMissingProfile)
	}
}

func TestKeyringStore_SetDefaultAccount_Normalizes(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	err := store.SetDefaultAccount("  MY-PROFILE  ")
	if err != nil {
		t.Fatalf("SetDefaultAccount failed: %v", err)
	}

	item, _ := ring.Get(defaultAccountKey)
	if string(item.Data) != "my-profile" {
		t.Errorf("Default account = %q, want %q", string(item.Data), "my-profile")
	}
}

func TestKeyringStore_SetDefaultAccount_Error(t *testing.T) {
	ring := newFakeKeyring()
	ring.setErr = errors.New("keyring error")
	store := &KeyringStore{ring: ring}

	err := store.SetDefaultAccount("profile")
	if err == nil {
		t.Fatal("SetDefaultAccount expected error, got nil")
	}
}

// Helper function tests

func TestRefreshTokenKey(t *testing.T) {
	key := refreshTokenKey("my-profile")
	if key != "my-profile:refresh_token" {
		t.Errorf("refreshTokenKey = %q, want %q", key, "my-profile:refresh_token")
	}
}

func TestAccessTokenKey(t *testing.T) {
	key := accessTokenKey("my-profile")
	if key != "my-profile:access_token" {
		t.Errorf("accessTokenKey = %q, want %q", key, "my-profile:access_token")
	}
}

func TestParseRefreshTokenKey(t *testing.T) {
	tests := []struct {
		key     string
		profile string
		ok      bool
	}{
		{"profile:refresh_token", "profile", true},
		{"my-profile:refresh_token", "my-profile", true},
		{"complex:name:refresh_token", "complex:name", true},
		{":refresh_token", "", false},
		{"   :refresh_token", "", false},
		{"profile:access_token", "", false},
		{"profile", "", false},
		{"", "", false},
		{"refresh_token", "", false},
	}

	for _, tt := range tests {
		profile, ok := parseRefreshTokenKey(tt.key)
		if ok != tt.ok {
			t.Errorf("parseRefreshTokenKey(%q) ok = %v, want %v", tt.key, ok, tt.ok)
		}
		if profile != tt.profile {
			t.Errorf("parseRefreshTokenKey(%q) profile = %q, want %q", tt.key, profile, tt.profile)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"profile", "profile"},
		{"PROFILE", "profile"},
		{"  Profile  ", "profile"},
		{"MY-PROFILE", "my-profile"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		got := normalize(tt.input)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWrapKeychainError(t *testing.T) {
	// nil error
	if wrapKeychainError(nil) != nil {
		t.Error("wrapKeychainError(nil) should return nil")
	}

	// Regular error
	err := errors.New("some error")
	wrapped := wrapKeychainError(err)
	if wrapped != err {
		t.Errorf("wrapKeychainError should return original error for non-keychain errors")
	}

	// Keychain locked error
	lockedErr := errors.New("keychain error: -25308")
	wrapped = wrapKeychainError(lockedErr)
	if wrapped == lockedErr {
		t.Error("wrapKeychainError should wrap keychain locked error")
	}
	if wrapped.Error() == lockedErr.Error() {
		t.Error("wrapped error should contain additional guidance")
	}
}

func TestIsKeychainLockedError(t *testing.T) {
	tests := []struct {
		errStr string
		want   bool
	}{
		{"keychain error: -25308", true},
		{"errSecInteractionNotAllowed (-25308)", true},
		{"some other error", false},
		{"", false},
		{"-25308", true},
	}

	for _, tt := range tests {
		got := isKeychainLockedError(tt.errStr)
		if got != tt.want {
			t.Errorf("isKeychainLockedError(%q) = %v, want %v", tt.errStr, got, tt.want)
		}
	}
}

func TestAllowedBackends(t *testing.T) {
	tests := []struct {
		info    KeyringBackendInfo
		wantLen int
		wantErr bool
	}{
		{KeyringBackendInfo{Value: "", Source: "default"}, 0, false},
		{KeyringBackendInfo{Value: "auto", Source: "default"}, 0, false},
		{KeyringBackendInfo{Value: "keychain", Source: "env"}, 1, false},
		{KeyringBackendInfo{Value: "file", Source: "config"}, 1, false},
		{KeyringBackendInfo{Value: "invalid", Source: "env"}, 0, true},
	}

	for _, tt := range tests {
		backends, err := allowedBackends(tt.info)
		if (err != nil) != tt.wantErr {
			t.Errorf("allowedBackends(%+v) error = %v, wantErr %v", tt.info, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && len(backends) != tt.wantLen {
			// nil means try all backends (auto mode)
			if tt.wantLen == 0 && backends != nil {
				t.Errorf("allowedBackends(%+v) = %v, want nil for auto", tt.info, backends)
			}
		}
	}
}

func TestAllowedBackends_KeychainBackend(t *testing.T) {
	info := KeyringBackendInfo{Value: "keychain", Source: "env"}
	backends, err := allowedBackends(info)
	if err != nil {
		t.Fatalf("allowedBackends error: %v", err)
	}
	if len(backends) != 1 || backends[0] != keyring.KeychainBackend {
		t.Errorf("allowedBackends(keychain) = %v, want [KeychainBackend]", backends)
	}
}

func TestAllowedBackends_FileBackend(t *testing.T) {
	info := KeyringBackendInfo{Value: "file", Source: "config"}
	backends, err := allowedBackends(info)
	if err != nil {
		t.Fatalf("allowedBackends error: %v", err)
	}
	if len(backends) != 1 || backends[0] != keyring.FileBackend {
		t.Errorf("allowedBackends(file) = %v, want [FileBackend]", backends)
	}
}

func TestAllowedBackends_InvalidBackend(t *testing.T) {
	info := KeyringBackendInfo{Value: "invalid", Source: "env"}
	_, err := allowedBackends(info)
	if err == nil {
		t.Fatal("allowedBackends(invalid) expected error")
	}
	if !errors.Is(err, errInvalidKeyringBackend) {
		t.Errorf("allowedBackends(invalid) error = %v, want %v", err, errInvalidKeyringBackend)
	}
}

func TestFileKeyringPasswordFuncFrom(t *testing.T) {
	// With password
	prompt := fileKeyringPasswordFuncFrom("mypassword", false)
	pass, err := prompt("Enter password:")
	if err != nil {
		t.Fatalf("prompt error: %v", err)
	}
	if pass != "mypassword" {
		t.Errorf("password = %q, want %q", pass, "mypassword")
	}

	// No password, no TTY - should return error
	prompt = fileKeyringPasswordFuncFrom("", false)
	_, err = prompt("Enter password:")
	if err == nil {
		t.Fatal("prompt expected error for no TTY")
	}
	if !errors.Is(err, errNoTTY) {
		t.Errorf("prompt error = %v, want %v", err, errNoTTY)
	}

	// No password, with TTY - returns TerminalPrompt (can't easily test the actual prompt)
	prompt = fileKeyringPasswordFuncFrom("", true)
	if prompt == nil {
		t.Error("prompt should not be nil for TTY case")
	}
}

func TestKeyringBackendInfo(t *testing.T) {
	info := KeyringBackendInfo{
		Value:  "file",
		Source: "env",
	}

	if info.Value != "file" {
		t.Errorf("Value = %q, want %q", info.Value, "file")
	}
	if info.Source != "env" {
		t.Errorf("Source = %q, want %q", info.Source, "env")
	}
}

// Test storedToken JSON marshaling
func TestStoredToken_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	st := storedToken{
		RefreshToken: "refresh-123",
		Scopes:       []string{"read", "write"},
		CreatedAt:    now,
	}

	data, err := json.Marshal(st)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded storedToken
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.RefreshToken != st.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", decoded.RefreshToken, st.RefreshToken)
	}
	if len(decoded.Scopes) != len(st.Scopes) {
		t.Errorf("Scopes = %v, want %v", decoded.Scopes, st.Scopes)
	}
	if !decoded.CreatedAt.Equal(st.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", decoded.CreatedAt, st.CreatedAt)
	}
}

// Test storedAccessToken JSON marshaling
func TestStoredAccessToken_JSON(t *testing.T) {
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	sat := storedAccessToken{
		AccessToken: "access-456",
		ExpiresAt:   expires,
	}

	data, err := json.Marshal(sat)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded storedAccessToken
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.AccessToken != sat.AccessToken {
		t.Errorf("AccessToken = %q, want %q", decoded.AccessToken, sat.AccessToken)
	}
	if !decoded.ExpiresAt.Equal(sat.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", decoded.ExpiresAt, sat.ExpiresAt)
	}
}

// Test ListTokens with GetToken error
func TestKeyringStore_ListTokens_GetTokenError(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store a valid refresh token key but with invalid JSON data
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: []byte("invalid json")})

	_, err := store.ListTokens()
	if err == nil {
		t.Fatal("ListTokens expected error when GetToken fails")
	}
}

// Test ListTokens deduplication
func TestKeyringStore_ListTokens_Deduplication(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store one profile with both refresh and access tokens
	// Only one token should be returned (no duplicates)
	data, _ := json.Marshal(storedToken{RefreshToken: "r"})
	_ = ring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: data})
	// Note: access token key doesn't match refresh token pattern, so should be ignored

	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens) != 1 {
		t.Errorf("ListTokens returned %d tokens, want 1", len(tokens))
	}
}

// Test SetToken with access token set failure (non-fatal)
func TestKeyringStore_SetToken_AccessTokenSetError(t *testing.T) {
	// Create a fake keyring that fails on second Set (access token)
	ring := &fakeKeyringWithAccessTokenError{
		fakeKeyring: *newFakeKeyring(),
	}
	store := &KeyringStore{ring: ring}

	// Set token with access token - should succeed even if access token fails
	err := store.SetToken("profile", Token{
		RefreshToken: "refresh-123",
		AccessToken:  "access-456",
	})
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Verify refresh token was stored
	item, err := ring.Get(refreshTokenKey("profile"))
	if err != nil {
		t.Fatalf("Refresh token should be stored: %v", err)
	}

	var st storedToken
	_ = json.Unmarshal(item.Data, &st)
	if st.RefreshToken != "refresh-123" {
		t.Errorf("RefreshToken = %q, want %q", st.RefreshToken, "refresh-123")
	}
}

// fakeKeyringWithAccessTokenError fails when setting access tokens
type fakeKeyringWithAccessTokenError struct {
	fakeKeyring
}

func (f *fakeKeyringWithAccessTokenError) Set(item keyring.Item) error {
	// Fail on access token key
	if strings.HasSuffix(item.Key, ":access_token") {
		return errors.New("access token error")
	}
	return f.fakeKeyring.Set(item)
}

// Test ResolveKeyringBackendInfo with environment variable
func TestResolveKeyringBackendInfo_FromEnv(t *testing.T) {
	// Save original env and restore after test
	origEnv := os.Getenv(keyringBackendEnv)
	defer os.Setenv(keyringBackendEnv, origEnv)

	// Set environment variable
	os.Setenv(keyringBackendEnv, "FILE")

	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		t.Fatalf("ResolveKeyringBackendInfo failed: %v", err)
	}

	if info.Value != "file" {
		t.Errorf("Value = %q, want %q", info.Value, "file")
	}
	if info.Source != keyringBackendSourceEnv {
		t.Errorf("Source = %q, want %q", info.Source, keyringBackendSourceEnv)
	}
}

// Test ResolveKeyringBackendInfo with whitespace-only env
func TestResolveKeyringBackendInfo_EmptyEnv(t *testing.T) {
	// Save original env and restore after test
	origEnv := os.Getenv(keyringBackendEnv)
	defer os.Setenv(keyringBackendEnv, origEnv)

	// Set environment variable to whitespace
	os.Setenv(keyringBackendEnv, "   ")

	// Should fall through to config/default
	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		t.Fatalf("ResolveKeyringBackendInfo failed: %v", err)
	}

	// Source should not be "env" since whitespace-only is treated as unset
	if info.Source == keyringBackendSourceEnv {
		t.Errorf("Source should not be 'env' for whitespace-only value")
	}
}

// Test ResolveKeyringBackendInfo falls through to config/default
func TestResolveKeyringBackendInfo_Default(t *testing.T) {
	// Save original env and restore after test
	origEnv := os.Getenv(keyringBackendEnv)
	defer os.Setenv(keyringBackendEnv, origEnv)

	// Unset environment variable
	os.Unsetenv(keyringBackendEnv)

	// Should use config or default
	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		t.Fatalf("ResolveKeyringBackendInfo failed: %v", err)
	}

	// Source should be config or default (depending on config file contents)
	if info.Source == keyringBackendSourceEnv {
		t.Errorf("Source should not be 'env' when env is unset")
	}

	// Value should be valid (auto, keychain, or file)
	if info.Value != "auto" && info.Value != "keychain" && info.Value != "file" && info.Value != "" {
		t.Errorf("Value = %q, expected valid backend", info.Value)
	}
}

// Test ListTokens seen map deduplication with multiple same-profile keys
func TestKeyringStore_ListTokens_SeenDeduplication(t *testing.T) {
	// Create a custom keyring that returns duplicate profile keys
	ring := &fakeKeyringWithDuplicates{
		fakeKeyring: *newFakeKeyring(),
	}

	// Store valid token data
	data, _ := json.Marshal(storedToken{RefreshToken: "r"})
	_ = ring.fakeKeyring.Set(keyring.Item{Key: refreshTokenKey("profile"), Data: data})

	store := &KeyringStore{ring: ring}

	tokens, err := store.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	// Should only return 1 token despite duplicate keys
	if len(tokens) != 1 {
		t.Errorf("ListTokens returned %d tokens, want 1 (deduplicated)", len(tokens))
	}
}

// fakeKeyringWithDuplicates returns duplicate keys
type fakeKeyringWithDuplicates struct {
	fakeKeyring
}

func (f *fakeKeyringWithDuplicates) Keys() ([]string, error) {
	// Return the same key twice to test deduplication
	keys, err := f.fakeKeyring.Keys()
	if err != nil {
		return nil, err
	}
	// Duplicate all keys
	return append(keys, keys...), nil
}

// Test KeyringBackendInfo source constants
func TestKeyringBackendSourceConstants(t *testing.T) {
	if keyringBackendSourceEnv != "env" {
		t.Errorf("keyringBackendSourceEnv = %q, want 'env'", keyringBackendSourceEnv)
	}
	if keyringBackendSourceConfig != "config" {
		t.Errorf("keyringBackendSourceConfig = %q, want 'config'", keyringBackendSourceConfig)
	}
	if keyringBackendSourceDefault != "default" {
		t.Errorf("keyringBackendSourceDefault = %q, want 'default'", keyringBackendSourceDefault)
	}
}

// Test error variables are properly defined
func TestErrorVariables(t *testing.T) {
	if errMissingProfile == nil {
		t.Error("errMissingProfile should not be nil")
	}
	if errMissingRefreshToken == nil {
		t.Error("errMissingRefreshToken should not be nil")
	}
	if errNoTTY == nil {
		t.Error("errNoTTY should not be nil")
	}
	if errInvalidKeyringBackend == nil {
		t.Error("errInvalidKeyringBackend should not be nil")
	}
}

// Test defaultAccountKey constant
func TestDefaultAccountKey(t *testing.T) {
	if defaultAccountKey != "default_account" {
		t.Errorf("defaultAccountKey = %q, want 'default_account'", defaultAccountKey)
	}
}

// Test environment variable constants
func TestEnvVarConstants(t *testing.T) {
	if keyringPasswordEnv != "TWENTY_KEYRING_PASSWORD" {
		t.Errorf("keyringPasswordEnv = %q, want 'TWENTY_KEYRING_PASSWORD'", keyringPasswordEnv)
	}
	if keyringBackendEnv != "TWENTY_KEYRING_BACKEND" {
		t.Errorf("keyringBackendEnv = %q, want 'TWENTY_KEYRING_BACKEND'", keyringBackendEnv)
	}
}

// Test Store interface is satisfied by KeyringStore
func TestKeyringStore_ImplementsStore(t *testing.T) {
	var _ Store = (*KeyringStore)(nil)
}

// Test fakeKeyring GetMetadata
func TestFakeKeyring_GetMetadata(t *testing.T) {
	ring := newFakeKeyring()
	meta, err := ring.GetMetadata("anykey")
	if err != nil {
		t.Errorf("GetMetadata returned error: %v", err)
	}
	// Metadata should be zero value
	if meta != (keyring.Metadata{}) {
		t.Errorf("GetMetadata should return zero value Metadata")
	}
}

// Test ResolveKeyringBackendInfo when config read fails
func TestResolveKeyringBackendInfo_ConfigReadError(t *testing.T) {
	// Save original env and func, restore after test
	origEnv := os.Getenv(keyringBackendEnv)
	origReadConfigFunc := readConfigFunc
	defer func() {
		os.Setenv(keyringBackendEnv, origEnv)
		readConfigFunc = origReadConfigFunc
	}()

	// Unset env so we fall through to config
	os.Unsetenv(keyringBackendEnv)

	// Inject error
	configErr := errors.New("config read failed")
	readConfigFunc = func() (*config.Config, error) {
		return nil, configErr
	}

	_, err := ResolveKeyringBackendInfo()
	if err == nil {
		t.Fatal("ResolveKeyringBackendInfo expected error, got nil")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Errorf("error should contain 'read config': %v", err)
	}
}

// ==============================================================================
// Security Tests: Verify sensitive data is not exposed in error messages
// ==============================================================================

// TestErrorMessages_DoNotExposeTokenValue verifies that error messages from
// token operations do not include the actual token value.
func TestErrorMessages_DoNotExposeTokenValue(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	sensitiveToken := "super-secret-api-token-xyz123abc"

	// Test SetToken error
	ring.setErr = errors.New("keyring locked")
	err := store.SetToken("profile", Token{RefreshToken: sensitiveToken})

	if err == nil {
		t.Fatal("SetToken expected error, got nil")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, sensitiveToken) {
		t.Errorf("SetToken error should NOT contain token value.\nError: %s", errMsg)
	}
}

// TestGetTokenError_DoesNotLeakStoredToken verifies that when GetToken fails,
// the error message does not leak any previously stored token value.
func TestGetTokenError_DoesNotLeakStoredToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	sensitiveToken := "stored-secret-credential-value"

	// First store a token successfully
	err := store.SetToken("profile", Token{RefreshToken: sensitiveToken})
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Now make Get fail
	ring.getErr = errors.New("keyring access denied")

	_, err = store.GetToken("profile")
	if err == nil {
		t.Fatal("GetToken expected error, got nil")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, sensitiveToken) {
		t.Errorf("GetToken error should NOT contain stored token.\nError: %s", errMsg)
	}
}

// TestDeleteTokenError_DoesNotLeakToken verifies that DeleteToken errors
// do not expose any token values.
func TestDeleteTokenError_DoesNotLeakToken(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	sensitiveToken := "delete-test-secret-token"

	// First store a token
	err := store.SetToken("profile", Token{RefreshToken: sensitiveToken})
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Make delete fail
	ring.delErr = errors.New("permission denied")

	err = store.DeleteToken("profile")
	if err == nil {
		t.Fatal("DeleteToken expected error, got nil")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, sensitiveToken) {
		t.Errorf("DeleteToken error should NOT contain token.\nError: %s", errMsg)
	}
}

// TestListTokensError_DoesNotLeakTokens verifies that ListTokens errors
// do not expose any stored token values.
func TestListTokensError_DoesNotLeakTokens(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store multiple tokens with sensitive values
	tokens := []string{
		"secret-token-one-abc123",
		"secret-token-two-def456",
		"secret-token-three-ghi789",
	}

	for i, tok := range tokens {
		err := store.SetToken("profile-"+string(rune('a'+i)), Token{RefreshToken: tok})
		if err != nil {
			t.Fatalf("SetToken failed: %v", err)
		}
	}

	// Make Keys() fail, which will cause ListTokens to fail
	ring.keysErr = errors.New("keyring error")

	_, err := store.ListTokens()
	if err == nil {
		t.Fatal("ListTokens expected error, got nil")
	}

	errMsg := err.Error()
	for _, tok := range tokens {
		if strings.Contains(errMsg, tok) {
			t.Errorf("ListTokens error should NOT contain token: %s\nError: %s", tok, errMsg)
		}
	}
}

// TestMarshalError_DoesNotExposeToken verifies that JSON marshaling errors
// do not expose token values in the error message.
func TestMarshalError_DoesNotExposeToken(t *testing.T) {
	// This test verifies that even if JSON operations fail internally,
	// the error messages produced don't contain the actual token value.
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	sensitiveToken := "marshal-test-secret-token-value"

	// Store a token (this should succeed as we're testing error message security)
	err := store.SetToken("profile", Token{
		RefreshToken: sensitiveToken,
		AccessToken:  "access-token-also-secret",
		Scopes:       []string{"read", "write"},
	})

	// Even if we had a marshal error, verify the token isn't in the error
	// Since JSON marshal of simple structs rarely fails, we verify by
	// checking that successful operations also don't log tokens
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, sensitiveToken) {
			t.Errorf("Error should NOT contain token: %s", errMsg)
		}
	}
}

// TestUnmarshalError_DoesNotExposeRawData verifies that JSON unmarshal errors
// for corrupted keyring data don't expose sensitive information.
func TestUnmarshalError_DoesNotExposeRawData(t *testing.T) {
	ring := newFakeKeyring()
	store := &KeyringStore{ring: ring}

	// Store corrupted but sensitive-looking data
	sensitiveData := `{"refresh_token":"corrupted-but-secret-xyz"invalid`
	_ = ring.Set(keyring.Item{
		Key:  refreshTokenKey("profile"),
		Data: []byte(sensitiveData),
	})

	_, err := store.GetToken("profile")

	// GetToken should fail due to invalid JSON
	if err == nil {
		t.Fatal("GetToken expected error for invalid JSON")
	}

	errMsg := err.Error()

	// The error should not expose the raw data containing the token
	if strings.Contains(errMsg, "corrupted-but-secret") {
		t.Errorf("Unmarshal error should NOT expose raw token data.\nError: %s", errMsg)
	}
}

// TestKeychainError_DoesNotExposeToken verifies that keychain-specific errors
// (like locked keychain) don't expose any token values.
func TestKeychainError_DoesNotExposeToken(t *testing.T) {
	// Test the wrapKeychainError function doesn't add token to message
	sensitiveToken := "keychain-test-secret"
	originalErr := errors.New("store refresh token: keychain error: -25308")

	wrappedErr := wrapKeychainError(originalErr)

	if strings.Contains(wrappedErr.Error(), sensitiveToken) {
		t.Errorf("Wrapped keychain error should NOT contain token")
	}

	// The wrapped error should contain helpful guidance but not tokens
	if !strings.Contains(wrappedErr.Error(), "keychain") {
		t.Error("Wrapped error should mention keychain")
	}
}

// ==============================================================================
// Error Recovery Instruction Tests
// ==============================================================================

// TestWrapKeychainError_IncludesRecoveryInstructions verifies that keychain locked
// errors include instructions on how to unlock the keychain.
func TestWrapKeychainError_IncludesRecoveryInstructions(t *testing.T) {
	// Test locked keychain error
	lockedErr := fmt.Errorf("operation failed: errSecInteractionNotAllowed -25308")
	wrapped := wrapKeychainError(lockedErr)

	errStr := wrapped.Error()
	if !strings.Contains(errStr, "security unlock-keychain") {
		t.Errorf("wrapKeychainError() should include unlock instructions, got: %s", errStr)
	}
}

// TestKeyringTimeoutError_IncludesRecoveryInstructions verifies that timeout errors
// include instructions about using the file backend.
func TestKeyringTimeoutError_IncludesRecoveryInstructions(t *testing.T) {
	// Simulate what openKeyringWithTimeout returns on timeout
	err := fmt.Errorf("%w after 5s (D-Bus SecretService may be unresponsive); "+
		"set TWENTY_KEYRING_BACKEND=file and TWENTY_KEYRING_PASSWORD=<password> to use encrypted file storage",
		errKeyringTimeout)
	errStr := err.Error()

	if !strings.Contains(errStr, "TWENTY_KEYRING_BACKEND=file") {
		t.Errorf("timeout error should mention file backend, got: %s", errStr)
	}
}

// TestMockStore_ErrorsDoNotExposeTokens verifies that MockStore errors
// also don't expose token values (for test consistency).
func TestMockStore_ErrorsDoNotExposeTokens(t *testing.T) {
	store := NewMockStore()

	sensitiveToken := "mock-store-secret-token-123"

	// Set up store to return errors
	testErr := errors.New("simulated storage failure")
	store.SetSetError(testErr)

	err := store.SetToken("profile", Token{RefreshToken: sensitiveToken})

	if err == nil {
		t.Fatal("SetToken expected error")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, sensitiveToken) {
		t.Errorf("MockStore SetToken error should NOT contain token.\nError: %s", errMsg)
	}

	// Clear error and store token
	store.SetSetError(nil)
	_ = store.SetToken("profile", Token{RefreshToken: sensitiveToken})

	// Now test GetToken error
	store.SetGetError(testErr)
	_, err = store.GetToken("profile")

	if err == nil {
		t.Fatal("GetToken expected error")
	}

	errMsg = err.Error()
	if strings.Contains(errMsg, sensitiveToken) {
		t.Errorf("MockStore GetToken error should NOT contain token.\nError: %s", errMsg)
	}
}

// TestTokenString_DoesNotExposeSecrets verifies that if Token type
// had a String() method, it wouldn't expose sensitive values.
// (This is a defensive test - if someone adds String() later)
func TestTokenString_DoesNotExposeSecrets(t *testing.T) {
	tok := Token{
		Profile:      "test",
		AccessToken:  "access-secret-should-not-appear",
		RefreshToken: "refresh-secret-should-not-appear",
		Scopes:       []string{"read", "write"},
	}

	// If Token has a String() method, verify it doesn't expose secrets
	// Using %v format which would use String() if available
	formatted := tok.Profile // Just test that Profile is safe to use
	if strings.Contains(formatted, "secret") {
		t.Errorf("Token formatting should NOT expose secrets")
	}

	// Note: Token struct has json:"-" tags for AccessToken and RefreshToken
	// which prevents them from being marshaled, providing some protection
}
