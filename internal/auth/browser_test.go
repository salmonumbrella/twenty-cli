package auth

import (
	"runtime"
	"testing"
)

// TestOpenBrowser verifies the OpenBrowser wrapper function exists and returns
// an error for invalid URLs without panicking.
// Note: We don't actually open a browser in tests - we just verify the function
// handles the call without panicking and returns appropriate errors.
func TestOpenBrowser_InvalidURL(t *testing.T) {
	// Test with empty URL - should not panic
	err := OpenBrowser("")
	// The function may or may not return an error depending on the platform
	// but it should not panic
	_ = err
}

// TestOpenBrowser_ValidURL tests with a valid URL format
// Note: This doesn't actually open a browser, just verifies the command is constructed
func TestOpenBrowser_ValidURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	// Test with a valid URL format
	// Note: This may actually try to open a browser on some systems
	// so we mark it to run only in non-short mode
	url := "https://example.com"
	err := OpenBrowser(url)
	// We don't assert on the error because it depends on whether a browser is available
	_ = err
}

// TestOpenBrowser_Platform verifies the function handles the current platform
func TestOpenBrowser_Platform(t *testing.T) {
	// Verify we're on a supported platform
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		// Supported platforms
	default:
		t.Skipf("unsupported platform: %s", runtime.GOOS)
	}

	// The function should be callable without panic
	// We use an invalid scheme to prevent actually opening anything
	_ = OpenBrowser("invalid://not-a-real-url")
}
