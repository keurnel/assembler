package lineMap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrack(t *testing.T) {
	// ==============================================================
	//
	// FR-11.1.1: Track() loads the file and returns a ready-to-use Tracker.
	//
	// ==============================================================
	t.Run("creates Tracker from valid .kasm file", func(t *testing.T) {
		// Create a temporary .kasm file.
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		content := "line1\nline2\nline3"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, err := Track(path)
		if err != nil {
			t.Fatalf("Expected Track to succeed, got error: %v", err)
		}
		if tracker == nil {
			t.Fatal("Expected non-nil Tracker")
		}
	})

	// ==============================================================
	//
	// FR-11.1.1: Track() returns an error if LoadSource fails.
	//
	// ==============================================================
	t.Run("returns error for non-.kasm file", func(t *testing.T) {
		_, err := Track("/tmp/test.txt")
		if err == nil {
			t.Fatal("Expected error for non-.kasm file, got nil")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := Track("/tmp/nonexistent_file.kasm")
		if err == nil {
			t.Fatal("Expected error for non-existent file, got nil")
		}
	})

	// ==============================================================
	//
	// FR-11.1.3: The file content is used for the initial snapshot.
	//
	// ==============================================================
	t.Run("initial source matches file content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		content := "mov rax, 1\nxor rbx, rbx"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, err := Track(path)
		if err != nil {
			t.Fatalf("Track failed: %v", err)
		}

		if tracker.Source() != content {
			t.Errorf("Expected source %q, got %q", content, tracker.Source())
		}
	})
}

func TestTracker_Snapshot(t *testing.T) {
	// ==============================================================
	//
	// FR-11.2.1: Snapshot records a new version of the source.
	// FR-11.2.2: Snapshot is infallible.
	//
	// ==============================================================
	t.Run("records pre-processing transformation", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		if err := os.WriteFile(path, []byte("line1\nline2"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		tracker.Snapshot("line1\nline2\nline3")

		if tracker.Source() != "line1\nline2\nline3" {
			t.Errorf("Expected source after snapshot to be 'line1\\nline2\\nline3', got %q", tracker.Source())
		}

		lines := tracker.Lines()
		if len(lines) != 3 {
			t.Errorf("Expected 3 lines, got %d", len(lines))
		}
	})
}

func TestTracker_Origin(t *testing.T) {
	// ==============================================================
	//
	// FR-11.3.1: Origin traces a line back to its original position.
	//
	// ==============================================================
	t.Run("traces unchanged line to original position", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		if err := os.WriteFile(path, []byte("line1\nline2\nline3"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		// Insert a line at the beginning — "line1" shifts from 0 to 1.
		tracker.Snapshot("new_line\nline1\nline2\nline3")

		origin := tracker.Origin(1)
		if origin != 0 {
			t.Errorf("Expected origin 0 for shifted line1, got %d", origin)
		}
	})

	t.Run("returns -1 for inserted line", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		if err := os.WriteFile(path, []byte("line1\nline2"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		tracker.Snapshot("line1\ninserted\nline2")

		origin := tracker.Origin(1)
		if origin != -1 {
			t.Errorf("Expected -1 for inserted line, got %d", origin)
		}
	})
}

func TestTracker_History(t *testing.T) {
	// ==============================================================
	//
	// FR-11.3.2: History returns chronological evolution.
	//
	// ==============================================================
	t.Run("returns chronological line history", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		if err := os.WriteFile(path, []byte("line1\nline2"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		// Step 1: Insert at beginning.
		tracker.Snapshot("header\nline1\nline2")

		// Step 2: Append at end.
		tracker.Snapshot("header\nline1\nline2\nfooter")

		// Trace "line2" which is now at index 2.
		history := tracker.History(2)
		if len(history) != 2 {
			t.Fatalf("Expected 2 history entries, got %d", len(history))
		}

		// Oldest first.
		if history[0].Type() != "unchanged" {
			t.Errorf("Expected oldest entry type 'unchanged', got '%s'", history[0].Type())
		}
	})
}

func TestTracker_ReadAccess(t *testing.T) {
	// ==============================================================
	//
	// FR-11.4: Read access methods.
	//
	// ==============================================================
	t.Run("Source returns current processed source", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		content := "mov rax, 1"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		if tracker.Source() != content {
			t.Errorf("Expected %q, got %q", content, tracker.Source())
		}
	})

	t.Run("Lines returns lines of current source", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		if err := os.WriteFile(path, []byte("a\nb\nc"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		lines := tracker.Lines()
		if len(lines) != 3 {
			t.Fatalf("Expected 3 lines, got %d", len(lines))
		}
		if lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
			t.Errorf("Expected [a, b, c], got %v", lines)
		}
	})

	t.Run("FilePath returns original path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.kasm")
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		if tracker.FilePath() != path {
			t.Errorf("Expected %q, got %q", path, tracker.FilePath())
		}
	})
}
