package debugcontext

import "testing"

func TestEntry_WithSnippet(t *testing.T) {
	// ==============================================================
	// FR-6.1: WithSnippet sets the snippet and returns *Entry.
	// ==============================================================
	entry := &Entry{severity: SeverityError, message: "test"}

	returned := entry.WithSnippet("  mov rax, 1")

	if returned != entry {
		t.Fatal("WithSnippet must return the same *Entry for chaining")
	}
	if entry.Snippet() != "  mov rax, 1" {
		t.Errorf("Expected snippet '  mov rax, 1', got '%s'", entry.Snippet())
	}
}

func TestEntry_WithHint(t *testing.T) {
	// ==============================================================
	// FR-6.2: WithHint sets the hint and returns *Entry.
	// ==============================================================
	entry := &Entry{severity: SeverityWarning, message: "test"}

	returned := entry.WithHint("did you mean 'mov'?")

	if returned != entry {
		t.Fatal("WithHint must return the same *Entry for chaining")
	}
	if entry.Hint() != "did you mean 'mov'?" {
		t.Errorf("Expected hint \"did you mean 'mov'?\", got '%s'", entry.Hint())
	}
}

func TestEntry_Chaining(t *testing.T) {
	// ==============================================================
	// FR-3.3.5: Chaining WithSnippet and WithHint.
	// ==============================================================
	entry := &Entry{severity: SeverityError, message: "unknown instruction"}

	entry.WithSnippet("  mvo rax, 1").WithHint("did you mean 'mov'?")

	if entry.Snippet() != "  mvo rax, 1" {
		t.Errorf("Expected snippet '  mvo rax, 1', got '%s'", entry.Snippet())
	}
	if entry.Hint() != "did you mean 'mov'?" {
		t.Errorf("Expected hint, got '%s'", entry.Hint())
	}
}

func TestEntry_String(t *testing.T) {
	// ==============================================================
	// FR-6.3: String() returns a single-line representation.
	// ==============================================================
	entry := &Entry{
		severity: SeverityError,
		phase:    "pre-processing/includes",
		message:  "unknown file 'missing.kasm'",
		location: Loc("main.kasm", 12, 0),
	}

	expected := "error [pre-processing/includes] main.kasm:12: unknown file 'missing.kasm'"
	if entry.String() != expected {
		t.Errorf("Expected %q, got %q", expected, entry.String())
	}
}

func TestEntry_Accessors(t *testing.T) {
	loc := Loc("test.kasm", 5, 3)
	entry := &Entry{
		severity: SeverityWarning,
		phase:    "lexing",
		message:  "test message",
		location: loc,
		snippet:  "some code",
		hint:     "fix it",
	}

	if entry.Severity() != SeverityWarning {
		t.Errorf("Expected severity '%s', got '%s'", SeverityWarning, entry.Severity())
	}
	if entry.Phase() != "lexing" {
		t.Errorf("Expected phase 'lexing', got '%s'", entry.Phase())
	}
	if entry.Message() != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message())
	}
	if entry.Location() != loc {
		t.Errorf("Expected location %v, got %v", loc, entry.Location())
	}
	if entry.Snippet() != "some code" {
		t.Errorf("Expected snippet 'some code', got '%s'", entry.Snippet())
	}
	if entry.Hint() != "fix it" {
		t.Errorf("Expected hint 'fix it', got '%s'", entry.Hint())
	}
}
