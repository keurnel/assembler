package debugcontext

import "fmt"

// Severity constants for entry classification.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
	SeverityTrace   = "trace"
)

// Entry is a single diagnostic event recorded by the assembler pipeline.
// It captures what happened, where it happened, and how severe it is.
//
// Entries are append-only — once created, their core fields (severity, phase,
// message, location) are immutable. Only the optional fields (snippet, hint)
// can be set via the With* chaining methods before the entry is considered
// complete.
//
// See FR-6 in .requirements/assembler/debug-information.md.
type Entry struct {
	severity string   // "error" | "warning" | "info" | "trace"
	phase    string   // Pipeline phase at recording time.
	message  string   // Human-readable description.
	location Location // Source position the entry refers to.
	snippet  string   // Optional: source line text for inline display.
	hint     string   // Optional: fix suggestion for the user.
}

// --- Accessor methods ---

// Severity returns the entry's severity level.
func (e *Entry) Severity() string { return e.severity }

// Phase returns the pipeline phase that was active when the entry was recorded.
func (e *Entry) Phase() string { return e.phase }

// Message returns the human-readable description.
func (e *Entry) Message() string { return e.message }

// Location returns the source position the entry refers to.
func (e *Entry) Location() Location { return e.location }

// Snippet returns the optional source line text, or empty string.
func (e *Entry) Snippet() string { return e.snippet }

// Hint returns the optional fix suggestion, or empty string.
func (e *Entry) Hint() string { return e.hint }

// --- Chaining methods (FR-6.1, FR-6.2) ---

// WithSnippet sets the source line snippet and returns the same *Entry for chaining.
func (e *Entry) WithSnippet(text string) *Entry {
	e.snippet = text
	return e
}

// WithHint sets the fix suggestion and returns the same *Entry for chaining.
func (e *Entry) WithHint(text string) *Entry {
	e.hint = text
	return e
}

// String returns a single-line human-readable representation for quick debugging.
// Format: "severity [phase] location: message"
func (e *Entry) String() string {
	return fmt.Sprintf("%s [%s] %s: %s", e.severity, e.phase, e.location.String(), e.message)
}
