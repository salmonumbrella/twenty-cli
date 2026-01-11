package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

// SetupServer handles the browser-based credential setup flow
type SetupServer struct {
	listener  net.Listener
	server    *http.Server
	csrfToken string
	resultCh  chan SetupResult
	errCh     chan error
	timeout   time.Duration
	store     secrets.Store
}

// SetupResult contains the result of a successful setup
type SetupResult struct {
	Profile string
	BaseURL string
}

// SetupServerOption configures the setup server
type SetupServerOption func(*SetupServer)

// WithSetupTimeout sets the setup timeout
func WithSetupTimeout(d time.Duration) SetupServerOption {
	return func(s *SetupServer) {
		s.timeout = d
	}
}

// WithStore sets the secrets store
func WithStore(store secrets.Store) SetupServerOption {
	return func(s *SetupServer) {
		s.store = store
	}
}

// NewSetupServer creates a new setup server on a random available port
func NewSetupServer(opts ...SetupServerOption) (*SetupServer, error) {
	// Generate CSRF token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate csrf token: %w", err)
	}
	csrfToken := hex.EncodeToString(tokenBytes)

	// Listen on random available port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	s := &SetupServer{
		listener:  ln,
		csrfToken: csrfToken,
		resultCh:  make(chan SetupResult, 1),
		errCh:     make(chan error, 1),
		timeout:   DefaultTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Open default store if not provided
	if s.store == nil {
		store, err := secrets.OpenDefault()
		if err != nil {
			ln.Close()
			return nil, fmt.Errorf("open secrets store: %w", err)
		}
		s.store = store
	}

	return s, nil
}

// Port returns the port the server is listening on
func (s *SetupServer) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// URL returns the setup page URL
func (s *SetupServer) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/", s.Port())
}

// Start starts the HTTP server
func (s *SetupServer) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)

	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()

	go func() {
		if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case s.errCh <- err:
			default:
			}
		}
	}()
}

// WaitForSetup waits for the setup to complete
func (s *SetupServer) WaitForSetup(ctx context.Context) (SetupResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	select {
	case result := <-s.resultCh:
		return result, nil
	case err := <-s.errCh:
		return SetupResult{}, err
	case <-ctx.Done():
		return SetupResult{}, fmt.Errorf("setup timeout: %w", ctx.Err())
	}
}

// Close shuts down the server
func (s *SetupServer) Close() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// handleSetup serves the setup form
func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderSetupPage(w, s.csrfToken)
}

// validateRequest is the request body for /validate
type validateRequest struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

// validateResponse is the response for /validate
type validateResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// handleValidate tests the connection
func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate CSRF
	if r.Header.Get("X-CSRF-Token") != s.csrfToken {
		writeJSON(w, http.StatusForbidden, validateResponse{Error: "Invalid CSRF token"})
		return
	}

	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, validateResponse{Error: "Invalid request"})
		return
	}

	// Validate URL format
	if err := validateBaseURL(req.BaseURL); err != nil {
		writeJSON(w, http.StatusBadRequest, validateResponse{Error: err.Error()})
		return
	}

	// Test the connection by making an API call
	client := rest.NewClient(req.BaseURL, req.Token, false)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Try to get current user/workspace info
	var result interface{}
	err := client.Get(ctx, "/rest/api-keys", &result)
	if err != nil {
		writeJSON(w, http.StatusOK, validateResponse{
			Valid: false,
			Error: "Invalid token or unable to connect: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, validateResponse{Valid: true})
}

// submitRequest is the request body for /submit
type submitRequest struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
	Profile string `json:"profile"`
}

// submitResponse is the response for /submit
type submitResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// handleSubmit saves the credentials
func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate CSRF
	if r.Header.Get("X-CSRF-Token") != s.csrfToken {
		writeJSON(w, http.StatusForbidden, submitResponse{Error: "Invalid CSRF token"})
		return
	}

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, submitResponse{Error: "Invalid request"})
		return
	}

	profile := strings.TrimSpace(req.Profile)
	if profile == "" {
		profile = "default"
	}

	// Save token to store
	tok := secrets.Token{
		Profile:      profile,
		RefreshToken: req.Token, // API tokens are stored as "refresh tokens"
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.store.SetToken(profile, tok); err != nil {
		writeJSON(w, http.StatusInternalServerError, submitResponse{Error: "Failed to save token: " + err.Error()})
		return
	}

	// Check if this is the first profile - if so, set it as primary
	tokens, err := s.store.ListTokens()
	if err == nil && len(tokens) == 1 {
		_ = s.store.SetDefaultAccount(profile)
	}

	// Also set as primary if no primary is set
	if primary, err := s.store.GetDefaultAccount(); err == nil && primary == "" {
		_ = s.store.SetDefaultAccount(profile)
	}

	writeJSON(w, http.StatusOK, submitResponse{Success: true})

	// Signal success
	select {
	case s.resultCh <- SetupResult{Profile: profile, BaseURL: req.BaseURL}:
	default:
	}
}

// handleSuccess shows the success page
func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderSuccessPage(w)

	// Close server after a delay
	go func() {
		time.Sleep(2 * time.Second)
		_ = s.Close()
	}()
}

// validateBaseURL checks if the URL is valid
func validateBaseURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("base URL is required")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must use http or https")
	}

	if u.Host == "" {
		return errors.New("URL must have a host")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
