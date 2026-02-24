package lineMap

import (
	"testing"
)

// newTestInstance is a helper that creates an Instance via New() for testing.
// It uses a zero-value Source since tests don't need file I/O.
func newTestInstance(t *testing.T, value string) *Instance {
	t.Helper()
	return New(value, Source{path: "test.kasm"})
}

func TestNew(t *testing.T) {
	// ==============================================================
	//
	// Creates a fully initialised Instance with an initial snapshot.
	// No separate InitialIndex() call needed.
	//
	// ==============================================================
	t.Run("creates Instance with initial snapshot", func(t *testing.T) {
		source := Source{path: "fakePath.kasm", content: "fake file content"}

		instance := New("line1\nline2\nline3", source)

		if instance == nil {
			t.Fatal("Expected a new instance of `Instance`, got nil")
		}

		if instance.value != "line1\nline2\nline3" {
			t.Errorf("Expected instance value to be 'line1\\nline2\\nline3', got '%s'", instance.value)
		}

		if instance.source.Path() != "fakePath.kasm" {
			t.Errorf("Expected source path 'fakePath.kasm', got '%s'", instance.source.Path())
		}

		if instance.source.Content() != "fake file content" {
			t.Errorf("Expected source content 'fake file content', got '%s'", instance.source.Content())
		}
	})

	// ==============================================================
	//
	// The initial snapshot is created as part of New() — history has exactly 1 item.
	//
	// ==============================================================
	t.Run("initial snapshot exists in history", func(t *testing.T) {
		instance := New("line1\nline2\nline3", Source{path: "test.kasm"})

		if len(instance.history.items) != 1 {
			t.Fatalf("Expected history to have 1 item after New(), got %d", len(instance.history.items))
		}

		snapshot := instance.history.items[0]
		if snapshot._type != LineSnapshotTypeInitial {
			t.Errorf("Expected snapshot type '%s', got '%s'", LineSnapshotTypeInitial, snapshot._type)
		}

		if snapshot.source != "line1\nline2\nline3" {
			t.Errorf("Expected snapshot source 'line1\\nline2\\nline3', got '%s'", snapshot.source)
		}

		if len(snapshot.lines) != 3 {
			t.Errorf("Expected 3 lines in snapshot, got %d", len(snapshot.lines))
		}

		expectedLines := []string{"line1", "line2", "line3"}
		for i, line := range snapshot.lines {
			if line != expectedLines[i] {
				t.Errorf("Expected line %d to be '%s', got '%s'", i, expectedLines[i], line)
			}
		}
	})

	// ==============================================================
	//
	// Update() can be called immediately after New() without any setup.
	//
	// ==============================================================
	t.Run("Update works immediately after New", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2")

		err := instance.Update("line1\nline2\nline3")
		if err != nil {
			t.Fatalf("Expected Update to succeed immediately after New(), got '%s'", err.Error())
		}

		if len(instance.history.items) != 2 {
			t.Errorf("Expected 2 history items, got %d", len(instance.history.items))
		}
	})
}

func TestInstance_Update(t *testing.T) {
	// ==============================================================
	//
	// Expanding a line
	//
	// ==============================================================
	t.Run("Expanding a line", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2\nline3")

		err := instance.Update("line1\nline2\nline3\nline4")
		if err != nil {
			t.Errorf("Expected Update to succeed when updating with a new value, got '%s'", err.Error())
		}

		if len(instance.history.items) != 2 {
			t.Errorf("Expected Instance.history to have 2 items after Update, got %d", len(instance.history.items))
		}

		snapshot := instance.history.items[1]

		if snapshot._type != LineSnapshotTypeChange {
			t.Errorf("Expected snapshot type to be '%s', got '%s'", LineSnapshotTypeChange, snapshot._type)
		}

		if snapshot.source != "line1\nline2\nline3\nline4" {
			t.Errorf("Expected snapshot source to be 'line1\\nline2\\nline3\\nline4', got '%s'", snapshot.source)
		}

		if len(snapshot.lines) != 4 {
			t.Errorf("Expected snapshot lines to have 4 lines, got %d", len(snapshot.lines))
		}
	})
	// ==============================================================
	//
	// Expanding a line stores changes in snapshot
	//
	// ==============================================================
	t.Run("Expanding a line stores changes in snapshot", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2\nline3")

		err := instance.Update("line1\nline2\nline3\nline4")
		if err != nil {
			t.Fatalf("Expected Update to succeed, got '%s'", err.Error())
		}

		snapshot := instance.history.items[1]
		if snapshot.changes == nil {
			t.Fatal("Expected changes to be non-nil on a change snapshot")
		}

		// line4 at index 3 should be an expanding change
		change, exists := (*snapshot.changes)[3]
		if !exists {
			t.Fatal("Expected an expanding change at new line index 3")
		}
		if change._type != "expanding" {
			t.Errorf("Expected change type 'expanding', got '%s'", change._type)
		}
	})
	// ==============================================================
	//
	// Contracting lines (removing lines)
	//
	// ==============================================================
	t.Run("Contracting lines", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2\nline3\nline4")

		err := instance.Update("line1\nline4")
		if err != nil {
			t.Fatalf("Expected Update to succeed, got '%s'", err.Error())
		}

		if len(instance.history.items) != 2 {
			t.Fatalf("Expected 2 history items, got %d", len(instance.history.items))
		}

		snapshot := instance.history.items[1]
		if snapshot.changes == nil {
			t.Fatal("Expected changes to be non-nil")
		}

		if len(snapshot.lines) != 2 {
			t.Errorf("Expected 2 lines after contraction, got %d", len(snapshot.lines))
		}
	})
	// ==============================================================
	//
	// No change update creates NoChange snapshot
	//
	// ==============================================================
	t.Run("No change update creates NoChange snapshot", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2")

		err := instance.Update("line1\nline2")
		if err != nil {
			t.Fatalf("Expected Update to succeed, got '%s'", err.Error())
		}

		if len(instance.history.items) != 2 {
			t.Fatalf("Expected 2 history items, got %d", len(instance.history.items))
		}

		snapshot := instance.history.items[1]
		if snapshot._type != LineSnapshotTypeNoChange {
			t.Errorf("Expected snapshot type '%s', got '%s'", LineSnapshotTypeNoChange, snapshot._type)
		}

		if snapshot.changes != nil {
			t.Error("Expected changes to be nil on no-change snapshot")
		}
	})
	// ==============================================================
	//
	// Multi-step preprocessing: simulating include expansion then macro expansion
	//
	// ==============================================================
	t.Run("Multi-step preprocessing history", func(t *testing.T) {
		instance := newTestInstance(t, "include_marker\nmov rax, 1")

		// Step 1: Include expands "include_marker" into 3 lines
		err := instance.Update("; FILE: header.kasm\nmov rbx, 0\nxor rcx, rcx\n; END FILE: header.kasm\nmov rax, 1")
		if err != nil {
			t.Fatalf("Step 1 failed: %s", err.Error())
		}

		// Step 2: Macro expansion replaces "mov rax, 1" with two lines
		err = instance.Update("; FILE: header.kasm\nmov rbx, 0\nxor rcx, rcx\n; END FILE: header.kasm\npush 1\npop rax")
		if err != nil {
			t.Fatalf("Step 2 failed: %s", err.Error())
		}

		if instance.SnapshotCount() != 3 {
			t.Errorf("Expected 3 snapshots (initial + 2 updates), got %d", instance.SnapshotCount())
		}

		// The final value should have 6 lines
		lines := instance.Lines()
		if len(lines) != 6 {
			t.Errorf("Expected 6 lines in final state, got %d", len(lines))
		}
	})
	// ==============================================================
	//
	// LineOrigin traces line back through history
	//
	// ==============================================================
	t.Run("LineOrigin traces unchanged line", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2\nline3")

		// Add a line at the beginning, shifting everything down
		err := instance.Update("new_line\nline1\nline2\nline3")
		if err != nil {
			t.Fatalf("Update failed: %s", err.Error())
		}

		// "line1" was at index 0, now at index 1
		origin := instance.LineOrigin(1)
		if origin != 0 {
			t.Errorf("Expected line 1 to originate from line 0, got %d", origin)
		}

		// "line3" was at index 2, now at index 3
		origin = instance.LineOrigin(3)
		if origin != 2 {
			t.Errorf("Expected line 3 to originate from line 2, got %d", origin)
		}
	})
	// ==============================================================
	//
	// LineOrigin returns -1 for inserted lines
	//
	// ==============================================================
	t.Run("LineOrigin returns -1 for inserted lines", func(t *testing.T) {
		instance := newTestInstance(t, "line1\nline2")

		// Insert a new line in the middle
		err := instance.Update("line1\nnew_inserted_line\nline2")
		if err != nil {
			t.Fatalf("Update failed: %s", err.Error())
		}

		// The inserted line should return -1 (no origin in the initial source)
		origin := instance.LineOrigin(1)
		if origin != -1 {
			t.Errorf("Expected inserted line to have origin -1, got %d", origin)
		}
	})
}
