package kasm_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/keurnel/assembler/v0/kasm"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// x86Profile is the default profile used by all tests unless a specific
// architecture edge case requires a custom profile.
var x86Profile = kasm.NewX8664Profile()

func requireTokenCount(t *testing.T, tokens []kasm.Token, expected int) {
	t.Helper()
	if len(tokens) != expected {
		t.Fatalf("expected %d tokens, got %d: %v", expected, len(tokens), tokens)
	}
}

func requireToken(t *testing.T, tok kasm.Token, expectedType kasm.TokenType, expectedLiteral string) {
	t.Helper()
	if tok.Type != expectedType {
		t.Errorf("expected token type %d, got %d (literal=%q)", expectedType, tok.Type, tok.Literal)
	}
	if tok.Literal != expectedLiteral {
		t.Errorf("expected literal %q, got %q", expectedLiteral, tok.Literal)
	}
}

// ---------------------------------------------------------------------------
// Tests: empty / whitespace-only input
// ---------------------------------------------------------------------------

func TestLexer_EmptyInput(t *testing.T) {
	tokens := kasm.LexerNew("", x86Profile).Start()
	requireTokenCount(t, tokens, 0)
}

func TestLexer_WhitespaceOnly(t *testing.T) {
	tokens := kasm.LexerNew("   \t\n\r\n  ", x86Profile).Start()
	requireTokenCount(t, tokens, 0)
}

// ---------------------------------------------------------------------------
// Tests: comments
// ---------------------------------------------------------------------------

func TestLexer_Comment(t *testing.T) {
	tokens := kasm.LexerNew("; this is a comment", x86Profile).Start()
	requireTokenCount(t, tokens, 0)
}

func TestLexer_CommentAfterNewline(t *testing.T) {
	tokens := kasm.LexerNew("\n; line two comment", x86Profile).Start()
	requireTokenCount(t, tokens, 0)
}

// ---------------------------------------------------------------------------
// Tests: directives
// ---------------------------------------------------------------------------

func TestLexer_Directive(t *testing.T) {
	tokens := kasm.LexerNew("%define", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenDirective, "%define")
}

func TestLexer_DirectiveInclude(t *testing.T) {
	tokens := kasm.LexerNew(`%include "file.kasm"`, x86Profile).Start()
	requireTokenCount(t, tokens, 2)
	requireToken(t, tokens[0], kasm.TokenDirective, "%include")
	requireToken(t, tokens[1], kasm.TokenString, "file.kasm")
}

// ---------------------------------------------------------------------------
// Tests: instructions
// ---------------------------------------------------------------------------

func TestLexer_InstructionMov(t *testing.T) {
	tokens := kasm.LexerNew("mov", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenInstruction, "mov")
}

func TestLexer_InstructionUpperCase(t *testing.T) {
	tokens := kasm.LexerNew("MOV", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenInstruction, "MOV")
}

func TestLexer_InstructionSyscall(t *testing.T) {
	tokens := kasm.LexerNew("syscall", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenInstruction, "syscall")
}

func TestLexer_MultipleInstructions(t *testing.T) {
	mnemonics := []string{
		"mov", "add", "sub", "push", "pop", "call", "ret", "jmp",
		"je", "jne", "cmp", "test", "xor", "nop", "hlt", "lea",
		"inc", "dec", "mul", "div", "shl", "shr", "and", "or",
	}
	for _, m := range mnemonics {
		t.Run(m, func(t *testing.T) {
			tokens := kasm.LexerNew(m, x86Profile).Start()
			requireTokenCount(t, tokens, 1)
			requireToken(t, tokens[0], kasm.TokenInstruction, m)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: registers
// ---------------------------------------------------------------------------

func TestLexer_Register64(t *testing.T) {
	regs := []string{"rax", "rbx", "rcx", "rdx", "rsi", "rdi", "rbp", "rsp"}
	for _, r := range regs {
		t.Run(r, func(t *testing.T) {
			tokens := kasm.LexerNew(r, x86Profile).Start()
			requireTokenCount(t, tokens, 1)
			requireToken(t, tokens[0], kasm.TokenRegister, r)
		})
	}
}

func TestLexer_Register32(t *testing.T) {
	tokens := kasm.LexerNew("eax", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenRegister, "eax")
}

func TestLexer_Register8(t *testing.T) {
	tokens := kasm.LexerNew("al", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenRegister, "al")
}

func TestLexer_RegisterExtended(t *testing.T) {
	for i := 8; i <= 15; i++ {
		name := fmt.Sprintf("r%d", i)
		t.Run(name, func(t *testing.T) {
			tokens := kasm.LexerNew(name, x86Profile).Start()
			requireTokenCount(t, tokens, 1)
			requireToken(t, tokens[0], kasm.TokenRegister, name)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: immediate values
// ---------------------------------------------------------------------------

func TestLexer_ImmediateDecimal(t *testing.T) {
	tokens := kasm.LexerNew("42", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenImmediate, "42")
}

func TestLexer_ImmediateHex(t *testing.T) {
	tokens := kasm.LexerNew("0xFF", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenImmediate, "0xFF")
}

func TestLexer_ImmediateHexUpper(t *testing.T) {
	tokens := kasm.LexerNew("0XAB", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenImmediate, "0XAB")
}

func TestLexer_ImmediateZero(t *testing.T) {
	tokens := kasm.LexerNew("0", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenImmediate, "0")
}

// ---------------------------------------------------------------------------
// Tests: string literals
// ---------------------------------------------------------------------------

func TestLexer_StringLiteral(t *testing.T) {
	tokens := kasm.LexerNew(`"Hello, World!"`, x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenString, "Hello, World!")
}

func TestLexer_EmptyString(t *testing.T) {
	tokens := kasm.LexerNew(`""`, x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenString, "")
}

// ---------------------------------------------------------------------------
// Tests: identifiers and labels
// ---------------------------------------------------------------------------

func TestLexer_Identifier(t *testing.T) {
	tokens := kasm.LexerNew("my_variable", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenIdentifier, "my_variable")
}

func TestLexer_Label(t *testing.T) {
	tokens := kasm.LexerNew("_start:", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenIdentifier, "_start:")
}

func TestLexer_DotPrefixedLabel(t *testing.T) {
	tokens := kasm.LexerNew(".loop:", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenIdentifier, ".loop:")
}

func TestLexer_IdentifierWithDot(t *testing.T) {
	tokens := kasm.LexerNew("section.text", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenIdentifier, "section.text")
}

// ---------------------------------------------------------------------------
// Tests: punctuation / single characters
// ---------------------------------------------------------------------------

func TestLexer_Comma(t *testing.T) {
	tokens := kasm.LexerNew(",", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenIdentifier, ",")
}

func TestLexer_Brackets(t *testing.T) {
	tokens := kasm.LexerNew("[rbp]", x86Profile).Start()
	requireTokenCount(t, tokens, 3)
	requireToken(t, tokens[0], kasm.TokenIdentifier, "[")
	requireToken(t, tokens[1], kasm.TokenRegister, "rbp")
	requireToken(t, tokens[2], kasm.TokenIdentifier, "]")
}

// ---------------------------------------------------------------------------
// Tests: full lines / multi-token sequences
// ---------------------------------------------------------------------------

func TestLexer_MovRegImm(t *testing.T) {
	tokens := kasm.LexerNew("mov rax, 1", x86Profile).Start()
	requireTokenCount(t, tokens, 4)
	requireToken(t, tokens[0], kasm.TokenInstruction, "mov")
	requireToken(t, tokens[1], kasm.TokenRegister, "rax")
	requireToken(t, tokens[2], kasm.TokenIdentifier, ",")
	requireToken(t, tokens[3], kasm.TokenImmediate, "1")
}

func TestLexer_MovRegReg(t *testing.T) {
	tokens := kasm.LexerNew("mov rax, rbx", x86Profile).Start()
	requireTokenCount(t, tokens, 4)
	requireToken(t, tokens[0], kasm.TokenInstruction, "mov")
	requireToken(t, tokens[1], kasm.TokenRegister, "rax")
	requireToken(t, tokens[2], kasm.TokenIdentifier, ",")
	requireToken(t, tokens[3], kasm.TokenRegister, "rbx")
}

func TestLexer_InstructionWithComment(t *testing.T) {
	tokens := kasm.LexerNew("mov rax, 1 ; set exit code", x86Profile).Start()
	requireTokenCount(t, tokens, 4)
	requireToken(t, tokens[0], kasm.TokenInstruction, "mov")
	requireToken(t, tokens[1], kasm.TokenRegister, "rax")
	requireToken(t, tokens[2], kasm.TokenIdentifier, ",")
	requireToken(t, tokens[3], kasm.TokenImmediate, "1")
}

func TestLexer_LabelFollowedByInstruction(t *testing.T) {
	tokens := kasm.LexerNew("_start:\n    mov rax, 60", x86Profile).Start()
	requireTokenCount(t, tokens, 5)
	requireToken(t, tokens[0], kasm.TokenIdentifier, "_start:")
	requireToken(t, tokens[1], kasm.TokenInstruction, "mov")
	requireToken(t, tokens[2], kasm.TokenRegister, "rax")
	requireToken(t, tokens[3], kasm.TokenIdentifier, ",")
	requireToken(t, tokens[4], kasm.TokenImmediate, "60")
}

func TestLexer_MemoryOperand(t *testing.T) {
	tokens := kasm.LexerNew("mov [rbp], rax", x86Profile).Start()
	requireTokenCount(t, tokens, 6)
	requireToken(t, tokens[0], kasm.TokenInstruction, "mov")
	requireToken(t, tokens[1], kasm.TokenIdentifier, "[")
	requireToken(t, tokens[2], kasm.TokenRegister, "rbp")
	requireToken(t, tokens[3], kasm.TokenIdentifier, "]")
	requireToken(t, tokens[4], kasm.TokenIdentifier, ",")
	requireToken(t, tokens[5], kasm.TokenRegister, "rax")
}

// ---------------------------------------------------------------------------
// Tests: multi-line programs
// ---------------------------------------------------------------------------

func TestLexer_MultiLineProgram(t *testing.T) {
	source := `_start:
    mov rax, 60
    xor rdi, rdi
    syscall`

	tokens := kasm.LexerNew(source, x86Profile).Start()

	// _start: + mov rax , 60 + xor rdi , rdi + syscall = 1 + 4 + 4 + 1 = 10
	requireTokenCount(t, tokens, 10)
	requireToken(t, tokens[0], kasm.TokenIdentifier, "_start:")
	requireToken(t, tokens[9], kasm.TokenInstruction, "syscall")
}

func TestLexer_ProgramWithDirectiveAndMacro(t *testing.T) {
	source := `%include "hulp.kasm"
%define DEBUG
_start:
    mov rax, 1`

	tokens := kasm.LexerNew(source, x86Profile).Start()

	requireToken(t, tokens[0], kasm.TokenDirective, "%include")
	requireToken(t, tokens[1], kasm.TokenString, "hulp.kasm")
	requireToken(t, tokens[2], kasm.TokenDirective, "%define")
	requireToken(t, tokens[3], kasm.TokenIdentifier, "DEBUG")
}

// ---------------------------------------------------------------------------
// Tests: line and column tracking
// ---------------------------------------------------------------------------

func TestLexer_LineTracking(t *testing.T) {
	source := "mov rax, 1\nadd rbx, 2"
	tokens := kasm.LexerNew(source, x86Profile).Start()

	// First line tokens should be line 1
	if tokens[0].Line != 1 {
		t.Errorf("expected line 1, got %d", tokens[0].Line)
	}

	// Second line tokens should be line 2
	last := tokens[len(tokens)-1]
	if last.Line != 2 {
		t.Errorf("expected line 2, got %d", last.Line)
	}
}

func TestLexer_ColumnTracking(t *testing.T) {
	tokens := kasm.LexerNew("mov rax", x86Profile).Start()
	requireTokenCount(t, tokens, 2)
	if tokens[0].Column != 1 {
		t.Errorf("expected column 1 for 'mov', got %d", tokens[0].Column)
	}
	if tokens[1].Column != 5 {
		t.Errorf("expected column 5 for 'rax', got %d", tokens[1].Column)
	}
}

// ---------------------------------------------------------------------------
// Tests: edge cases
// ---------------------------------------------------------------------------

func TestLexer_UnterminatedString(t *testing.T) {
	// Should not panic; reads until EOF
	tokens := kasm.LexerNew(`"unterminated`, x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenString, "unterminated")
}

func TestLexer_DirectiveAlone(t *testing.T) {
	tokens := kasm.LexerNew("%endif", x86Profile).Start()
	requireTokenCount(t, tokens, 1)
	requireToken(t, tokens[0], kasm.TokenDirective, "%endif")
}

func TestLexer_ConsecutiveComments(t *testing.T) {
	source := "; first\n; second"
	tokens := kasm.LexerNew(source, x86Profile).Start()
	requireTokenCount(t, tokens, 0)
}

func TestLexer_HexImmediateInInstruction(t *testing.T) {
	tokens := kasm.LexerNew("mov rax, 0xDEAD", x86Profile).Start()
	requireTokenCount(t, tokens, 4)
	requireToken(t, tokens[3], kasm.TokenImmediate, "0xDEAD")
}

func TestLexer_UseInstruction(t *testing.T) {
	tokens := kasm.LexerNew("use mymodule", x86Profile).Start()
	requireTokenCount(t, tokens, 2)
	requireToken(t, tokens[0], kasm.TokenInstruction, "use")
	requireToken(t, tokens[1], kasm.TokenIdentifier, "mymodule")
}

func TestLexer_MacroDirectives(t *testing.T) {
	source := `%macro een_macro 2
    mov rax, %1
%endmacro`
	tokens := kasm.LexerNew(source, x86Profile).Start()

	requireToken(t, tokens[0], kasm.TokenDirective, "%macro")
	// "een_macro" is not a known instruction, so it's an identifier
	requireToken(t, tokens[1], kasm.TokenIdentifier, "een_macro")
	requireToken(t, tokens[2], kasm.TokenImmediate, "2")
}

func TestLexer_OnlyNewlines(t *testing.T) {
	tokens := kasm.LexerNew("\n\n\n", x86Profile).Start()
	requireTokenCount(t, tokens, 0)
}

func TestLexer_TabsAndSpacesMixed(t *testing.T) {
	tokens := kasm.LexerNew("\t  \t mov \t rax \t", x86Profile).Start()
	requireTokenCount(t, tokens, 2)
	requireToken(t, tokens[0], kasm.TokenInstruction, "mov")
	requireToken(t, tokens[1], kasm.TokenRegister, "rax")
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLexer_EmptyInput(b *testing.B) {
	for b.Loop() {
		kasm.LexerNew("", x86Profile).Start()
	}
}

func BenchmarkLexer_SingleInstruction(b *testing.B) {
	for b.Loop() {
		kasm.LexerNew("mov rax, 1", x86Profile).Start()
	}
}

func BenchmarkLexer_Comment(b *testing.B) {
	input := "; this is a long comment with many words in it for benchmarking purposes"
	for b.Loop() {
		kasm.LexerNew(input, x86Profile).Start()
	}
}

func BenchmarkLexer_SmallProgram(b *testing.B) {
	source := `_start:
    mov rax, 60
    xor rdi, rdi
    syscall`
	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}

func BenchmarkLexer_MediumProgram(b *testing.B) {
	source := `%include "hulp.kasm"
%define DEBUG

_start:
    mov rax, 1
    mov rdi, 1
    mov rsi, message
    mov rdx, 13
    syscall

    mov rax, 60
    xor rdi, rdi
    syscall

message:
    ; "Hello, World!"
`
	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}

func BenchmarkLexer_LargeProgram(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("_start:\n")
	for i := range 500 {
		sb.WriteString(fmt.Sprintf("    mov rax, %d\n", i))
		sb.WriteString(fmt.Sprintf("    add rbx, %d\n", i*2))
		sb.WriteString("    push rax\n")
		sb.WriteString("    pop rbx\n")
	}
	sb.WriteString("    syscall\n")
	source := sb.String()

	b.ResetTimer()
	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}

func BenchmarkLexer_ManyDirectives(b *testing.B) {
	var sb strings.Builder
	for i := range 100 {
		sb.WriteString(fmt.Sprintf("%%define SYMBOL_%d\n", i))
	}
	source := sb.String()

	b.ResetTimer()
	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}

func BenchmarkLexer_ManyStrings(b *testing.B) {
	var sb strings.Builder
	for i := range 100 {
		sb.WriteString(fmt.Sprintf(`msg%d: "string literal number %d"`+"\n", i, i))
	}
	source := sb.String()

	b.ResetTimer()
	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}

func BenchmarkLexer_ManyComments(b *testing.B) {
	var sb strings.Builder
	for range 200 {
		sb.WriteString("; this is a comment line with some text\n")
	}
	source := sb.String()

	b.ResetTimer()
	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}

func BenchmarkLexer_MixedContent(b *testing.B) {
	source := `; Keurnel assembler mixed benchmark
%include "hulp.kasm"
%define OS_LINUX
%define STDOUT 1

section.text:

_start:
    ; Write syscall
    mov rax, 1          ; sys_write
    mov rdi, 1          ; stdout
    mov rsi, message    ; buffer
    mov rdx, 13         ; length
    syscall

    ; Exit syscall
    mov rax, 60         ; sys_exit
    xor rdi, rdi        ; exit code 0
    syscall

section.data:
message:
    ; "Hello, World!"
`

	for b.Loop() {
		kasm.LexerNew(source, x86Profile).Start()
	}
}
