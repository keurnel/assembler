package x86_64

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/internal/lineMap"
)

// ---------------------------------------------------------------------------
// FR-1.6: Circular Include Detection
// ---------------------------------------------------------------------------

// TestPreProcessIncludes_CircularTwoFiles verifies the canonical circular
// inclusion scenario (FR-1.6): file1.kasm includes file2.kasm, and
// file2.kasm includes file1.kasm.
func TestPreProcessIncludes_CircularTwoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.kasm")
	file2 := filepath.Join(tmpDir, "file2.kasm")

	// file1 includes file2, file2 includes file1 — a cycle.
	os.WriteFile(file1, []byte(`%include "`+file2+`"`), 0644)
	os.WriteFile(file2, []byte(`%include "`+file1+`"`), 0644)

	source, _ := os.ReadFile(file1)
	debugCtx := debugcontext.NewDebugContext(file1)
	tracker, err := lineMap.Track(file1)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	_ = preProcessIncludes(string(source), file1, tracker, debugCtx)

	if !debugCtx.HasErrors() {
		t.Fatal("expected circular inclusion error, got none")
	}

	errors := debugCtx.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	msg := errors[0].String()
	if !strings.Contains(msg, "circular inclusion") {
		t.Errorf("expected error message to contain 'circular inclusion', got: %s", msg)
	}
	if !strings.Contains(msg, file1) {
		t.Errorf("expected error message to contain the offending file path '%s', got: %s", file1, msg)
	}
}

// TestPreProcessIncludes_CircularThreeFiles verifies circular inclusion
// detection across a three-file chain: a → b → c → a.
func TestPreProcessIncludes_CircularThreeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	fileA := filepath.Join(tmpDir, "a.kasm")
	fileB := filepath.Join(tmpDir, "b.kasm")
	fileC := filepath.Join(tmpDir, "c.kasm")

	os.WriteFile(fileA, []byte(`%include "`+fileB+`"`), 0644)
	os.WriteFile(fileB, []byte(`%include "`+fileC+`"`), 0644)
	os.WriteFile(fileC, []byte(`%include "`+fileA+`"`), 0644)

	source, _ := os.ReadFile(fileA)
	debugCtx := debugcontext.NewDebugContext(fileA)
	tracker, err := lineMap.Track(fileA)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	_ = preProcessIncludes(string(source), fileA, tracker, debugCtx)

	if !debugCtx.HasErrors() {
		t.Fatal("expected circular inclusion error, got none")
	}

	msg := debugCtx.Errors()[0].String()
	if !strings.Contains(msg, "circular inclusion") {
		t.Errorf("expected error message to contain 'circular inclusion', got: %s", msg)
	}
}

// TestPreProcessIncludes_SelfInclude verifies that a file including itself
// is caught. The pre-processor function itself detects this within a single
// invocation (FR-1.6.1 / FR-1.2.2), but the orchestrator's root-seeding
// (FR-1.6.6) provides an additional layer.
func TestPreProcessIncludes_SelfInclude(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "self.kasm")
	os.WriteFile(file, []byte(`%include "`+file+`"`), 0644)

	source, _ := os.ReadFile(file)
	debugCtx := debugcontext.NewDebugContext(file)
	tracker, err := lineMap.Track(file)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	// The pre-processor's own include handling will read the file and inline
	// it, then the orchestrator loop will detect that the file path is already
	// in the seen set (seeded by FR-1.6.6).
	_ = preProcessIncludes(string(source), file, tracker, debugCtx)

	if !debugCtx.HasErrors() {
		t.Fatal("expected circular inclusion error for self-include, got none")
	}

	msg := debugCtx.Errors()[0].String()
	if !strings.Contains(msg, "circular inclusion") {
		t.Errorf("expected 'circular inclusion' in error, got: %s", msg)
	}
}

// TestPreProcessIncludes_NoCircular verifies that legitimate includes (no
// cycle) do not trigger a circular inclusion error.
func TestPreProcessIncludes_NoCircular(t *testing.T) {
	tmpDir := t.TempDir()

	helper := filepath.Join(tmpDir, "helper.kasm")
	os.WriteFile(helper, []byte("mov rax, 1"), 0644)

	root := filepath.Join(tmpDir, "main.kasm")
	os.WriteFile(root, []byte(`%include "`+helper+`"`+"\nmov rbx, 2"), 0644)

	source, _ := os.ReadFile(root)
	debugCtx := debugcontext.NewDebugContext(root)
	tracker, err := lineMap.Track(root)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	result := preProcessIncludes(string(source), root, tracker, debugCtx)

	if debugCtx.HasErrors() {
		t.Fatalf("expected no errors, got: %v", debugCtx.Errors())
	}

	if !strings.Contains(result, "mov rax, 1") {
		t.Error("expected included content in result")
	}
	if !strings.Contains(result, "mov rbx, 2") {
		t.Error("expected original content in result")
	}
}

// TestPreProcessIncludes_RecursiveNonCircular verifies that multi-level
// includes (a includes b, b includes c) work without triggering circular
// inclusion when there is no cycle.
func TestPreProcessIncludes_RecursiveNonCircular(t *testing.T) {
	tmpDir := t.TempDir()

	fileC := filepath.Join(tmpDir, "c.kasm")
	os.WriteFile(fileC, []byte("nop"), 0644)

	fileB := filepath.Join(tmpDir, "b.kasm")
	os.WriteFile(fileB, []byte(`%include "`+fileC+`"`), 0644)

	fileA := filepath.Join(tmpDir, "a.kasm")
	os.WriteFile(fileA, []byte(`%include "`+fileB+`"`), 0644)

	source, _ := os.ReadFile(fileA)
	debugCtx := debugcontext.NewDebugContext(fileA)
	tracker, err := lineMap.Track(fileA)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	result := preProcessIncludes(string(source), fileA, tracker, debugCtx)

	if debugCtx.HasErrors() {
		t.Fatalf("expected no errors, got: %v", debugCtx.Errors())
	}

	if !strings.Contains(result, "nop") {
		t.Error("expected deeply included content 'nop' in result")
	}
}

// TestPreProcessIncludes_RootFileSeeded verifies FR-1.6.6: the root file
// path is in the seen set so that an included file that re-includes the
// root is caught.
func TestPreProcessIncludes_RootFileSeeded(t *testing.T) {
	tmpDir := t.TempDir()

	root := filepath.Join(tmpDir, "root.kasm")
	child := filepath.Join(tmpDir, "child.kasm")

	// child includes the root — cycle through the root file.
	os.WriteFile(root, []byte(`%include "`+child+`"`), 0644)
	os.WriteFile(child, []byte(`%include "`+root+`"`), 0644)

	source, _ := os.ReadFile(root)
	debugCtx := debugcontext.NewDebugContext(root)
	tracker, err := lineMap.Track(root)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	_ = preProcessIncludes(string(source), root, tracker, debugCtx)

	if !debugCtx.HasErrors() {
		t.Fatal("expected circular inclusion error when child re-includes root")
	}

	msg := debugCtx.Errors()[0].String()
	if !strings.Contains(msg, "circular inclusion") {
		t.Errorf("expected 'circular inclusion' in error, got: %s", msg)
	}
	// The offending path should be the root file (re-included by child).
	if !strings.Contains(msg, root) {
		t.Errorf("expected root path in error message, got: %s", msg)
	}
}

// TestPreProcessIncludes_ErrorMessageFormat verifies FR-1.6.7: the error
// message must use the phrase "circular inclusion" and include the file path.
func TestPreProcessIncludes_ErrorMessageFormat(t *testing.T) {
	tmpDir := t.TempDir()

	fileX := filepath.Join(tmpDir, "x.kasm")
	fileY := filepath.Join(tmpDir, "y.kasm")

	os.WriteFile(fileX, []byte(`%include "`+fileY+`"`), 0644)
	os.WriteFile(fileY, []byte(`%include "`+fileX+`"`), 0644)

	source, _ := os.ReadFile(fileX)
	debugCtx := debugcontext.NewDebugContext(fileX)
	tracker, err := lineMap.Track(fileX)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	_ = preProcessIncludes(string(source), fileX, tracker, debugCtx)

	errors := debugCtx.Errors()
	if len(errors) == 0 {
		t.Fatal("expected at least one error")
	}

	msg := errors[0].String()

	// FR-1.6.7: must contain "circular inclusion".
	if !strings.Contains(msg, "circular inclusion") {
		t.Errorf("error message must contain 'circular inclusion', got: %s", msg)
	}

	// FR-1.6.7: must contain the file path for grep-based log analysis.
	if !strings.Contains(msg, fileX) {
		t.Errorf("error message must contain the file path, got: %s", msg)
	}
}
