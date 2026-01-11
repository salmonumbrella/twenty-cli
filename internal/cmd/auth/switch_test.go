package auth

import (
	"bytes"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

func TestSwitchCmd_Success(t *testing.T) {
	mock := setupMockStore(t)

	// Add the profile
	_ = mock.SetToken("work", secrets.Token{RefreshToken: "work-token"})

	cmd := switchCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"work"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify default was set
	defaultProfile, err := mock.GetDefaultAccount()
	if err != nil {
		t.Fatalf("Failed to get default account: %v", err)
	}
	if defaultProfile != "work" {
		t.Errorf("Default profile = %q, want %q", defaultProfile, "work")
	}
}

func TestSwitchCmd_ProfileNotFound(t *testing.T) {
	mock := setupMockStore(t)

	// Add a different profile
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "default-token"})

	cmd := switchCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"nonexistent"})
	if err == nil {
		t.Error("Expected error for nonexistent profile")
	}

	expectedMsg := `profile "nonexistent" not found`
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want containing %q", err.Error(), expectedMsg)
	}
}

func TestSwitchCmd_MultipleProfiles(t *testing.T) {
	mock := setupMockStore(t)

	// Add multiple profiles
	_ = mock.SetToken("default", secrets.Token{RefreshToken: "default-token"})
	_ = mock.SetToken("work", secrets.Token{RefreshToken: "work-token"})
	_ = mock.SetToken("personal", secrets.Token{RefreshToken: "personal-token"})
	_ = mock.SetDefaultAccount("default")

	cmd := switchCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Switch to personal
	err := cmd.RunE(cmd, []string{"personal"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify default was changed
	defaultProfile, _ := mock.GetDefaultAccount()
	if defaultProfile != "personal" {
		t.Errorf("Default profile = %q, want %q", defaultProfile, "personal")
	}
}
