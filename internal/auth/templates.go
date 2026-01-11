package auth

import (
	_ "embed"
	"html/template"
	"io"
)

//go:embed templates/waiting.html
var waitingTemplateContent string

//go:embed templates/success.html
var successTemplateContent string

//go:embed templates/error.html
var errorTemplateContent string

//go:embed templates/setup.html
var setupTemplateContent string //nolint:unused // used by setupTemplate below

// Template variables that can be overridden in tests to simulate parse failures
var (
	waitingTemplate = waitingTemplateContent
	successTemplate = successTemplateContent
	errorTemplate   = errorTemplateContent
	setupTemplate   = setupTemplateContent //nolint:unused // used by renderSetupPage
)

// SuccessTemplateData holds data for the success page
type SuccessTemplateData struct {
	Profile          string
	CountdownSeconds int
}

// ErrorTemplateData holds data for the error page
type ErrorTemplateData struct {
	Error string
}

// SetupTemplateData holds data for the setup page
type SetupTemplateData struct {
	CSRFToken string
}

// renderWaitingPage renders the waiting/authenticating page
func renderWaitingPage(w io.Writer) {
	tmpl, err := template.New("waiting").Parse(waitingTemplate)
	if err != nil {
		_, _ = w.Write([]byte("Authenticating with twenty... Please complete the login in your browser."))
		return
	}
	_ = tmpl.Execute(w, nil)
}

// renderSuccessPage renders the authentication success page
func renderSuccessPage(w io.Writer) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		_, _ = w.Write([]byte("Authentication successful! You can close this window and return to the terminal."))
		return
	}
	data := SuccessTemplateData{
		CountdownSeconds: PostSuccessDisplaySeconds,
	}
	_ = tmpl.Execute(w, data)
}

// renderErrorPage renders the authentication error page
func renderErrorPage(w io.Writer, errorMsg string) {
	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		_, _ = w.Write([]byte("Authentication failed: " + errorMsg))
		return
	}
	_ = tmpl.Execute(w, ErrorTemplateData{Error: errorMsg})
}

// renderSetupPage renders the credential setup form
//
//nolint:unused // will be used by setup server in upcoming task
func renderSetupPage(w io.Writer, csrfToken string) {
	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		_, _ = w.Write([]byte("Error loading setup page. Please use: twenty auth login --no-browser"))
		return
	}
	_ = tmpl.Execute(w, SetupTemplateData{CSRFToken: csrfToken})
}
