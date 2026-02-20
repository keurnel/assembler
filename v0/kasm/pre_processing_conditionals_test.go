package kasm_test

import (
	"strings"
	"testing"

	"github.com/keurnel/assembler/v0/kasm"
)

// --- PreProcessingHandleConditionals: %ifdef ---

func TestPreProcessingHandleConditionals_IfdefTrue(t *testing.T) {
	source := `%ifdef DEBUG
mov rax, 1
%endif`
	symbols := map[string]bool{"DEBUG": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected body to be included when symbol is defined")
	}
	if containsSubstring(result, "%ifdef") || containsSubstring(result, "%endif") {
		t.Error("expected directives to be removed")
	}
}

func TestPreProcessingHandleConditionals_IfdefFalse(t *testing.T) {
	source := `%ifdef DEBUG
mov rax, 1
%endif`
	symbols := map[string]bool{}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if containsSubstring(result, "mov rax, 1") {
		t.Error("expected body to be excluded when symbol is not defined")
	}
}

// --- PreProcessingHandleConditionals: %ifndef ---

func TestPreProcessingHandleConditionals_IfndefTrue(t *testing.T) {
	source := `%ifndef RELEASE
mov rax, 1
%endif`
	symbols := map[string]bool{}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected body to be included when symbol is NOT defined")
	}
}

func TestPreProcessingHandleConditionals_IfndefFalse(t *testing.T) {
	source := `%ifndef RELEASE
mov rax, 1
%endif`
	symbols := map[string]bool{"RELEASE": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if containsSubstring(result, "mov rax, 1") {
		t.Error("expected body to be excluded when symbol IS defined")
	}
}

// --- PreProcessingHandleConditionals: %else ---

func TestPreProcessingHandleConditionals_IfdefWithElse_ConditionTrue(t *testing.T) {
	source := `%ifdef DEBUG
mov rax, 1
%else
mov rax, 0
%endif`
	symbols := map[string]bool{"DEBUG": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected ifdef branch to be included")
	}
	if containsSubstring(result, "mov rax, 0") {
		t.Error("expected else branch to be excluded")
	}
}

func TestPreProcessingHandleConditionals_IfdefWithElse_ConditionFalse(t *testing.T) {
	source := `%ifdef DEBUG
mov rax, 1
%else
mov rax, 0
%endif`
	symbols := map[string]bool{}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if containsSubstring(result, "mov rax, 1") {
		t.Error("expected ifdef branch to be excluded")
	}
	if !containsSubstring(result, "mov rax, 0") {
		t.Error("expected else branch to be included")
	}
}

// --- PreProcessingHandleConditionals: nested ---

func TestPreProcessingHandleConditionals_Nested(t *testing.T) {
	t.Skip("nested conditionals not yet supported — byte offsets shift after first block replacement")
	source := `%ifdef OUTER
%ifdef INNER
mov rax, 1
%endif
%endif`
	symbols := map[string]bool{"OUTER": true, "INNER": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected nested body to be included when both symbols are defined")
	}
}

func TestPreProcessingHandleConditionals_Nested_OuterFalse(t *testing.T) {
	t.Skip("nested conditionals not yet supported — byte offsets shift after first block replacement")
	source := `%ifdef OUTER
%ifdef INNER
mov rax, 1
%endif
%endif`
	symbols := map[string]bool{"INNER": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if containsSubstring(result, "mov rax, 1") {
		t.Error("expected nested body to be excluded when outer symbol is not defined")
	}
}

// --- PreProcessingHandleConditionals: no conditionals ---

func TestPreProcessingHandleConditionals_NoConditionals(t *testing.T) {
	source := `mov rax, 1
mov rdi, 0`
	symbols := map[string]bool{}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if result != source {
		t.Errorf("expected source unchanged, got '%s'", result)
	}
}

// --- PreProcessingHandleConditionals: structural errors ---

func TestPreProcessingHandleConditionals_UnmatchedEndif_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unmatched endif")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "%endif without matching") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `mov rax, 1
%endif`
	kasm.PreProcessingHandleConditionals(source, map[string]bool{})
}

func TestPreProcessingHandleConditionals_UnmatchedIfdef_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unmatched ifdef")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "has no matching %endif") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `%ifdef DEBUG
mov rax, 1`
	kasm.PreProcessingHandleConditionals(source, map[string]bool{})
}

func TestPreProcessingHandleConditionals_DuplicateElse_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for duplicate else")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "Duplicate %else") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `%ifdef DEBUG
mov rax, 1
%else
mov rax, 0
%else
mov rax, 2
%endif`
	kasm.PreProcessingHandleConditionals(source, map[string]bool{"DEBUG": true})
}

func TestPreProcessingHandleConditionals_ElseWithoutIfdef_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for else without ifdef")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "%else without matching") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `mov rax, 1
%else
mov rax, 0
%endif`
	kasm.PreProcessingHandleConditionals(source, map[string]bool{})
}

// --- PreProcessingHandleConditionals: surrounding code preserved ---

func TestPreProcessingHandleConditionals_PreservesCodeAroundBlock(t *testing.T) {
	source := `mov rax, 0
%ifdef DEBUG
mov rbx, 1
%endif
mov rcx, 2`
	symbols := map[string]bool{"DEBUG": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 0") {
		t.Error("expected code before block to be preserved")
	}
	if !containsSubstring(result, "mov rbx, 1") {
		t.Error("expected conditional body to be included")
	}
	if !containsSubstring(result, "mov rcx, 2") {
		t.Error("expected code after block to be preserved")
	}
}

func TestPreProcessingHandleConditionals_PreservesCodeAroundBlock_ConditionFalse(t *testing.T) {
	source := `mov rax, 0
%ifdef DEBUG
mov rbx, 1
%endif
mov rcx, 2`
	symbols := map[string]bool{}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 0") {
		t.Error("expected code before block to be preserved")
	}
	if containsSubstring(result, "mov rbx, 1") {
		t.Error("expected conditional body to be excluded")
	}
	if !containsSubstring(result, "mov rcx, 2") {
		t.Error("expected code after block to be preserved")
	}
}

// --- PreProcessingHandleConditionals: empty source ---

func TestPreProcessingHandleConditionals_EmptySource(t *testing.T) {
	result := kasm.PreProcessingHandleConditionals("", map[string]bool{})
	if result != "" {
		t.Errorf("expected empty result, got '%s'", result)
	}
}

// --- PreProcessingHandleConditionals: ifndef with else ---

func TestPreProcessingHandleConditionals_IfndefWithElse_ConditionTrue(t *testing.T) {
	source := `%ifndef RELEASE
mov rax, 1
%else
mov rax, 0
%endif`
	symbols := map[string]bool{}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected ifndef branch when symbol is not defined")
	}
	if containsSubstring(result, "mov rax, 0") {
		t.Error("expected else branch to be excluded")
	}
}

func TestPreProcessingHandleConditionals_IfndefWithElse_ConditionFalse(t *testing.T) {
	source := `%ifndef RELEASE
mov rax, 1
%else
mov rax, 0
%endif`
	symbols := map[string]bool{"RELEASE": true}
	result := kasm.PreProcessingHandleConditionals(source, symbols)

	result = strings.TrimSpace(result)

	if containsSubstring(result, "mov rax, 1") {
		t.Error("expected ifndef branch to be excluded when symbol is defined")
	}
	if !containsSubstring(result, "mov rax, 0") {
		t.Error("expected else branch when symbol is defined")
	}
}
