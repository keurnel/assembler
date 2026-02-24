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

func TestTracker_SnapshotWithInclusions(t *testing.T) {
	// ==============================================================
	//
	// FR-11.2.3: SnapshotWithInclusions annotates expanding lines
	// with the source file they were included from.
	//
	// ==============================================================
	t.Run("annotates expanding lines with source file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "main.kasm")
		// Original source has a single %include directive placeholder.
		if err := os.WriteFile(path, []byte("%include \"header.kasm\""), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		// Simulate what PreProcessingHandleIncludes produces:
		// the %include is replaced with ; FILE: / content / ; END FILE: markers.
		expandedSource := "; FILE: header.kasm\nmov rax, 1\nxor rbx, rbx\n; END FILE: header.kasm"

		tracker.SnapshotWithInclusions(expandedSource, []Inclusion{
			{FilePath: "header.kasm", LineNumber: 1},
		})

		// Verify the history of an included line (index 1: "mov rax, 1").
		history := tracker.History(1)
		if len(history) != 1 {
			t.Fatalf("Expected 1 history entry, got %d", len(history))
		}

		entry := history[0]
		if entry.Type() != "expanding" {
			t.Errorf("Expected type 'expanding', got '%s'", entry.Type())
		}
		if entry.SourceFile() != "header.kasm" {
			t.Errorf("Expected sourceFile 'header.kasm', got '%s'", entry.SourceFile())
		}
		if entry.Content() != "mov rax, 1" {
			t.Errorf("Expected content 'mov rax, 1', got '%s'", entry.Content())
		}
	})

	t.Run("does not annotate lines outside file markers", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "main.kasm")
		// Original: two lines, first is an include placeholder, second is real code.
		if err := os.WriteFile(path, []byte("%include \"h.kasm\"\nmov rcx, 0"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		// After include expansion, "mov rcx, 0" stays as-is at the end.
		expandedSource := "; FILE: h.kasm\npush rbp\n; END FILE: h.kasm\nmov rcx, 0"

		tracker.SnapshotWithInclusions(expandedSource, []Inclusion{
			{FilePath: "h.kasm", LineNumber: 1},
		})

		// "mov rcx, 0" at index 3 is unchanged (it was in the original).
		history := tracker.History(3)
		if len(history) != 1 {
			t.Fatalf("Expected 1 history entry, got %d", len(history))
		}

		entry := history[0]
		// Unchanged line should have empty sourceFile.
		if entry.SourceFile() != "" {
			t.Errorf("Expected empty sourceFile for unchanged line, got '%s'", entry.SourceFile())
		}
	})

	t.Run("FILE marker lines are annotated with the included file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "main.kasm")
		if err := os.WriteFile(path, []byte("original_line"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		expandedSource := "; FILE: lib.kasm\nadd rax, rbx\n; END FILE: lib.kasm"

		tracker.SnapshotWithInclusions(expandedSource, []Inclusion{
			{FilePath: "lib.kasm", LineNumber: 1},
		})

		// Index 0: "; FILE: lib.kasm" — expanding, annotated.
		snapshot := tracker.instance.LatestSnapshot()
		change0 := (*snapshot.changes)[0]
		if change0.SourceFile() != "lib.kasm" {
			t.Errorf("Expected '; FILE:' marker annotated with 'lib.kasm', got '%s'", change0.SourceFile())
		}

		// Index 2: "; END FILE: lib.kasm" — expanding, annotated.
		change2 := (*snapshot.changes)[2]
		if change2.SourceFile() != "lib.kasm" {
			t.Errorf("Expected '; END FILE:' marker annotated with 'lib.kasm', got '%s'", change2.SourceFile())
		}
	})

	t.Run("multiple included files are annotated correctly", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "main.kasm")
		if err := os.WriteFile(path, []byte("%include \"a.kasm\"\n%include \"b.kasm\""), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		tracker, _ := Track(path)

		expandedSource := "; FILE: a.kasm\nline_a\n; END FILE: a.kasm\n; FILE: b.kasm\nline_b\n; END FILE: b.kasm"

		tracker.SnapshotWithInclusions(expandedSource, []Inclusion{
			{FilePath: "a.kasm", LineNumber: 1},
			{FilePath: "b.kasm", LineNumber: 2},
		})

		snapshot := tracker.instance.LatestSnapshot()

		// line_a at index 1 should be from a.kasm.
		changeA := (*snapshot.changes)[1]
		if changeA.SourceFile() != "a.kasm" {
			t.Errorf("Expected sourceFile 'a.kasm', got '%s'", changeA.SourceFile())
		}

		// line_b at index 4 should be from b.kasm.
		changeB := (*snapshot.changes)[4]
		if changeB.SourceFile() != "b.kasm" {
			t.Errorf("Expected sourceFile 'b.kasm', got '%s'", changeB.SourceFile())
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
