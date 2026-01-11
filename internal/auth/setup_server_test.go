package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestNewSetupServer(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	if server.Port() == 0 {
		t.Error("Expected non-zero port")
	}

	if server.csrfToken == "" {
		t.Error("Expected non-empty CSRF token")
	}

	// Verify CSRF token is hex-encoded 32 bytes (64 hex chars)
	if len(server.csrfToken) != 64 {
		t.Errorf("Expected CSRF token of 64 hex chars, got %d", len(server.csrfToken))
	}
}

func TestSetupServer_HandleSetup(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("Connect to Twenty")) {
		t.Error("Expected setup page content 'Connect to Twenty'")
	}

	// Verify CSRF token is embedded in the page
	if !bytes.Contains(w.Body.Bytes(), []byte(server.csrfToken)) {
		t.Error("Expected CSRF token to be embedded in page")
	}
}

func TestSetupServer_HandleSetup_NotFound(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestSetupServer_HandleValidate_CSRFRequired(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	body := `{"base_url":"https://test.com","token":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// No CSRF token
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var resp validateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != "Invalid CSRF token" {
		t.Errorf("Expected 'Invalid CSRF token' error, got '%s'", resp.Error)
	}
}

func TestSetupServer_HandleValidate_WrongCSRF(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	body := `{"base_url":"https://test.com","token":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "wrong-token")
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestSetupServer_HandleValidate_MethodNotAllowed(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSetupServer_HandleValidate_InvalidJSON(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var resp validateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != "Invalid request" {
		t.Errorf("Expected 'Invalid request' error, got '%s'", resp.Error)
	}
}

func TestSetupServer_HandleValidate_InvalidBaseURL(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	body := `{"base_url":"not-a-url","token":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var resp validateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error message for invalid URL")
	}
}

func TestSetupServer_HandleSubmit_SavesToken(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	body := `{"base_url":"https://test.com","token":"test-token","profile":"myprofile"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp submitResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success, got error: %s", resp.Error)
	}

	// Verify token was saved
	tok, err := mockStore.GetToken("myprofile")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if tok.RefreshToken != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", tok.RefreshToken)
	}
	if tok.Profile != "myprofile" {
		t.Errorf("Expected profile 'myprofile', got '%s'", tok.Profile)
	}
}

func TestSetupServer_HandleSubmit_DefaultProfile(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	// Submit without profile name
	body := `{"base_url":"https://test.com","token":"test-token","profile":""}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify token was saved with "default" profile
	tok, err := mockStore.GetToken("default")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if tok.Profile != "default" {
		t.Errorf("Expected profile 'default', got '%s'", tok.Profile)
	}
}

func TestSetupServer_HandleSubmit_CSRFRequired(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	body := `{"base_url":"https://test.com","token":"test-token","profile":"myprofile"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// No CSRF token
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	// Verify token was NOT saved
	_, err = mockStore.GetToken("myprofile")
	if err == nil {
		t.Error("Expected token not to be saved without CSRF")
	}
}

func TestSetupServer_HandleSubmit_MethodNotAllowed(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/submit", nil)
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSetupServer_FirstProfileSetAsPrimary(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	body := `{"base_url":"https://test.com","token":"test-token","profile":"first"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify first profile was set as primary
	primary, err := mockStore.GetDefaultAccount()
	if err != nil {
		t.Fatalf("Failed to get default account: %v", err)
	}
	if primary != "first" {
		t.Errorf("Expected primary 'first', got '%s'", primary)
	}
}

func TestSetupServer_SecondProfileDoesNotChangePrimary(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	// Add first profile
	body1 := `{"base_url":"https://test.com","token":"token1","profile":"first"}`
	req1 := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-CSRF-Token", server.csrfToken)
	w1 := httptest.NewRecorder()
	server.handleSubmit(w1, req1)

	// Add second profile
	body2 := `{"base_url":"https://test.com","token":"token2","profile":"second"}`
	req2 := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-CSRF-Token", server.csrfToken)
	w2 := httptest.NewRecorder()
	server.handleSubmit(w2, req2)

	// Verify primary is still "first"
	primary, err := mockStore.GetDefaultAccount()
	if err != nil {
		t.Fatalf("Failed to get default account: %v", err)
	}
	if primary != "first" {
		t.Errorf("Expected primary 'first', got '%s'", primary)
	}
}

func TestSetupServer_NoPrimarySetsPrimary(t *testing.T) {
	mockStore := secrets.NewMockStore()

	// Pre-populate with a profile but no primary set
	err := mockStore.SetToken("existing", secrets.Token{RefreshToken: "existing-token"})
	if err != nil {
		t.Fatalf("Failed to set existing token: %v", err)
	}

	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	// Add another profile
	body := `{"base_url":"https://test.com","token":"new-token","profile":"newprofile"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	// Verify primary was set to the new profile (since no primary existed)
	primary, err := mockStore.GetDefaultAccount()
	if err != nil {
		t.Fatalf("Failed to get default account: %v", err)
	}
	if primary != "newprofile" {
		t.Errorf("Expected primary 'newprofile', got '%s'", primary)
	}
}

func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid https URL",
			url:     "https://app.twenty.com",
			wantErr: false,
		},
		{
			name:    "valid http localhost",
			url:     "http://localhost:3000",
			wantErr: false,
		},
		{
			name:    "valid https with path",
			url:     "https://example.com/api",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name:    "URL without scheme",
			url:     "app.twenty.com",
			wantErr: true,
			errMsg:  "URL must use http or https",
		},
		{
			name:    "ftp scheme",
			url:     "ftp://invalid.com",
			wantErr: true,
			errMsg:  "URL must use http or https",
		},
		{
			name:    "file scheme",
			url:     "file:///etc/passwd",
			wantErr: true,
			errMsg:  "URL must use http or https",
		},
		{
			name:    "URL without host",
			url:     "https://",
			wantErr: true,
			errMsg:  "URL must have a host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBaseURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBaseURL(%q) error = %v, wantErr = %v", tt.url, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("validateBaseURL(%q) error = %q, want %q", tt.url, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestSetupServer_URL(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	url := server.URL()
	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// URL should be localhost with port
	expectedPrefix := "http://127.0.0.1:"
	if len(url) <= len(expectedPrefix) {
		t.Errorf("URL too short: %s", url)
	}
	if url[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected URL to start with %s, got %s", expectedPrefix, url)
	}
}

func TestSetupServer_HandleSuccess(t *testing.T) {
	mockStore := secrets.NewMockStore()
	server, err := NewSetupServer(WithStore(mockStore))
	if err != nil {
		t.Fatalf("NewSetupServer failed: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/success", nil)
	w := httptest.NewRecorder()

	server.handleSuccess(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected key=value, got key=%s", result["key"])
	}
}
