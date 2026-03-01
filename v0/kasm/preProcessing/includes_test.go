package preProcessing_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keurnel/assembler/v0/kasm/preProcessing"
)

// --- PreProcessingHandleIncludes ---

func TestPreProcessingHandleIncludes_SingleInclude(t *testing.T) {
	// Create a temporary .kasm file to include
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "hulp.kasm")
	os.WriteFile(includePath, []byte("mov rax, 1\nmov rdi, 0"), 0644)

	source := `%include "` + includePath + `"`
	result, inclusions := preProcessing.HandleIncludes(source, nil)

	if len(inclusions) != 1 {
		t.Fatalf("expected 1 inclusion, got %d", len(inclusions))
	}

	if inclusions[0].IncludedFilePath != includePath {
		t.Errorf("expected path '%s', got '%s'", includePath, inclusions[0].IncludedFilePath)
	}

	if inclusions[0].LineNumber != 1 {
		t.Errorf("expected line number 1, got %d", inclusions[0].LineNumber)
	}

	if !containsSubstring(result, "; FILE: "+includePath) {
		t.Error("expected ; FILE: comment in result")
	}

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected included content in result")
	}

	if !containsSubstring(result, "; END FILE: "+includePath) {
		t.Error("expected ; END FILE: comment in result")
	}
}

func TestPreProcessingHandleIncludes_MultipleIncludes(t *testing.T) {
	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "a.kasm")
	path2 := filepath.Join(tmpDir, "b.kasm")
	os.WriteFile(path1, []byte("mov rax, 1"), 0644)
	os.WriteFile(path2, []byte("mov rbx, 2"), 0644)

	source := `%include "` + path1 + `"
%include "` + path2 + `"`
	result, inclusions := preProcessing.HandleIncludes(source, nil)

	if len(inclusions) != 2 {
		t.Fatalf("expected 2 inclusions, got %d", len(inclusions))
	}

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected content from first include")
	}
	if !containsSubstring(result, "mov rbx, 2") {
		t.Error("expected content from second include")
	}
}

func TestPreProcessingHandleIncludes_NoIncludes(t *testing.T) {
	source := `mov rax, 1`
	result, inclusions := preProcessing.HandleIncludes(source, nil)

	if len(inclusions) != 0 {
		t.Fatalf("expected 0 inclusions, got %d", len(inclusions))
	}

	if result != source {
		t.Errorf("expected source unchanged, got '%s'", result)
	}
}

func TestPreProcessingHandleIncludes_NonKasmExtension_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-.kasm file")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "must have a .kasm extension") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `%include "module.asm"`
	preProcessing.HandleIncludes(source, nil)
}

// TestPreProcessingHandleIncludes_DuplicateInclude_Deduplicated verifies
// FR-1.7: duplicate %include directives within a single invocation are
// silently deduplicated — the file is inlined once, the duplicate is stripped.
func TestPreProcessingHandleIncludes_DuplicateInclude_Deduplicated(t *testing.T) {
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "hulp.kasm")
	os.WriteFile(includePath, []byte("nop"), 0644)

	source := `%include "` + includePath + `"
%include "` + includePath + `"`
	result, inclusions := preProcessing.HandleIncludes(source, nil)

	// Only one inclusion should be recorded (the first occurrence).
	if len(inclusions) != 1 {
		t.Fatalf("expected 1 inclusion, got %d", len(inclusions))
	}

	// Content should appear exactly once.
	count := strings.Count(result, "nop")
	if count != 1 {
		t.Errorf("expected 'nop' to appear once, got %d times:\n%s", count, result)
	}

	// No remaining %include directives.
	if strings.Contains(result, "%include") {
		t.Errorf("expected no remaining %%include directives, got:\n%s", result)
	}
}

func TestPreProcessingHandleIncludes_FileNotFound_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing file")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "Failed to read included file") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `%include "nonexistent.kasm"`
	preProcessing.HandleIncludes(source, nil)
}

func TestPreProcessingHandleIncludes_LineNumber(t *testing.T) {
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "hulp.kasm")
	os.WriteFile(includePath, []byte("nop"), 0644)

	source := `; line 1
; line 2
%include "` + includePath + `"`
	_, inclusions := preProcessing.HandleIncludes(source, nil)

	if len(inclusions) != 1 {
		t.Fatalf("expected 1 inclusion, got %d", len(inclusions))
	}
	if inclusions[0].LineNumber != 3 {
		t.Errorf("expected line number 3, got %d", inclusions[0].LineNumber)
	}
}

func TestPreProcessingHandleIncludes_TrimWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "hulp.kasm")
	os.WriteFile(includePath, []byte("\n  mov rax, 1\n\n"), 0644)

	source := `%include "` + includePath + `"`
	result, _ := preProcessing.HandleIncludes(source, nil)

	// Should not start with newline inside FILE block (trimmed)
	if !containsSubstring(result, "; FILE: "+includePath+"\nmov rax, 1\n; END FILE:") {
		t.Errorf("expected trimmed content in result, got:\n%s", result)
	}
}

// ---------------------------------------------------------------------------
// FR-1.7: Shared Dependency Deduplication
// ---------------------------------------------------------------------------

// TestPreProcessingHandleIncludes_SharedDependency_Skipped verifies FR-1.7.1:
// when a %include path is in the alreadyIncluded set, the directive is silently
// removed and the file is not inlined again.
func TestPreProcessingHandleIncludes_SharedDependency_Skipped(t *testing.T) {
	tmpDir := t.TempDir()
	sharedPath := filepath.Join(tmpDir, "shared.kasm")
	os.WriteFile(sharedPath, []byte("mov rax, 42"), 0644)

	source := `%include "` + sharedPath + `"` + "\nmov rbx, 1"
	alreadyIncluded := map[string]bool{sharedPath: true}

	result, inclusions := preProcessing.HandleIncludes(source, alreadyIncluded)

	// FR-1.7.1: The directive should be stripped.
	if strings.Contains(result, "%include") {
		t.Errorf("expected %%include directive to be stripped, got:\n%s", result)
	}

	// FR-1.7.3: The file content should NOT be inlined a second time.
	if strings.Contains(result, "mov rax, 42") {
		t.Errorf("expected shared dependency content to NOT be inlined, got:\n%s", result)
	}

	// FR-1.7.4: Should not be an error — no panic, no inclusion entry.
	if len(inclusions) != 0 {
		t.Errorf("expected 0 inclusions for shared dependency, got %d", len(inclusions))
	}

	// Original content should be preserved.
	if !strings.Contains(result, "mov rbx, 1") {
		t.Error("expected original content to be preserved")
	}
}

// TestPreProcessingHandleIncludes_SharedDependency_MixedWithNew verifies
// FR-1.7.2: new includes are inlined normally while shared dependencies are
// silently skipped.
func TestPreProcessingHandleIncludes_SharedDependency_MixedWithNew(t *testing.T) {
	tmpDir := t.TempDir()
	sharedPath := filepath.Join(tmpDir, "shared.kasm")
	newPath := filepath.Join(tmpDir, "new.kasm")
	os.WriteFile(sharedPath, []byte("mov rax, 42"), 0644)
	os.WriteFile(newPath, []byte("mov rcx, 99"), 0644)

	source := `%include "` + sharedPath + `"` + "\n" + `%include "` + newPath + `"`
	alreadyIncluded := map[string]bool{sharedPath: true}

	result, inclusions := preProcessing.HandleIncludes(source, alreadyIncluded)

	// FR-1.7.2: The new file should be inlined normally.
	if len(inclusions) != 1 {
		t.Fatalf("expected 1 inclusion (new file only), got %d", len(inclusions))
	}
	if inclusions[0].IncludedFilePath != newPath {
		t.Errorf("expected inclusion path '%s', got '%s'", newPath, inclusions[0].IncludedFilePath)
	}

	// New file content should be inlined.
	if !strings.Contains(result, "mov rcx, 99") {
		t.Error("expected new file content in result")
	}
	if !strings.Contains(result, "; FILE: "+newPath) {
		t.Error("expected ; FILE: comment for new include")
	}

	// Shared file content should NOT be inlined.
	if strings.Contains(result, "mov rax, 42") {
		t.Error("shared dependency content should not be inlined")
	}
}

// TestPreProcessingHandleIncludes_SharedDependency_NilSet verifies that
// passing nil for alreadyIncluded works identically to the previous behaviour
// (no deduplication).
func TestPreProcessingHandleIncludes_SharedDependency_NilSet(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "helper.kasm")
	os.WriteFile(path, []byte("nop"), 0644)

	source := `%include "` + path + `"`
	result, inclusions := preProcessing.HandleIncludes(source, nil)

	if len(inclusions) != 1 {
		t.Fatalf("expected 1 inclusion, got %d", len(inclusions))
	}
	if !strings.Contains(result, "nop") {
		t.Error("expected included content in result")
	}
}

func BenchmarkPreProcessingHandleIncludes_NoIncludes(b *testing.B) {
	source := "mov rax, 1\nmov rdi, 0\nsyscall\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HandleIncludes(source, nil)
	}
}

func BenchmarkPreProcessingHandleIncludes_SingleInclude(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "hulp.kasm")
	os.WriteFile(path, []byte("mov rax, 1\nmov rdi, 0\nsyscall"), 0644)

	source := `%include "` + path + `"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HandleIncludes(source, nil)
	}
}

func BenchmarkPreProcessingHandleIncludes_MultipleIncludes(b *testing.B) {
	tmpDir := b.TempDir()
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("module_%d.kasm", i))
		os.WriteFile(path, []byte(fmt.Sprintf("mov rax, %d\nmov rdi, 0\nsyscall", i)), 0644)
		sb.WriteString(fmt.Sprintf("%%include \"%s\"\n", path))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HandleIncludes(source, nil)
	}
}

func BenchmarkPreProcessingHandleIncludes_LargeIncludedFile(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "large.kasm")

	var content strings.Builder
	for i := 0; i < 500; i++ {
		content.WriteString(fmt.Sprintf("mov r%d, %d\n", i%16, i))
	}
	os.WriteFile(path, []byte(content.String()), 0644)

	source := `%include "` + path + `"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HandleIncludes(source, nil)
	}
}

func BenchmarkPreProcessingHandleIncludes_ManyIncludes(b *testing.B) {
	tmpDir := b.TempDir()
	var sb strings.Builder
	for i := 0; i < 20; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("mod_%d.kasm", i))
		os.WriteFile(path, []byte(fmt.Sprintf("mov rax, %d", i)), 0644)
		sb.WriteString(fmt.Sprintf("%%include \"%s\"\n", path))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HandleIncludes(source, nil)
	}
}

func BenchmarkPreProcessingHandleIncludes_IncludeDeepInSource(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "tail.kasm")
	os.WriteFile(path, []byte("mov rax, 1"), 0644)

	var sb strings.Builder
	for i := 0; i < 200; i++ {
		sb.WriteString(fmt.Sprintf("; line %d\n", i))
	}
	sb.WriteString(`%include "` + path + `"`)
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HandleIncludes(source, nil)
	}
}
