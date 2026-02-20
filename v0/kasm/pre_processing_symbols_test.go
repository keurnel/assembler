package kasm_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/keurnel/assembler/v0/kasm"
)

// --- PreProcessingCreateSymbolTable: %define ---

func TestPreProcessingCreateSymbolTable_SingleDefine(t *testing.T) {
	source := `%define DEBUG`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if !symbols["DEBUG"] {
		t.Error("expected 'DEBUG' to be defined")
	}
	if len(symbols) != 1 {
		t.Errorf("expected 1 symbol, got %d", len(symbols))
	}
}

func TestPreProcessingCreateSymbolTable_MultipleDefines(t *testing.T) {
	source := `%define DEBUG
%define VERBOSE
%define TRACE`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}

	for _, name := range []string{"DEBUG", "VERBOSE", "TRACE"} {
		if !symbols[name] {
			t.Errorf("expected '%s' to be defined", name)
		}
	}
}

func TestPreProcessingCreateSymbolTable_NoDefines(t *testing.T) {
	source := `mov rax, 1`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(symbols))
	}
}

func TestPreProcessingCreateSymbolTable_EmptySource(t *testing.T) {
	symbols := kasm.PreProcessingCreateSymbolTable("", nil)

	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(symbols))
	}
}

func TestPreProcessingCreateSymbolTable_DuplicateDefine_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for duplicate define")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "Duplicate %define") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	source := `%define DEBUG
%define DEBUG`
	kasm.PreProcessingCreateSymbolTable(source, nil)
}

func TestPreProcessingCreateSymbolTable_DuplicateDefine_ReportsLineNumbers(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for duplicate define")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsSubstring(msg, "at line 3") {
			t.Errorf("expected duplicate at line 3, got: %s", msg)
		}
		if !containsSubstring(msg, "first defined at line 1") {
			t.Errorf("expected first defined at line 1, got: %s", msg)
		}
	}()

	source := `%define DEBUG
; comment
%define DEBUG`
	kasm.PreProcessingCreateSymbolTable(source, nil)
}

// --- PreProcessingCreateSymbolTable: macros as symbols ---

func TestPreProcessingCreateSymbolTable_MacrosAddedAsSymbols(t *testing.T) {
	source := ``
	macroTable := map[string]kasm.Macro{
		"my_macro": {
			Name: "my_macro",
		},
	}
	symbols := kasm.PreProcessingCreateSymbolTable(source, macroTable)

	if !symbols["my_macro"] {
		t.Error("expected macro 'my_macro' to be in symbol table")
	}
}

func TestPreProcessingCreateSymbolTable_DefinesAndMacrosCombined(t *testing.T) {
	source := `%define DEBUG`
	macroTable := map[string]kasm.Macro{
		"my_macro": {
			Name: "my_macro",
		},
	}
	symbols := kasm.PreProcessingCreateSymbolTable(source, macroTable)

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}
	if !symbols["DEBUG"] {
		t.Error("expected 'DEBUG' to be defined")
	}
	if !symbols["my_macro"] {
		t.Error("expected 'my_macro' to be defined")
	}
}

func TestPreProcessingCreateSymbolTable_NilMacroTable(t *testing.T) {
	source := `%define FOO`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}
	if !symbols["FOO"] {
		t.Error("expected 'FOO' to be defined")
	}
}

func TestPreProcessingCreateSymbolTable_MultipleMacros(t *testing.T) {
	source := ``
	macroTable := map[string]kasm.Macro{
		"macro_a": {Name: "macro_a"},
		"macro_b": {Name: "macro_b"},
		"macro_c": {Name: "macro_c"},
	}
	symbols := kasm.PreProcessingCreateSymbolTable(source, macroTable)

	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}
	for _, name := range []string{"macro_a", "macro_b", "macro_c"} {
		if !symbols[name] {
			t.Errorf("expected '%s' to be defined", name)
		}
	}
}

// --- PreProcessingCreateSymbolTable: whitespace handling ---

func TestPreProcessingCreateSymbolTable_LeadingWhitespace(t *testing.T) {
	source := `   %define DEBUG`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if !symbols["DEBUG"] {
		t.Error("expected 'DEBUG' to be defined despite leading whitespace")
	}
}

func TestPreProcessingCreateSymbolTable_TabIndent(t *testing.T) {
	source := "\t%define DEBUG"
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if !symbols["DEBUG"] {
		t.Error("expected 'DEBUG' to be defined despite tab indent")
	}
}

// --- PreProcessingCreateSymbolTable: ignores non-define lines ---

func TestPreProcessingCreateSymbolTable_IgnoresComments(t *testing.T) {
	source := `; %define NOT_A_SYMBOL
%define REAL_SYMBOL`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}
	if !symbols["REAL_SYMBOL"] {
		t.Error("expected 'REAL_SYMBOL' to be defined")
	}
	if symbols["NOT_A_SYMBOL"] {
		t.Error("expected 'NOT_A_SYMBOL' to NOT be defined")
	}
}

func TestPreProcessingCreateSymbolTable_IgnoresInlineMacroDirectives(t *testing.T) {
	source := `%macro my_macro 1
    mov rax, %1
%endmacro
%define ENABLED`
	symbols := kasm.PreProcessingCreateSymbolTable(source, nil)

	if !symbols["ENABLED"] {
		t.Error("expected 'ENABLED' to be defined")
	}
	// macro directive is not a define directive
	if symbols["my_macro"] {
		t.Error("expected 'my_macro' to NOT be in symbol table (not via define)")
	}
}

func BenchmarkPreProcessingCreateSymbolTable_NoDefines(b *testing.B) {
	source := "mov rax, 1\nmov rdi, 0\nsyscall\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, nil)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_SingleDefine(b *testing.B) {
	source := "%define DEBUG\nmov rax, 1\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, nil)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_ManyDefines(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		sb.WriteString(fmt.Sprintf("%%define SYM_%d\n", i))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, nil)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_DefinesAtEnd(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 300; i++ {
		sb.WriteString(fmt.Sprintf("mov r%d, %d\n", i%16, i))
	}
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf("%%define SYM_%d\n", i))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, nil)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_WithMacroTable(b *testing.B) {
	source := "%define DEBUG\n%define RELEASE\n"
	macroTable := make(map[string]kasm.Macro, 20)
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("macro_%d", i)
		macroTable[name] = kasm.Macro{Name: name}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, macroTable)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_OnlyMacroTable(b *testing.B) {
	macroTable := make(map[string]kasm.Macro, 50)
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("macro_%d", i)
		macroTable[name] = kasm.Macro{Name: name}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable("", macroTable)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_LargeSource_NoDefines(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(fmt.Sprintf("mov r%d, %d\n", i%16, i))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, nil)
	}
}

func BenchmarkPreProcessingCreateSymbolTable_ManyDefinesWithMacros(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 30; i++ {
		sb.WriteString(fmt.Sprintf("%%define SYM_%d\n", i))
	}
	source := sb.String()
	macroTable := make(map[string]kasm.Macro, 30)
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("macro_%d", i)
		macroTable[name] = kasm.Macro{Name: name}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kasm.PreProcessingCreateSymbolTable(source, macroTable)
	}
}
