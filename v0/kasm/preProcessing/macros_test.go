package preProcessing_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/keurnel/assembler/v0/kasm/preProcessing"
)

// --- PreProcessingHasMacros ---

func TestPreProcessingHasMacros_WithMacro(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro`
	if !preProcessing.HasMacros(source) {
		t.Error("expected true for source containing a macro definition")
	}
}

func TestPreProcessingHasMacros_WithoutMacro(t *testing.T) {
	source := `mov rax, 1
mov rdi, 2`
	if preProcessing.HasMacros(source) {
		t.Error("expected false for source without macro definitions")
	}
}

func TestPreProcessingHasMacros_EmptySource(t *testing.T) {
	if preProcessing.HasMacros("") {
		t.Error("expected false for empty source")
	}
}

func TestPreProcessingHasMacros_MacroWithoutParams(t *testing.T) {
	source := `%macro nop_macro
    nop
%endmacro`
	if !preProcessing.HasMacros(source) {
		t.Error("expected true for macro without parameter count")
	}
}

// --- PreProcessingMacroTable ---

func TestPreProcessingMacroTable_SingleMacro(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro`
	table := preProcessing.MacroTable(source)

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
	table := preProcessing.MacroTable(source)

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
	table := preProcessing.MacroTable(source)

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
	table := preProcessing.MacroTable(source)
	macro := table["test_macro"]

	expectedParams := []string{"paramA", "paramB", "paramC"}
	for _, name := range expectedParams {
		if _, ok := macro.Parameters[name]; !ok {
			t.Errorf("expected parameter '%s' to be present", name)
		}
	}
}

// --- PreProcessingCollectMacroCalls ---

func TestPreProcessingCollectMacroCalls_SingleCall(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2`

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)

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

func TestPreProcessingCollectMacroCalls_MultipleCalls(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2
my_macro 3, 4`

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)

	macro := table["my_macro"]
	if len(macro.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(macro.Calls))
	}
}

func TestPreProcessingCollectMacroCalls_WrongArgCount_Panics(t *testing.T) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1`

	table := preProcessing.MacroTable(source)

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

	preProcessing.CollectMacroCalls(source, table)
}

func TestPreProcessingCollectMacroCalls_LineNumber(t *testing.T) {
	source := `%macro my_macro 1
    mov rax, %1
%endmacro
; comment line
my_macro 42`

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)

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

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	result := preProcessing.ReplaceMacroCalls(source, table)

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

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	result := preProcessing.ReplaceMacroCalls(source, table)

	if containsSubstring(result, "    mov rax, 42") {
		t.Error("expected leading indentation to be stripped from expanded body")
	}
	if !containsSubstring(result, "mov rax, 42") {
		t.Error("expected 'mov rax, 42' in expanded output")
	}
}

// --- FR-2.2.6: %macro without %endmacro ---

func TestPreProcessingMacroTable_NoEndmacro_Panics(t *testing.T) {
	source := `%macro my_macro 1
    mov rax, %1
`
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for macro without matching endmacro")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatal("expected panic message to be a string")
		}
		if !containsSubstring(msg, "my_macro") {
			t.Errorf("expected panic message to contain macro name, got: %s", msg)
		}
		if !containsSubstring(msg, "endmacro") {
			t.Errorf("expected panic message to mention endmacro, got: %s", msg)
		}
	}()
	preProcessing.MacroTable(source)
}

// --- FR-2.5: Macro definition removal ---

func TestPreProcessingReplaceMacroCalls_RemovesDefinitionBlock(t *testing.T) {
	source := `%macro my_macro 1
    mov rax, %1
%endmacro
my_macro 42`

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	result := preProcessing.ReplaceMacroCalls(source, table)

	if containsSubstring(result, "%macro") {
		t.Error("expected macro definition to be removed from output")
	}
	if containsSubstring(result, "%endmacro") {
		t.Error("expected endmacro to be removed from output")
	}
	if !containsSubstring(result, "mov rax, 42") {
		t.Error("expected expanded macro call in output")
	}
}

func TestPreProcessingReplaceMacroCalls_RemovesUnusedDefinition(t *testing.T) {
	source := `%macro unused_macro 1
    mov rax, %1
%endmacro
mov rbx, 1`

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	result := preProcessing.ReplaceMacroCalls(source, table)

	if containsSubstring(result, "%macro") {
		t.Error("expected unused macro definition to be removed (FR-2.5.3)")
	}
	if containsSubstring(result, "%endmacro") {
		t.Error("expected endmacro to be removed")
	}
	if !containsSubstring(result, "mov rbx, 1") {
		t.Error("expected non-macro code to be preserved")
	}
}

func TestPreProcessingReplaceMacroCalls_RemovesMultipleDefinitions(t *testing.T) {
	source := `%macro mac_a 1
    mov rax, %1
%endmacro
%macro mac_b 1
    mov rbx, %1
%endmacro
mac_a 1
mac_b 2`

	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	result := preProcessing.ReplaceMacroCalls(source, table)

	if containsSubstring(result, "%macro") {
		t.Error("expected all macro definitions to be removed")
	}
	if !containsSubstring(result, "mov rax, 1") {
		t.Error("expected mac_a expansion")
	}
	if !containsSubstring(result, "mov rbx, 2") {
		t.Error("expected mac_b expansion")
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

// --- PreProcessingHasMacros ---

func BenchmarkPreProcessingHasMacros_NoMacros(b *testing.B) {
	source := "mov rax, 1\nmov rdi, 0\nsyscall\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HasMacros(source)
	}
}

func BenchmarkPreProcessingHasMacros_WithMacro(b *testing.B) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HasMacros(source)
	}
}

func BenchmarkPreProcessingHasMacros_LargeSource_NoMacro(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 500; i++ {
		sb.WriteString(fmt.Sprintf("mov r%d, %d\n", i%16, i))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HasMacros(source)
	}
}

func BenchmarkPreProcessingHasMacros_LargeSource_MacroAtEnd(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 500; i++ {
		sb.WriteString(fmt.Sprintf("mov r%d, %d\n", i%16, i))
	}
	sb.WriteString("%macro tail_macro 1\n    mov rax, %1\n%endmacro\n")
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.HasMacros(source)
	}
}

// --- PreProcessingMacroTable ---

func BenchmarkPreProcessingMacroTable_NoMacros(b *testing.B) {
	source := "mov rax, 1\nmov rdi, 0\nsyscall\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.MacroTable(source)
	}
}

func BenchmarkPreProcessingMacroTable_SingleMacro(b *testing.B) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.MacroTable(source)
	}
}

func BenchmarkPreProcessingMacroTable_ManyMacros(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString(fmt.Sprintf("%%macro macro_%d 2\n    mov rax, %%1\n    mov rbx, %%2\n%%endmacro\n", i))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.MacroTable(source)
	}
}

func BenchmarkPreProcessingMacroTable_MacroWithLargeBody(b *testing.B) {
	var body strings.Builder
	body.WriteString("%macro big_macro 3\n")
	for i := 0; i < 100; i++ {
		body.WriteString(fmt.Sprintf("    mov r%d, %%1\n", i%16))
	}
	body.WriteString("%endmacro\n")
	source := body.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.MacroTable(source)
	}
}

// --- PreProcessingCollectMacroCalls ---

func BenchmarkPreProcessingCollectMacroCalls_NoCalls(b *testing.B) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
mov rax, 1`
	table := preProcessing.MacroTable(source)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset calls each iteration by rebuilding the table
		t2 := preProcessing.MacroTable(source)
		preProcessing.CollectMacroCalls(source, t2)
		_ = t2
	}
	_ = table
}

func BenchmarkPreProcessingCollectMacroCalls_SingleCall(b *testing.B) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table := preProcessing.MacroTable(source)
		preProcessing.CollectMacroCalls(source, table)
	}
}

func BenchmarkPreProcessingCollectMacroCalls_ManyCalls(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("%macro my_macro 2\n    mov rax, %1\n    mov rdi, %2\n%endmacro\n")
	for i := 0; i < 50; i++ {
		sb.WriteString(fmt.Sprintf("my_macro %d, %d\n", i, i+1))
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table := preProcessing.MacroTable(source)
		preProcessing.CollectMacroCalls(source, table)
	}
}

func BenchmarkPreProcessingCollectMacroCalls_MultipleMacros(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		sb.WriteString(fmt.Sprintf("%%macro mac_%d 1\n    mov rax, %%1\n%%endmacro\n", i))
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			sb.WriteString(fmt.Sprintf("mac_%d %d\n", i, j))
		}
	}
	source := sb.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table := preProcessing.MacroTable(source)
		preProcessing.CollectMacroCalls(source, table)
	}
}

// --- PreProcessingReplaceMacroCalls ---

func BenchmarkPreProcessingReplaceMacroCalls_SingleCall(b *testing.B) {
	source := `%macro my_macro 2
    mov rax, %1
    mov rdi, %2
%endmacro
my_macro 1, 2`
	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.ReplaceMacroCalls(source, table)
	}
}

func BenchmarkPreProcessingReplaceMacroCalls_ManyCalls(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("%macro my_macro 2\n    mov rax, %1\n    mov rdi, %2\n%endmacro\n")
	for i := 0; i < 50; i++ {
		sb.WriteString(fmt.Sprintf("my_macro %d, %d\n", i, i+1))
	}
	source := sb.String()
	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.ReplaceMacroCalls(source, table)
	}
}

func BenchmarkPreProcessingReplaceMacroCalls_LargeBody(b *testing.B) {
	var body strings.Builder
	body.WriteString("%macro big_macro 2\n")
	for i := 0; i < 100; i++ {
		body.WriteString(fmt.Sprintf("    mov r%d, %%1\n    add r%d, %%2\n", i%16, i%16))
	}
	body.WriteString("%endmacro\nbig_macro rax, rbx\n")
	source := body.String()
	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.ReplaceMacroCalls(source, table)
	}
}

func BenchmarkPreProcessingReplaceMacroCalls_MultipleMacros(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		sb.WriteString(fmt.Sprintf("%%macro mac_%d 1\n    mov rax, %%1\n%%endmacro\n", i))
	}
	for i := 0; i < 5; i++ {
		sb.WriteString(fmt.Sprintf("mac_%d %d\n", i, i*10))
	}
	source := sb.String()
	table := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, table)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preProcessing.ReplaceMacroCalls(source, table)
	}
}
