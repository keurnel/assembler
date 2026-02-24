package debugcontext

import (
	"sync"
	"testing"
)

func TestNewDebugContext(t *testing.T) {
	// ==============================================================
	// FR-1.1: Sole constructor returns a fully initialised context.
	// FR-1.2: Primary file path is stored.
	// ==============================================================
	t.Run("creates context with file path and empty state", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")

		if ctx == nil {
			t.Fatal("Expected non-nil DebugContext")
		}
		if ctx.FilePath() != "main.kasm" {
			t.Errorf("Expected file path 'main.kasm', got '%s'", ctx.FilePath())
		}
		if ctx.Phase() != "" {
			t.Errorf("Expected empty phase, got '%s'", ctx.Phase())
		}
		if ctx.Count() != 0 {
			t.Errorf("Expected 0 entries, got %d", ctx.Count())
		}
	})
}

func TestDebugContext_Phases(t *testing.T) {
	// ==============================================================
	// FR-2.1: SetPhase sets the current phase.
	// FR-2.3: Phase() returns the current phase name.
	// ==============================================================
	t.Run("SetPhase and Phase", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")

		ctx.SetPhase("pre-processing/includes")
		if ctx.Phase() != "pre-processing/includes" {
			t.Errorf("Expected phase 'pre-processing/includes', got '%s'", ctx.Phase())
		}

		ctx.SetPhase("lexing")
		if ctx.Phase() != "lexing" {
			t.Errorf("Expected phase 'lexing', got '%s'", ctx.Phase())
		}
	})

	// ==============================================================
	// FR-2.2 / FR-3.2.2: Phase is attached to entries automatically.
	// ==============================================================
	t.Run("entries inherit the current phase", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")

		ctx.SetPhase("pre-processing/macros")
		ctx.Error(ctx.Loc(1, 0), "macro error")

		ctx.SetPhase("lexing")
		ctx.Warning(ctx.Loc(5, 3), "lex warning")

		entries := ctx.Entries()
		if entries[0].Phase() != "pre-processing/macros" {
			t.Errorf("Expected first entry phase 'pre-processing/macros', got '%s'", entries[0].Phase())
		}
		if entries[1].Phase() != "lexing" {
			t.Errorf("Expected second entry phase 'lexing', got '%s'", entries[1].Phase())
		}
	})
}

func TestDebugContext_Location(t *testing.T) {
	// ==============================================================
	// FR-4.1: Loc uses the primary file path.
	// ==============================================================
	t.Run("Loc uses primary file path", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		loc := ctx.Loc(10, 5)

		if loc.FilePath() != "main.kasm" {
			t.Errorf("Expected file path 'main.kasm', got '%s'", loc.FilePath())
		}
		if loc.Line() != 10 {
			t.Errorf("Expected line 10, got %d", loc.Line())
		}
		if loc.Column() != 5 {
			t.Errorf("Expected column 5, got %d", loc.Column())
		}
	})

	// ==============================================================
	// FR-4.2: LocIn uses an explicit file path.
	// ==============================================================
	t.Run("LocIn uses explicit file path", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		loc := ctx.LocIn("header.kasm", 3, 0)

		if loc.FilePath() != "header.kasm" {
			t.Errorf("Expected file path 'header.kasm', got '%s'", loc.FilePath())
		}
		if loc.Line() != 3 {
			t.Errorf("Expected line 3, got %d", loc.Line())
		}
	})
}

func TestDebugContext_Recording(t *testing.T) {
	// ==============================================================
	// FR-3.3.1 through FR-3.3.4: Recording methods for each severity.
	// ==============================================================
	t.Run("Error records entry with severity error", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		ctx.SetPhase("parsing")

		entry := ctx.Error(ctx.Loc(10, 0), "unknown instruction")

		if entry.Severity() != SeverityError {
			t.Errorf("Expected severity '%s', got '%s'", SeverityError, entry.Severity())
		}
		if entry.Message() != "unknown instruction" {
			t.Errorf("Expected message 'unknown instruction', got '%s'", entry.Message())
		}
		if ctx.Count() != 1 {
			t.Errorf("Expected 1 entry, got %d", ctx.Count())
		}
	})

	t.Run("Warning records entry with severity warning", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		entry := ctx.Warning(ctx.Loc(5, 0), "unused label")

		if entry.Severity() != SeverityWarning {
			t.Errorf("Expected severity '%s', got '%s'", SeverityWarning, entry.Severity())
		}
	})

	t.Run("Info records entry with severity info", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		entry := ctx.Info(ctx.Loc(1, 0), "macro expanded")

		if entry.Severity() != SeverityInfo {
			t.Errorf("Expected severity '%s', got '%s'", SeverityInfo, entry.Severity())
		}
	})

	t.Run("Trace records entry with severity trace", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		entry := ctx.Trace(ctx.Loc(1, 0), "internal debug info")

		if entry.Severity() != SeverityTrace {
			t.Errorf("Expected severity '%s', got '%s'", SeverityTrace, entry.Severity())
		}
	})

	// ==============================================================
	// FR-3.3.5: Recording methods return *Entry for chaining.
	// ==============================================================
	t.Run("chaining WithSnippet and WithHint from recording method", func(t *testing.T) {
		ctx := NewDebugContext("main.kasm")
		ctx.SetPhase("parsing")

		ctx.Error(ctx.Loc(10, 3), "unknown instruction").
			WithSnippet("  mvo rax, 1").
			WithHint("did you mean 'mov'?")

		entries := ctx.Entries()
		if len(entries) != 1 {
			t.Fatalf("Expected 1 entry, got %d", len(entries))
		}

		e := entries[0]
		if e.Snippet() != "  mvo rax, 1" {
			t.Errorf("Expected snippet '  mvo rax, 1', got '%s'", e.Snippet())
		}
		if e.Hint() != "did you mean 'mov'?" {
			t.Errorf("Expected hint, got '%s'", e.Hint())
		}
	})
}

func TestDebugContext_Querying(t *testing.T) {
	ctx := NewDebugContext("main.kasm")

	ctx.Error(ctx.Loc(1, 0), "error 1")
	ctx.Warning(ctx.Loc(2, 0), "warning 1")
	ctx.Error(ctx.Loc(3, 0), "error 2")
	ctx.Info(ctx.Loc(4, 0), "info 1")
	ctx.Trace(ctx.Loc(5, 0), "trace 1")

	// ==============================================================
	// FR-5.1: Entries() returns all in insertion order.
	// ==============================================================
	t.Run("Entries returns all in order", func(t *testing.T) {
		entries := ctx.Entries()
		if len(entries) != 5 {
			t.Fatalf("Expected 5 entries, got %d", len(entries))
		}
		if entries[0].Message() != "error 1" {
			t.Errorf("Expected first entry 'error 1', got '%s'", entries[0].Message())
		}
		if entries[4].Message() != "trace 1" {
			t.Errorf("Expected last entry 'trace 1', got '%s'", entries[4].Message())
		}
	})

	// ==============================================================
	// FR-5.2: Errors() returns only error entries.
	// ==============================================================
	t.Run("Errors returns only errors", func(t *testing.T) {
		errors := ctx.Errors()
		if len(errors) != 2 {
			t.Fatalf("Expected 2 errors, got %d", len(errors))
		}
		if errors[0].Message() != "error 1" || errors[1].Message() != "error 2" {
			t.Error("Errors returned wrong entries")
		}
	})

	// ==============================================================
	// FR-5.3: Warnings() returns only warning entries.
	// ==============================================================
	t.Run("Warnings returns only warnings", func(t *testing.T) {
		warnings := ctx.Warnings()
		if len(warnings) != 1 {
			t.Fatalf("Expected 1 warning, got %d", len(warnings))
		}
		if warnings[0].Message() != "warning 1" {
			t.Errorf("Expected 'warning 1', got '%s'", warnings[0].Message())
		}
	})

	// ==============================================================
	// FR-5.4: HasErrors() returns true when errors exist.
	// ==============================================================
	t.Run("HasErrors returns true when errors exist", func(t *testing.T) {
		if !ctx.HasErrors() {
			t.Error("Expected HasErrors() to return true")
		}
	})

	t.Run("HasErrors returns false when no errors", func(t *testing.T) {
		clean := NewDebugContext("clean.kasm")
		clean.Warning(clean.Loc(1, 0), "just a warning")

		if clean.HasErrors() {
			t.Error("Expected HasErrors() to return false")
		}
	})

	// ==============================================================
	// FR-5.5: Count() returns total number of entries.
	// ==============================================================
	t.Run("Count returns total entries", func(t *testing.T) {
		if ctx.Count() != 5 {
			t.Errorf("Expected 5, got %d", ctx.Count())
		}
	})
}

func TestDebugContext_Entries_ReturnsCopy(t *testing.T) {
	// ==============================================================
	// FR-8.2: Entries are append-only; returned slice is a copy.
	// ==============================================================
	ctx := NewDebugContext("main.kasm")
	ctx.Error(ctx.Loc(1, 0), "original")

	entries := ctx.Entries()
	entries[0] = nil // Mutate the returned slice.

	// The context's internal entries must be unaffected.
	if ctx.Entries()[0] == nil {
		t.Error("Entries() must return a copy, not a reference to the internal slice")
	}
}

func TestDebugContext_ThreadSafety(t *testing.T) {
	// ==============================================================
	// FR-3.3.6 / FR-8.3: Thread-safe concurrent writes.
	// ==============================================================
	ctx := NewDebugContext("main.kasm")

	var wg sync.WaitGroup
	const goroutines = 100

	wg.Add(goroutines)
	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			ctx.Error(ctx.Loc(n, 0), "concurrent error")
		}(i)
	}
	wg.Wait()

	if ctx.Count() != goroutines {
		t.Errorf("Expected %d entries from concurrent writes, got %d", goroutines, ctx.Count())
	}
}

func TestDebugContext_InsertionOrder(t *testing.T) {
	// ==============================================================
	// FR-5.1 / FR-8.2: Entries preserve insertion order.
	// ==============================================================
	ctx := NewDebugContext("main.kasm")

	ctx.SetPhase("pre-processing/includes")
	ctx.Error(ctx.Loc(1, 0), "first")

	ctx.SetPhase("lexing")
	ctx.Warning(ctx.Loc(2, 0), "second")

	ctx.SetPhase("parsing")
	ctx.Info(ctx.Loc(3, 0), "third")

	entries := ctx.Entries()
	expected := []string{"first", "second", "third"}
	for i, msg := range expected {
		if entries[i].Message() != msg {
			t.Errorf("Entry %d: expected message '%s', got '%s'", i, msg, entries[i].Message())
		}
	}
}

func TestDebugContext_IncludedFileLocation(t *testing.T) {
	// ==============================================================
	// FR-4.2 / FR-9.2: Entries from included files use LocIn.
	// ==============================================================
	ctx := NewDebugContext("main.kasm")
	ctx.SetPhase("pre-processing/includes")

	loc := ctx.LocIn("header.kasm", 5, 0)
	ctx.Error(loc, "syntax error in included file")

	entry := ctx.Entries()[0]
	if entry.Location().FilePath() != "header.kasm" {
		t.Errorf("Expected file path 'header.kasm', got '%s'", entry.Location().FilePath())
	}
	if entry.String() != "error [pre-processing/includes] header.kasm:5: syntax error in included file" {
		t.Errorf("Unexpected String(): %s", entry.String())
	}
}
