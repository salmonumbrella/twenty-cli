package auth

import (
	"strings"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

var store secrets.Store

// storeOpener is the function used to open a new store.
// It can be overridden in tests to simulate errors.
var storeOpener = secrets.OpenDefault

// SetStore overrides the store (used for testing).
func SetStore(s secrets.Store) {
	store = s
}

// SetStoreOpener overrides the store opener function (used for testing).
func SetStoreOpener(opener func() (secrets.Store, error)) {
	storeOpener = opener
}

func getStore() (secrets.Store, error) {
	if store != nil {
		return store, nil
	}
	var err error
	store, err = storeOpener()
	return store, err
}

// SaveToken stores an API token for a profile
func SaveToken(profile, token string) error {
	s, err := getStore()
	if err != nil {
		return err
	}
	// Adapt to existing store interface - use token as "RefreshToken"
	// since that's what the existing store expects for the primary credential
	return s.SetToken(profile, secrets.Token{
		Profile:      profile,
		RefreshToken: token, // Store API token as refresh token
	})
}

// GetToken retrieves the API token for a profile
func GetToken(profile string) (string, error) {
	s, err := getStore()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(profile) == "" {
		profile, err = ResolveProfile("")
		if err != nil {
			return "", err
		}
	}
	tok, err := s.GetToken(profile)
	if err != nil {
		return "", err
	}
	return tok.RefreshToken, nil // API token stored as refresh token
}

// DeleteToken removes the token for a profile
func DeleteToken(profile string) error {
	s, err := getStore()
	if err != nil {
		return err
	}
	return s.DeleteToken(profile)
}

// ListProfiles returns all stored profile names
func ListProfiles() ([]string, error) {
	s, err := getStore()
	if err != nil {
		return nil, err
	}
	tokens, err := s.ListTokens()
	if err != nil {
		return nil, err
	}
	profiles := make([]string, len(tokens))
	for i, t := range tokens {
		profiles[i] = t.Profile
	}
	return profiles, nil
}
