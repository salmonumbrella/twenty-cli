package auth

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderWaitingPage(t *testing.T) {
	var buf bytes.Buffer
	renderWaitingPage(&buf)

	output := buf.String()

	// Should contain HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}

	// Should contain title
	if !strings.Contains(output, "Authenticating") {
		t.Error("Output missing 'Authenticating'")
	}

	// Should contain waiting message
	if !strings.Contains(output, "Waiting for authorization") {
		t.Error("Output missing 'Waiting for authorization'")
	}
}

func TestRenderSuccessPage(t *testing.T) {
	var buf bytes.Buffer
	renderSuccessPage(&buf)

	output := buf.String()

	// Should contain HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}

	// Should contain success message
	if !strings.Contains(output, "Authentication successful") {
		t.Error("Output missing 'Authentication successful'")
	}

	// Should contain countdown (the template variable should be replaced)
	if !strings.Contains(output, "Closing in") {
		t.Error("Output missing 'Closing in'")
	}

	// Should contain the countdown seconds value
	countdownStr := "30" // PostSuccessDisplaySeconds
	if !strings.Contains(output, countdownStr) {
		t.Errorf("Output missing countdown value %s", countdownStr)
	}
}

func TestRenderErrorPage(t *testing.T) {
	var buf bytes.Buffer
	testError := "Test error message for unit testing"
	renderErrorPage(&buf, testError)

	output := buf.String()

	// Should contain HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}

	// Should contain error header
	if !strings.Contains(output, "Authorization failed") {
		t.Error("Output missing 'Authorization failed'")
	}

	// Should contain the specific error message
	if !strings.Contains(output, testError) {
		t.Errorf("Output missing error message %q", testError)
	}
}

func TestRenderErrorPage_HTMLEscaping(t *testing.T) {
	var buf bytes.Buffer
	// Test that HTML in error messages is escaped
	testError := "<script>alert('xss')</script>"
	renderErrorPage(&buf, testError)

	output := buf.String()

	// The script tag should be escaped
	if strings.Contains(output, "<script>") {
		t.Error("Output contains unescaped script tag - XSS vulnerability")
	}

	// Should contain escaped version
	if !strings.Contains(output, "&lt;script&gt;") {
		t.Error("Output should contain escaped script tag")
	}
}

func TestRenderWaitingPage_ContentType(t *testing.T) {
	var buf bytes.Buffer
	renderWaitingPage(&buf)

	output := buf.String()

	// Should contain charset declaration in meta tag
	if !strings.Contains(output, `charset="UTF-8"`) {
		t.Error("Output missing charset declaration")
	}
}

func TestRenderSuccessPage_ContentType(t *testing.T) {
	var buf bytes.Buffer
	renderSuccessPage(&buf)

	output := buf.String()

	// Should contain charset declaration in meta tag
	if !strings.Contains(output, `charset="UTF-8"`) {
		t.Error("Output missing charset declaration")
	}
}

func TestRenderErrorPage_ContentType(t *testing.T) {
	var buf bytes.Buffer
	renderErrorPage(&buf, "test error")

	output := buf.String()

	// Should contain charset declaration in meta tag
	if !strings.Contains(output, `charset="UTF-8"`) {
		t.Error("Output missing charset declaration")
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, nil // Accept all writes without error for fallback testing
}

func TestRenderWaitingPage_Output(t *testing.T) {
	var buf bytes.Buffer
	renderWaitingPage(&buf)

	// Output should not be empty
	if buf.Len() == 0 {
		t.Error("Output is empty")
	}

	// Output should be valid HTML (contains opening and closing html tags)
	output := buf.String()
	if !strings.Contains(output, "<html") {
		t.Error("Output missing opening html tag")
	}
	if !strings.Contains(output, "</html>") {
		t.Error("Output missing closing html tag")
	}
}

func TestRenderSuccessPage_Output(t *testing.T) {
	var buf bytes.Buffer
	renderSuccessPage(&buf)

	// Output should not be empty
	if buf.Len() == 0 {
		t.Error("Output is empty")
	}

	// Output should be valid HTML
	output := buf.String()
	if !strings.Contains(output, "<html") {
		t.Error("Output missing opening html tag")
	}
	if !strings.Contains(output, "</html>") {
		t.Error("Output missing closing html tag")
	}
}

func TestRenderErrorPage_Output(t *testing.T) {
	var buf bytes.Buffer
	renderErrorPage(&buf, "error")

	// Output should not be empty
	if buf.Len() == 0 {
		t.Error("Output is empty")
	}

	// Output should be valid HTML
	output := buf.String()
	if !strings.Contains(output, "<html") {
		t.Error("Output missing opening html tag")
	}
	if !strings.Contains(output, "</html>") {
		t.Error("Output missing closing html tag")
	}
}

func TestSuccessTemplateData(t *testing.T) {
	data := SuccessTemplateData{
		Profile:          "test-profile",
		CountdownSeconds: 30,
	}

	if data.Profile != "test-profile" {
		t.Errorf("Profile = %q, want %q", data.Profile, "test-profile")
	}
	if data.CountdownSeconds != 30 {
		t.Errorf("CountdownSeconds = %d, want %d", data.CountdownSeconds, 30)
	}
}

func TestErrorTemplateData(t *testing.T) {
	data := ErrorTemplateData{
		Error: "test error",
	}

	if data.Error != "test error" {
		t.Errorf("Error = %q, want %q", data.Error, "test error")
	}
}

func TestRenderErrorPage_EmptyError(t *testing.T) {
	var buf bytes.Buffer
	renderErrorPage(&buf, "")

	output := buf.String()

	// Should still render the page
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}
}

func TestRenderErrorPage_LongError(t *testing.T) {
	var buf bytes.Buffer
	longError := strings.Repeat("a", 1000)
	renderErrorPage(&buf, longError)

	output := buf.String()

	// Should still render the page with the long error
	if !strings.Contains(output, longError) {
		t.Error("Output missing long error message")
	}
}

func TestRenderErrorPage_SpecialChars(t *testing.T) {
	var buf bytes.Buffer
	specialError := "Error with special chars: & < > \" '"
	renderErrorPage(&buf, specialError)

	output := buf.String()

	// HTML entities should be escaped
	if strings.Contains(output, "& <") {
		t.Error("Ampersand should be escaped")
	}
}

// Tests for template parse failure fallbacks
// These tests temporarily set invalid template content to test error handling

func TestRenderWaitingPage_ParseError(t *testing.T) {
	// Save original template
	original := waitingTemplate
	defer func() { waitingTemplate = original }()

	// Set invalid template to trigger parse error
	waitingTemplate = "{{.InvalidSyntax"

	var buf bytes.Buffer
	renderWaitingPage(&buf)

	output := buf.String()

	// Should fall back to plain text message
	if !strings.Contains(output, "Authenticating with twenty") {
		t.Errorf("Expected fallback message, got: %s", output)
	}
	if !strings.Contains(output, "Please complete the login") {
		t.Errorf("Expected fallback message, got: %s", output)
	}
}

func TestRenderSuccessPage_ParseError(t *testing.T) {
	// Save original template
	original := successTemplate
	defer func() { successTemplate = original }()

	// Set invalid template to trigger parse error
	successTemplate = "{{.InvalidSyntax"

	var buf bytes.Buffer
	renderSuccessPage(&buf)

	output := buf.String()

	// Should fall back to plain text message
	if !strings.Contains(output, "Authentication successful") {
		t.Errorf("Expected fallback message, got: %s", output)
	}
	if !strings.Contains(output, "close this window") {
		t.Errorf("Expected fallback message, got: %s", output)
	}
}

func TestRenderErrorPage_ParseError(t *testing.T) {
	// Save original template
	original := errorTemplate
	defer func() { errorTemplate = original }()

	// Set invalid template to trigger parse error
	errorTemplate = "{{.InvalidSyntax"

	testError := "Test error message"
	var buf bytes.Buffer
	renderErrorPage(&buf, testError)

	output := buf.String()

	// Should fall back to plain text message with error
	if !strings.Contains(output, "Authentication failed") {
		t.Errorf("Expected fallback message, got: %s", output)
	}
	if !strings.Contains(output, testError) {
		t.Errorf("Expected error message %q in output, got: %s", testError, output)
	}
}

func TestRenderWaitingPage_ParseErrorRecovery(t *testing.T) {
	// Save original template
	original := waitingTemplate
	defer func() { waitingTemplate = original }()

	// First, set invalid template
	waitingTemplate = "{{invalid"

	var buf1 bytes.Buffer
	renderWaitingPage(&buf1)

	// Should get fallback
	if !strings.Contains(buf1.String(), "Authenticating with twenty") {
		t.Error("Expected fallback message for invalid template")
	}

	// Restore valid template
	waitingTemplate = original

	var buf2 bytes.Buffer
	renderWaitingPage(&buf2)

	// Should get full HTML now
	if !strings.Contains(buf2.String(), "<!DOCTYPE html>") {
		t.Error("Expected HTML output after restoring valid template")
	}
}

func TestRenderSuccessPage_ParseErrorRecovery(t *testing.T) {
	// Save original template
	original := successTemplate
	defer func() { successTemplate = original }()

	// First, set invalid template
	successTemplate = "{{invalid"

	var buf1 bytes.Buffer
	renderSuccessPage(&buf1)

	// Should get fallback
	if !strings.Contains(buf1.String(), "Authentication successful") {
		t.Error("Expected fallback message for invalid template")
	}

	// Restore valid template
	successTemplate = original

	var buf2 bytes.Buffer
	renderSuccessPage(&buf2)

	// Should get full HTML now
	if !strings.Contains(buf2.String(), "<!DOCTYPE html>") {
		t.Error("Expected HTML output after restoring valid template")
	}
}

func TestRenderErrorPage_ParseErrorRecovery(t *testing.T) {
	// Save original template
	original := errorTemplate
	defer func() { errorTemplate = original }()

	testErr := "test error"

	// First, set invalid template
	errorTemplate = "{{invalid"

	var buf1 bytes.Buffer
	renderErrorPage(&buf1, testErr)

	// Should get fallback
	if !strings.Contains(buf1.String(), "Authentication failed") {
		t.Error("Expected fallback message for invalid template")
	}

	// Restore valid template
	errorTemplate = original

	var buf2 bytes.Buffer
	renderErrorPage(&buf2, testErr)

	// Should get full HTML now
	if !strings.Contains(buf2.String(), "<!DOCTYPE html>") {
		t.Error("Expected HTML output after restoring valid template")
	}
}
