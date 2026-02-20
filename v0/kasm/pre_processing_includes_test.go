package kasm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keurnel/assembler/v0/kasm"
)

// --- PreProcessingHandleIncludes ---

func TestPreProcessingHandleIncludes_SingleInclude(t *testing.T) {
	// Create a temporary .kasm file to include
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "module.kasm")
	os.WriteFile(includePath, []byte("mov rax, 1\nmov rdi, 0"), 0644)

	source := `%include "` + includePath + `"`
	result, inclusions := kasm.PreProcessingHandleIncludes(source)

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
	result, inclusions := kasm.PreProcessingHandleIncludes(source)

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
	result, inclusions := kasm.PreProcessingHandleIncludes(source)

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
	kasm.PreProcessingHandleIncludes(source)
}

func TestPreProcessingHandleIncludes_DuplicateInclude_Panics(t *testing.T) {
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "module.kasm")
	os.WriteFile(includePath, []byte("nop"), 0644)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for duplicate include")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "Duplicate %include") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `%include "` + includePath + `"
%include "` + includePath + `"`
	kasm.PreProcessingHandleIncludes(source)
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
	kasm.PreProcessingHandleIncludes(source)
}

func TestPreProcessingHandleIncludes_LineNumber(t *testing.T) {
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "module.kasm")
	os.WriteFile(includePath, []byte("nop"), 0644)

	source := `; line 1
; line 2
%include "` + includePath + `"`
	_, inclusions := kasm.PreProcessingHandleIncludes(source)

	if len(inclusions) != 1 {
		t.Fatalf("expected 1 inclusion, got %d", len(inclusions))
	}
	if inclusions[0].LineNumber != 3 {
		t.Errorf("expected line number 3, got %d", inclusions[0].LineNumber)
	}
}

func TestPreProcessingHandleIncludes_TrimWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	includePath := filepath.Join(tmpDir, "module.kasm")
	os.WriteFile(includePath, []byte("\n  mov rax, 1\n\n"), 0644)

	source := `%include "` + includePath + `"`
	result, _ := kasm.PreProcessingHandleIncludes(source)

	// Should not start with newline inside FILE block (trimmed)
	if !containsSubstring(result, "; FILE: "+includePath+"\nmov rax, 1\n; END FILE:") {
		t.Errorf("expected trimmed content in result, got:\n%s", result)
	}
}
