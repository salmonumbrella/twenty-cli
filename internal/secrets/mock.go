package secrets

import (
	"time"

	"github.com/99designs/keyring"
)

// MockStore implements Store interface for testing
type MockStore struct {
	tokens         map[string]Token
	defaultAccount string
	getErr         error // Simulate errors on GetToken
	setErr         error // Simulate errors on SetToken
}

// NewMockStore creates a new MockStore for testing
func NewMockStore() *MockStore {
	return &MockStore{
		tokens: make(map[string]Token),
	}
}

// Keys returns all token profile keys
func (m *MockStore) Keys() ([]string, error) {
	keys := make([]string, 0, len(m.tokens))
	for k := range m.tokens {
		keys = append(keys, refreshTokenKey(k))
	}
	return keys, nil
}

// SetToken stores a token for the given profile
func (m *MockStore) SetToken(profile string, tok Token) error {
	if m.setErr != nil {
		return m.setErr
	}

	profile = normalize(profile)
	if profile == "" {
		return errMissingProfile
	}

	if tok.RefreshToken == "" {
		return errMissingRefreshToken
	}

	tok.Profile = profile
	if tok.CreatedAt.IsZero() {
		tok.CreatedAt = time.Now().UTC()
	}

	m.tokens[profile] = tok
	return nil
}

// GetToken retrieves a token for the given profile
func (m *MockStore) GetToken(profile string) (Token, error) {
	if m.getErr != nil {
		return Token{}, m.getErr
	}

	profile = normalize(profile)
	if profile == "" {
		return Token{}, errMissingProfile
	}

	tok, ok := m.tokens[profile]
	if !ok {
		return Token{}, keyring.ErrKeyNotFound
	}

	return tok, nil
}

// DeleteToken removes a token for the given profile
func (m *MockStore) DeleteToken(profile string) error {
	profile = normalize(profile)
	if profile == "" {
		return errMissingProfile
	}

	if _, ok := m.tokens[profile]; !ok {
		return keyring.ErrKeyNotFound
	}

	delete(m.tokens, profile)
	return nil
}

// ListTokens returns all stored tokens
func (m *MockStore) ListTokens() ([]Token, error) {
	tokens := make([]Token, 0, len(m.tokens))
	for _, tok := range m.tokens {
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

// GetDefaultAccount returns the default account profile
func (m *MockStore) GetDefaultAccount() (string, error) {
	return m.defaultAccount, nil
}

// SetDefaultAccount sets the default account profile
func (m *MockStore) SetDefaultAccount(profile string) error {
	profile = normalize(profile)
	if profile == "" {
		return errMissingProfile
	}

	m.defaultAccount = profile
	return nil
}

// SetGetError sets an error to be returned by GetToken
func (m *MockStore) SetGetError(err error) {
	m.getErr = err
}

// SetSetError sets an error to be returned by SetToken
func (m *MockStore) SetSetError(err error) {
	m.setErr = err
}

// Reset clears all tokens and errors
func (m *MockStore) Reset() {
	m.tokens = make(map[string]Token)
	m.defaultAccount = ""
	m.getErr = nil
	m.setErr = nil
}

// Verify MockStore implements Store interface
var _ Store = (*MockStore)(nil)
