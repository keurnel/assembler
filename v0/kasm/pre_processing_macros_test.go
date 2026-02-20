package kasm_test

import (
	"testing"

	"github.com/keurnel/assembler/v0/kasm"
)

// --- PreProcessingHasMacros ---

func TestPreProcessingHasMacros_WithMacro(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro`
	if !kasm.PreProcessingHasMacros(source) {
		t.Error("expected true for source containing a macro definition")
	}
}

func TestPreProcessingHasMacros_WithoutMacro(t *testing.T) {
	source := `mov rax, 1
mov rdi, 2`
	if kasm.PreProcessingHasMacros(source) {
		t.Error("expected false for source without macro definitions")
	}
}

func TestPreProcessingHasMacros_EmptySource(t *testing.T) {
	if kasm.PreProcessingHasMacros("") {
		t.Error("expected false for empty source")
	}
}

func TestPreProcessingHasMacros_MacroWithoutParams(t *testing.T) {
	source := `%macro nop_macro
    nop
%endmacro`
	if !kasm.PreProcessingHasMacros(source) {
		t.Error("expected true for macro without parameter count")
	}
}

// --- PreProcessingMacroTable ---

func TestPreProcessingMacroTable_SingleMacro(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro`
	table := kasm.PreProcessingMacroTable(source)

	if len(table) != 1 {
		t.Fatalf("expected 1 macro, got %d", len(table))
	}

	macro, ok := table["my_macro"]
	if !ok {
		t.Fatal("expected macro 'my_macro' to be present")
	}

	if macro.Name != "my_macro" {
		t.Errorf("expected name 'my_macro', got '%s'", macro.Name)
	}

	if len(macro.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(macro.Parameters))
	}

	if macro.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestPreProcessingMacroTable_MultipleMacros(t *testing.T) {
	source := `%macro macro_a 1
    mov rax, %1
%endmacro
%macro macro_b 3
    mov rax, %1
    mov rbx, %2
    mov rcx, %3
%endmacro`
	table := kasm.PreProcessingMacroTable(source)

	if len(table) != 2 {
		t.Fatalf("expected 2 macros, got %d", len(table))
	}

	if _, ok := table["macro_a"]; !ok {
		t.Error("expected macro 'macro_a' to be present")
	}
	if _, ok := table["macro_b"]; !ok {
		t.Error("expected macro 'macro_b' to be present")
	}

	if len(table["macro_a"].Parameters) != 1 {
		t.Errorf("expected 1 parameter for macro_a, got %d", len(table["macro_a"].Parameters))
	}
	if len(table["macro_b"].Parameters) != 3 {
		t.Errorf("expected 3 parameters for macro_b, got %d", len(table["macro_b"].Parameters))
	}
}

func TestPreProcessingMacroTable_NoMacros(t *testing.T) {
	source := `mov rax, 1`
	table := kasm.PreProcessingMacroTable(source)

	if len(table) != 0 {
		t.Errorf("expected 0 macros, got %d", len(table))
	}
}

func TestPreProcessingMacroTable_ParameterNaming(t *testing.T) {
	source := `%macro test_macro 3
    mov rax, %1
    mov rbx, %2
    mov rcx, %3
%endmacro`
	table := kasm.PreProcessingMacroTable(source)
	macro := table["test_macro"]

	expectedParams := []string{"paramA", "paramB", "paramC"}
	for _, name := range expectedParams {
		if _, ok := macro.Parameters[name]; !ok {
			t.Errorf("expected parameter '%s' to be present", name)
		}
	}
}

// --- PreProcessingColectMacroCalls ---

func TestPreProcessingColectMacroCalls_SingleCall(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2`

	table := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingColectMacroCalls(source, table)

	macro := table["my_macro"]
	if len(macro.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(macro.Calls))
	}

	call := macro.Calls[0]
	if call.Name != "my_macro" {
		t.Errorf("expected call name 'my_macro', got '%s'", call.Name)
	}
	if len(call.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(call.Arguments))
	}
	if call.Arguments[0] != "1" || call.Arguments[1] != "2" {
		t.Errorf("expected arguments ['1', '2'], got %v", call.Arguments)
	}
}

func TestPreProcessingColectMacroCalls_MultipleCalls(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2
my_macro 3, 4`

	table := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingColectMacroCalls(source, table)

	macro := table["my_macro"]
	if len(macro.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(macro.Calls))
	}
}

func TestPreProcessingColectMacroCalls_WrongArgCount_Panics(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1`

	table := kasm.PreProcessingMacroTable(source)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for wrong argument count")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "expects 2 arguments, but got 1") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	kasm.PreProcessingColectMacroCalls(source, table)
}

func TestPreProcessingColectMacroCalls_LineNumber(t *testing.T) {
	source := `%macro my_macro 1
    mov rax, %1
%endmacro
; comment line
my_macro 42`

	table := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingColectMacroCalls(source, table)

	macro := table["my_macro"]
	if len(macro.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(macro.Calls))
	}
	if macro.Calls[0].LineNumber != 5 {
		t.Errorf("expected line number 5, got %d", macro.Calls[0].LineNumber)
	}
}

// --- PreProcessingReplaceMacroCalls ---

func TestPreProcessingReplaceMacroCalls_BasicExpansion(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2`

	table := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingColectMacroCalls(source, table)
	result := kasm.PreProcessingReplaceMacroCalls(source, table)

	if containsSubstring(result, "my_macro 1, 2") {
		t.Error("expected macro call to be replaced")
	}
	if !containsSubstring(result, "; MACRO: my_macro") {
		t.Error("expected macro comment in expanded output")
	}
	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected %1 to be replaced with '1'")
	}
	if !containsSubstring(result, "mov rdi, 2") {
		t.Error("expected %2 to be replaced with '2'")
	}
}

func TestPreProcessingReplaceMacroCalls_StripsIndentation(t *testing.T) {
	source := `%macro my_macro 1
    mov rax, %1
%endmacro
my_macro 42`

	table := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingColectMacroCalls(source, table)
	result := kasm.PreProcessingReplaceMacroCalls(source, table)

	if containsSubstring(result, "    mov rax, 42") {
		t.Error("expected leading indentation to be stripped from expanded body")
	}
	if !containsSubstring(result, "mov rax, 42") {
		t.Error("expected 'mov rax, 42' in expanded output")
	}
}

// --- helpers ---

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
