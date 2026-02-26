package kasm_test

import (
	"testing"

	"github.com/keurnel/assembler/v0/kasm"
	"github.com/keurnel/assembler/v0/kasm/profile"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func tok(tokenType kasm.TokenType, literal string, line, col int) kasm.Token {
	return kasm.Token{Type: tokenType, Literal: literal, Line: line, Column: col}
}

func requireStatementCount(t *testing.T, program *kasm.Program, expected int) {
	t.Helper()
	if len(program.Statements) != expected {
		t.Fatalf("expected %d statements, got %d", expected, len(program.Statements))
	}
}

func requireErrorCount(t *testing.T, errors []kasm.ParseError, expected int) {
	t.Helper()
	if len(errors) != expected {
		t.Fatalf("expected %d errors, got %d: %v", expected, len(errors), errors)
	}
}

func requireNoErrors(t *testing.T, errors []kasm.ParseError) {
	t.Helper()
	requireErrorCount(t, errors, 0)
}

// ---------------------------------------------------------------------------
// FR-1: Construction
// ---------------------------------------------------------------------------

func TestParserNew_EmptySlice(t *testing.T) {
	p := kasm.ParserNew([]kasm.Token{})
	if p.Position != 0 {
		t.Errorf("expected Position 0, got %d", p.Position)
	}
}

func TestParserNew_NilSlice(t *testing.T) {
	p := kasm.ParserNew(nil)
	program, errors := p.Parse()
	requireStatementCount(t, program, 0)
	requireNoErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-2: Parse — empty input
// ---------------------------------------------------------------------------

func TestParse_EmptyInput(t *testing.T) {
	program, errors := kasm.ParserNew([]kasm.Token{}).Parse()
	requireStatementCount(t, program, 0)
	requireNoErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-3.3 / FR-7: InstructionStmt
// ---------------------------------------------------------------------------

func TestParse_InstructionNoOperands(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "ret", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*kasm.InstructionStmt)
	if !ok {
		t.Fatalf("expected *InstructionStmt, got %T", program.Statements[0])
	}
	if stmt.Mnemonic != "ret" {
		t.Errorf("expected mnemonic %q, got %q", "ret", stmt.Mnemonic)
	}
	if len(stmt.Operands) != 0 {
		t.Errorf("expected 0 operands, got %d", len(stmt.Operands))
	}
	if stmt.Line != 1 || stmt.Column != 1 {
		t.Errorf("expected position 1:1, got %d:%d", stmt.Line, stmt.Column)
	}
}

func TestParse_InstructionOneRegisterOperand(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "push", 1, 1),
		tok(kasm.TokenRegister, "rax", 1, 6),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	if stmt.Mnemonic != "push" {
		t.Errorf("expected mnemonic %q, got %q", "push", stmt.Mnemonic)
	}
	if len(stmt.Operands) != 1 {
		t.Fatalf("expected 1 operand, got %d", len(stmt.Operands))
	}
	reg, ok := stmt.Operands[0].(*kasm.RegisterOperand)
	if !ok {
		t.Fatalf("expected *RegisterOperand, got %T", stmt.Operands[0])
	}
	if reg.Name != "rax" {
		t.Errorf("expected register name %q, got %q", "rax", reg.Name)
	}
}

func TestParse_InstructionTwoOperands(t *testing.T) {
	// mov rax, 1
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "mov", 1, 1),
		tok(kasm.TokenRegister, "rax", 1, 5),
		tok(kasm.TokenIdentifier, ",", 1, 8),
		tok(kasm.TokenImmediate, "1", 1, 10),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	if stmt.Mnemonic != "mov" {
		t.Errorf("expected mnemonic %q, got %q", "mov", stmt.Mnemonic)
	}
	if len(stmt.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(stmt.Operands))
	}

	if _, ok := stmt.Operands[0].(*kasm.RegisterOperand); !ok {
		t.Errorf("expected operand[0] to be *RegisterOperand, got %T", stmt.Operands[0])
	}
	imm, ok := stmt.Operands[1].(*kasm.ImmediateOperand)
	if !ok {
		t.Fatalf("expected operand[1] to be *ImmediateOperand, got %T", stmt.Operands[1])
	}
	if imm.Value != "1" {
		t.Errorf("expected immediate value %q, got %q", "1", imm.Value)
	}
}

func TestParse_InstructionStringOperand(t *testing.T) {
	// Hypothetical: some_instr "hello"
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "db", 1, 1),
		tok(kasm.TokenString, "hello", 1, 4),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	if len(stmt.Operands) != 1 {
		t.Fatalf("expected 1 operand, got %d", len(stmt.Operands))
	}
	str, ok := stmt.Operands[0].(*kasm.StringOperand)
	if !ok {
		t.Fatalf("expected *StringOperand, got %T", stmt.Operands[0])
	}
	if str.Value != "hello" {
		t.Errorf("expected string value %q, got %q", "hello", str.Value)
	}
}

func TestParse_InstructionIdentifierOperand(t *testing.T) {
	// jmp label_name
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "jmp", 1, 1),
		tok(kasm.TokenIdentifier, "label_name", 1, 5),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	if len(stmt.Operands) != 1 {
		t.Fatalf("expected 1 operand, got %d", len(stmt.Operands))
	}
	ident, ok := stmt.Operands[0].(*kasm.IdentifierOperand)
	if !ok {
		t.Fatalf("expected *IdentifierOperand, got %T", stmt.Operands[0])
	}
	if ident.Name != "label_name" {
		t.Errorf("expected identifier %q, got %q", "label_name", ident.Name)
	}
}

func TestParse_InstructionMultipleInstructions(t *testing.T) {
	// mov rax, 60
	// syscall
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "mov", 1, 1),
		tok(kasm.TokenRegister, "rax", 1, 5),
		tok(kasm.TokenIdentifier, ",", 1, 8),
		tok(kasm.TokenImmediate, "60", 1, 10),
		tok(kasm.TokenInstruction, "syscall", 2, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)

	stmt1 := program.Statements[0].(*kasm.InstructionStmt)
	if stmt1.Mnemonic != "mov" {
		t.Errorf("expected mnemonic %q, got %q", "mov", stmt1.Mnemonic)
	}
	if len(stmt1.Operands) != 2 {
		t.Errorf("expected 2 operands, got %d", len(stmt1.Operands))
	}

	stmt2 := program.Statements[1].(*kasm.InstructionStmt)
	if stmt2.Mnemonic != "syscall" {
		t.Errorf("expected mnemonic %q, got %q", "syscall", stmt2.Mnemonic)
	}
	if len(stmt2.Operands) != 0 {
		t.Errorf("expected 0 operands, got %d", len(stmt2.Operands))
	}
}

// ---------------------------------------------------------------------------
// FR-3.4.5 / FR-7.4: MemoryOperand
// ---------------------------------------------------------------------------

func TestParse_MemoryOperandSimple(t *testing.T) {
	// mov [rbp], rax
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "mov", 1, 1),
		tok(kasm.TokenIdentifier, "[", 1, 5),
		tok(kasm.TokenRegister, "rbp", 1, 6),
		tok(kasm.TokenIdentifier, "]", 1, 9),
		tok(kasm.TokenIdentifier, ",", 1, 10),
		tok(kasm.TokenRegister, "rax", 1, 12),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	if len(stmt.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(stmt.Operands))
	}

	mem, ok := stmt.Operands[0].(*kasm.MemoryOperand)
	if !ok {
		t.Fatalf("expected *MemoryOperand, got %T", stmt.Operands[0])
	}
	if len(mem.Components) != 1 {
		t.Fatalf("expected 1 memory component, got %d", len(mem.Components))
	}
	if mem.Components[0].Token.Literal != "rbp" {
		t.Errorf("expected component %q, got %q", "rbp", mem.Components[0].Token.Literal)
	}
}

func TestParse_MemoryOperandWithDisplacement(t *testing.T) {
	// mov [rax + 8], rbx
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "mov", 1, 1),
		tok(kasm.TokenIdentifier, "[", 1, 5),
		tok(kasm.TokenRegister, "rax", 1, 6),
		tok(kasm.TokenIdentifier, "+", 1, 10),
		tok(kasm.TokenImmediate, "8", 1, 12),
		tok(kasm.TokenIdentifier, "]", 1, 13),
		tok(kasm.TokenIdentifier, ",", 1, 14),
		tok(kasm.TokenRegister, "rbx", 1, 16),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	mem := stmt.Operands[0].(*kasm.MemoryOperand)
	if len(mem.Components) != 3 {
		t.Fatalf("expected 3 memory components, got %d", len(mem.Components))
	}
	if mem.Components[0].Token.Literal != "rax" {
		t.Errorf("expected component[0] %q, got %q", "rax", mem.Components[0].Token.Literal)
	}
	if mem.Components[1].Token.Literal != "+" {
		t.Errorf("expected component[1] %q, got %q", "+", mem.Components[1].Token.Literal)
	}
	if mem.Components[2].Token.Literal != "8" {
		t.Errorf("expected component[2] %q, got %q", "8", mem.Components[2].Token.Literal)
	}
}

func TestParse_MemoryOperandUnterminated(t *testing.T) {
	// mov [rax   (no closing bracket, followed by next instruction)
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "mov", 1, 1),
		tok(kasm.TokenIdentifier, "[", 1, 5),
		tok(kasm.TokenRegister, "rax", 1, 6),
		tok(kasm.TokenInstruction, "ret", 2, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	if errors[0].Message != "unterminated memory operand, expected ']'" {
		t.Errorf("unexpected error message: %q", errors[0].Message)
	}
	// The ret instruction should still be parsed.
	requireStatementCount(t, program, 2)
}

// ---------------------------------------------------------------------------
// FR-3.5 / FR-8: LabelStmt
// ---------------------------------------------------------------------------

func TestParse_Label(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, "_start:", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*kasm.LabelStmt)
	if !ok {
		t.Fatalf("expected *LabelStmt, got %T", program.Statements[0])
	}
	if stmt.Name != "_start" {
		t.Errorf("expected label name %q, got %q", "_start", stmt.Name)
	}
	if stmt.Line != 1 || stmt.Column != 1 {
		t.Errorf("expected position 1:1, got %d:%d", stmt.Line, stmt.Column)
	}
}

func TestParse_DotLabel(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, ".loop:", 2, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.LabelStmt)
	if stmt.Name != ".loop" {
		t.Errorf("expected label name %q, got %q", ".loop", stmt.Name)
	}
}

func TestParse_LabelFollowedByInstruction(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, "_start:", 1, 1),
		tok(kasm.TokenInstruction, "mov", 2, 5),
		tok(kasm.TokenRegister, "rax", 2, 9),
		tok(kasm.TokenIdentifier, ",", 2, 12),
		tok(kasm.TokenImmediate, "60", 2, 14),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)

	if _, ok := program.Statements[0].(*kasm.LabelStmt); !ok {
		t.Errorf("expected statement[0] to be *LabelStmt, got %T", program.Statements[0])
	}
	if _, ok := program.Statements[1].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected statement[1] to be *InstructionStmt, got %T", program.Statements[1])
	}
}

// ---------------------------------------------------------------------------
// FR-3.6 / FR-9: NamespaceStmt
// ---------------------------------------------------------------------------

func TestParse_Namespace(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenKeyword, "namespace", 1, 1),
		tok(kasm.TokenIdentifier, "mymodule", 1, 11),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*kasm.NamespaceStmt)
	if !ok {
		t.Fatalf("expected *NamespaceStmt, got %T", program.Statements[0])
	}
	if stmt.Name != "mymodule" {
		t.Errorf("expected namespace name %q, got %q", "mymodule", stmt.Name)
	}
	if stmt.Line != 1 || stmt.Column != 1 {
		t.Errorf("expected position 1:1, got %d:%d", stmt.Line, stmt.Column)
	}
}

func TestParse_NamespaceMissingName_EndOfInput(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenKeyword, "namespace", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_NamespaceMissingName_WrongToken(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenKeyword, "namespace", 1, 1),
		tok(kasm.TokenImmediate, "42", 1, 11),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	// Two errors: namespace expects an identifier, and then the stray 42
	// causes a second error when the main loop encounters it.
	requireErrorCount(t, errors, 2)
	requireStatementCount(t, program, 0)
}

// ---------------------------------------------------------------------------
// FR-3.7 / FR-10: UseStmt
// ---------------------------------------------------------------------------

func TestParse_Use(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "use", 1, 1),
		tok(kasm.TokenIdentifier, "mymodule", 1, 5),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*kasm.UseStmt)
	if !ok {
		t.Fatalf("expected *UseStmt, got %T", program.Statements[0])
	}
	if stmt.ModuleName != "mymodule" {
		t.Errorf("expected module name %q, got %q", "mymodule", stmt.ModuleName)
	}
	if stmt.Line != 1 || stmt.Column != 1 {
		t.Errorf("expected position 1:1, got %d:%d", stmt.Line, stmt.Column)
	}
}

func TestParse_UseMissingName_EndOfInput(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "use", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_UseMissingName_WrongToken(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "use", 1, 1),
		tok(kasm.TokenRegister, "rax", 1, 5),
	}
	// Two errors: use expects an identifier, and then the stray register
	// causes a second error when the main loop encounters it.
	_, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 2)
}

func TestParse_UseCaseInsensitive(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "USE", 1, 1),
		tok(kasm.TokenIdentifier, "mod", 1, 5),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.UseStmt)
	if stmt.ModuleName != "mod" {
		t.Errorf("expected module name %q, got %q", "mod", stmt.ModuleName)
	}
}

// ---------------------------------------------------------------------------
// FR-3.8 / FR-11: DirectiveStmt
// ---------------------------------------------------------------------------

func TestParse_DirectiveNoArgs(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenDirective, "%endif", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*kasm.DirectiveStmt)
	if !ok {
		t.Fatalf("expected *DirectiveStmt, got %T", program.Statements[0])
	}
	if stmt.Literal != "%endif" {
		t.Errorf("expected literal %q, got %q", "%endif", stmt.Literal)
	}
	if len(stmt.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(stmt.Args))
	}
}

func TestParse_DirectiveWithArgs(t *testing.T) {
	// %define DEBUG
	tokens := []kasm.Token{
		tok(kasm.TokenDirective, "%define", 1, 1),
		tok(kasm.TokenIdentifier, "DEBUG", 1, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.DirectiveStmt)
	if stmt.Literal != "%define" {
		t.Errorf("expected literal %q, got %q", "%define", stmt.Literal)
	}
	if len(stmt.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(stmt.Args))
	}
	if stmt.Args[0].Literal != "DEBUG" {
		t.Errorf("expected arg literal %q, got %q", "DEBUG", stmt.Args[0].Literal)
	}
}

func TestParse_DirectiveWithStringArg(t *testing.T) {
	// %include "file.kasm"
	tokens := []kasm.Token{
		tok(kasm.TokenDirective, "%include", 1, 1),
		tok(kasm.TokenString, "file.kasm", 1, 10),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.DirectiveStmt)
	if len(stmt.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(stmt.Args))
	}
	if stmt.Args[0].Literal != "file.kasm" {
		t.Errorf("expected arg literal %q, got %q", "file.kasm", stmt.Args[0].Literal)
	}
}

func TestParse_DirectiveStopsAtNextStatement(t *testing.T) {
	// %define FOO
	// mov rax, 1
	tokens := []kasm.Token{
		tok(kasm.TokenDirective, "%define", 1, 1),
		tok(kasm.TokenIdentifier, "FOO", 1, 9),
		tok(kasm.TokenInstruction, "mov", 2, 1),
		tok(kasm.TokenRegister, "rax", 2, 5),
		tok(kasm.TokenIdentifier, ",", 2, 8),
		tok(kasm.TokenImmediate, "1", 2, 10),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)

	if _, ok := program.Statements[0].(*kasm.DirectiveStmt); !ok {
		t.Errorf("expected statement[0] to be *DirectiveStmt, got %T", program.Statements[0])
	}
	if _, ok := program.Statements[1].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected statement[1] to be *InstructionStmt, got %T", program.Statements[1])
	}
}

func TestParse_ConsecutiveDirectives(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenDirective, "%define", 1, 1),
		tok(kasm.TokenIdentifier, "A", 1, 9),
		tok(kasm.TokenDirective, "%define", 2, 1),
		tok(kasm.TokenIdentifier, "B", 2, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)
}

// ---------------------------------------------------------------------------
// FR-3.9 / FR-12: SectionStmt
// ---------------------------------------------------------------------------

func TestParse_Section(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenSection, "section", 1, 1),
		tok(kasm.TokenIdentifier, ".data:", 1, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*kasm.SectionStmt)
	if !ok {
		t.Fatalf("expected *SectionStmt, got %T", program.Statements[0])
	}
	if stmt.Name != ".data" {
		t.Errorf("expected section name %q, got %q", ".data", stmt.Name)
	}
	if stmt.Line != 1 || stmt.Column != 1 {
		t.Errorf("expected position 1:1, got %d:%d", stmt.Line, stmt.Column)
	}
}

func TestParse_SectionTextWithColon(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenSection, "section", 1, 1),
		tok(kasm.TokenIdentifier, ".text:", 1, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.SectionStmt)
	if stmt.Name != ".text" {
		t.Errorf("expected section name %q, got %q", ".text", stmt.Name)
	}
}

func TestParse_SectionNameWithoutColon(t *testing.T) {
	// Section name without trailing ':' — stored as-is (FR-12.4).
	tokens := []kasm.Token{
		tok(kasm.TokenSection, "section", 1, 1),
		tok(kasm.TokenIdentifier, ".bss", 1, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.SectionStmt)
	if stmt.Name != ".bss" {
		t.Errorf("expected section name %q, got %q", ".bss", stmt.Name)
	}
}

func TestParse_SectionMissingName_EndOfInput(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenSection, "section", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_SectionMissingName_WrongToken(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenSection, "section", 1, 1),
		tok(kasm.TokenImmediate, "42", 1, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	// Two errors: section expects an identifier, then the stray 42
	// causes a second error when the main loop encounters it.
	requireErrorCount(t, errors, 2)
	requireStatementCount(t, program, 0)
}

func TestParse_SectionFollowedByInstruction(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenSection, "section", 1, 1),
		tok(kasm.TokenIdentifier, ".text:", 1, 9),
		tok(kasm.TokenInstruction, "mov", 2, 5),
		tok(kasm.TokenRegister, "rax", 2, 9),
		tok(kasm.TokenIdentifier, ",", 2, 12),
		tok(kasm.TokenImmediate, "1", 2, 14),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)

	if _, ok := program.Statements[0].(*kasm.SectionStmt); !ok {
		t.Errorf("expected statement[0] *SectionStmt, got %T", program.Statements[0])
	}
	if _, ok := program.Statements[1].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected statement[1] *InstructionStmt, got %T", program.Statements[1])
	}
}

func TestParse_SectionRecovery(t *testing.T) {
	// Stray register, then a section — recovery should stop at section.
	tokens := []kasm.Token{
		tok(kasm.TokenRegister, "rax", 1, 1),
		tok(kasm.TokenSection, "section", 2, 1),
		tok(kasm.TokenIdentifier, ".data:", 2, 9),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 1)

	if _, ok := program.Statements[0].(*kasm.SectionStmt); !ok {
		t.Errorf("expected *SectionStmt after recovery, got %T", program.Statements[0])
	}
}

func TestParse_Integration_Section(t *testing.T) {
	source := `section .data:
_start:
    mov rax, 60`

	x86Profile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, x86Profile).Start()
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)

	// section .data: + _start: + mov rax, 60 = 3 statements
	requireStatementCount(t, program, 3)

	sec, ok := program.Statements[0].(*kasm.SectionStmt)
	if !ok {
		t.Fatalf("expected *SectionStmt, got %T", program.Statements[0])
	}
	if sec.Name != ".data" {
		t.Errorf("expected section name %q, got %q", ".data", sec.Name)
	}

	if _, ok := program.Statements[1].(*kasm.LabelStmt); !ok {
		t.Errorf("expected *LabelStmt, got %T", program.Statements[1])
	}
	if _, ok := program.Statements[2].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected *InstructionStmt, got %T", program.Statements[2])
	}
}

// ---------------------------------------------------------------------------
// FR-5: Error handling and recovery
// ---------------------------------------------------------------------------

func TestParse_StrayRegister(t *testing.T) {
	// rax (register outside instruction)
	tokens := []kasm.Token{
		tok(kasm.TokenRegister, "rax", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_StrayImmediate(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenImmediate, "42", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_StrayString(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenString, "hello", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_StrayIdentifier(t *testing.T) {
	// identifier without ':' at top level
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, "something", 1, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 0)
}

func TestParse_StrayPunctuation(t *testing.T) {
	// single comma at top level
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, ",", 1, 1),
	}
	_, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
}

func TestParse_RecoveryAfterError(t *testing.T) {
	// Stray register, then a valid instruction.
	tokens := []kasm.Token{
		tok(kasm.TokenRegister, "rax", 1, 1),
		tok(kasm.TokenInstruction, "nop", 2, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 1)

	if _, ok := program.Statements[0].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected *InstructionStmt after recovery, got %T", program.Statements[0])
	}
}

func TestParse_MultipleErrors(t *testing.T) {
	// Stray register → error + recovery → nop is parsed → stray string →
	// error + recovery → ret is parsed. The string cannot be absorbed by
	// nop because nop has no following operand tokens before the string.
	tokens := []kasm.Token{
		tok(kasm.TokenRegister, "rax", 1, 1),    // error 1
		tok(kasm.TokenInstruction, "nop", 2, 1), // parsed (zero operands because next is string at top level)
		tok(kasm.TokenString, "oops", 3, 1),     // error 2 — string is valid operand so nop will eat it
	}
	// Actually: nop will consume the string as an operand. So we need a
	// layout where the second bad token is truly at statement level.
	// Use: stray_reg, label, stray_reg, instruction.
	tokens = []kasm.Token{
		tok(kasm.TokenRegister, "rax", 1, 1),       // error 1: stray register
		tok(kasm.TokenIdentifier, "_start:", 2, 1), // recovery stops here → parsed as label
		tok(kasm.TokenRegister, "rbx", 3, 1),       // error 2: stray register
		tok(kasm.TokenInstruction, "ret", 4, 1),    // recovery stops here → parsed as instruction
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 2)
	requireStatementCount(t, program, 2) // label + ret
}

func TestParse_RecoverySkipsToLabel(t *testing.T) {
	// Stray register, then a label
	tokens := []kasm.Token{
		tok(kasm.TokenRegister, "rax", 1, 1),
		tok(kasm.TokenIdentifier, "_start:", 2, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireErrorCount(t, errors, 1)
	requireStatementCount(t, program, 1)

	if _, ok := program.Statements[0].(*kasm.LabelStmt); !ok {
		t.Errorf("expected *LabelStmt after recovery, got %T", program.Statements[0])
	}
}

// ---------------------------------------------------------------------------
// FR-6: Statement dispatch — full program
// ---------------------------------------------------------------------------

func TestParse_FullProgram(t *testing.T) {
	// %include "file.kasm"
	// _start:
	//     mov rax, 60
	//     xor rdi, rdi
	//     syscall
	// namespace myns
	tokens := []kasm.Token{
		tok(kasm.TokenDirective, "%include", 1, 1),
		tok(kasm.TokenString, "file.kasm", 1, 10),
		tok(kasm.TokenIdentifier, "_start:", 2, 1),
		tok(kasm.TokenInstruction, "mov", 3, 5),
		tok(kasm.TokenRegister, "rax", 3, 9),
		tok(kasm.TokenIdentifier, ",", 3, 12),
		tok(kasm.TokenImmediate, "60", 3, 14),
		tok(kasm.TokenInstruction, "xor", 4, 5),
		tok(kasm.TokenRegister, "rdi", 4, 9),
		tok(kasm.TokenIdentifier, ",", 4, 12),
		tok(kasm.TokenRegister, "rdi", 4, 14),
		tok(kasm.TokenInstruction, "syscall", 5, 5),
		tok(kasm.TokenKeyword, "namespace", 6, 1),
		tok(kasm.TokenIdentifier, "myns", 6, 11),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 6) // directive + label + mov + xor + syscall + namespace

	if _, ok := program.Statements[0].(*kasm.DirectiveStmt); !ok {
		t.Errorf("expected statement[0] *DirectiveStmt, got %T", program.Statements[0])
	}
	if _, ok := program.Statements[1].(*kasm.LabelStmt); !ok {
		t.Errorf("expected statement[1] *LabelStmt, got %T", program.Statements[1])
	}
	if _, ok := program.Statements[2].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected statement[2] *InstructionStmt, got %T", program.Statements[2])
	}
	if _, ok := program.Statements[3].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected statement[3] *InstructionStmt, got %T", program.Statements[3])
	}
	if _, ok := program.Statements[4].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected statement[4] *InstructionStmt, got %T", program.Statements[4])
	}
	ns, ok := program.Statements[5].(*kasm.NamespaceStmt)
	if !ok {
		t.Errorf("expected statement[5] *NamespaceStmt, got %T", program.Statements[5])
	} else if ns.Name != "myns" {
		t.Errorf("expected namespace name %q, got %q", "myns", ns.Name)
	}
}

// ---------------------------------------------------------------------------
// FR-7.6: Use via instruction dispatch
// ---------------------------------------------------------------------------

func TestParse_UseFollowedByInstruction(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenInstruction, "use", 1, 1),
		tok(kasm.TokenIdentifier, "mymod", 1, 5),
		tok(kasm.TokenInstruction, "ret", 2, 1),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)

	if _, ok := program.Statements[0].(*kasm.UseStmt); !ok {
		t.Errorf("expected *UseStmt, got %T", program.Statements[0])
	}
	if _, ok := program.Statements[1].(*kasm.InstructionStmt); !ok {
		t.Errorf("expected *InstructionStmt, got %T", program.Statements[1])
	}
}

// ---------------------------------------------------------------------------
// NFR-2.2: Source positions
// ---------------------------------------------------------------------------

func TestParse_PositionTracking(t *testing.T) {
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, "_start:", 1, 1),
		tok(kasm.TokenInstruction, "mov", 2, 5),
		tok(kasm.TokenRegister, "rax", 2, 9),
		tok(kasm.TokenIdentifier, ",", 2, 12),
		tok(kasm.TokenImmediate, "1", 2, 14),
	}
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)

	label := program.Statements[0].(*kasm.LabelStmt)
	if label.Line != 1 || label.Column != 1 {
		t.Errorf("label position: expected 1:1, got %d:%d", label.Line, label.Column)
	}

	instr := program.Statements[1].(*kasm.InstructionStmt)
	if instr.Line != 2 || instr.Column != 5 {
		t.Errorf("instruction position: expected 2:5, got %d:%d", instr.Line, instr.Column)
	}

	reg := instr.Operands[0].(*kasm.RegisterOperand)
	if reg.Line != 2 || reg.Column != 9 {
		t.Errorf("register operand position: expected 2:9, got %d:%d", reg.Line, reg.Column)
	}

	imm := instr.Operands[1].(*kasm.ImmediateOperand)
	if imm.Line != 2 || imm.Column != 14 {
		t.Errorf("immediate operand position: expected 2:14, got %d:%d", imm.Line, imm.Column)
	}
}

// ---------------------------------------------------------------------------
// NFR-4.1: Integration — lexer → parser round-trip
// ---------------------------------------------------------------------------

func TestParse_Integration_LexerToParser(t *testing.T) {
	source := `_start:
    mov rax, 60
    xor rdi, rdi
    syscall`

	x86Profile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, x86Profile).Start()
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)

	// _start: + mov + xor + syscall = 4 statements
	requireStatementCount(t, program, 4)

	if _, ok := program.Statements[0].(*kasm.LabelStmt); !ok {
		t.Errorf("expected *LabelStmt, got %T", program.Statements[0])
	}
	for i := 1; i <= 3; i++ {
		if _, ok := program.Statements[i].(*kasm.InstructionStmt); !ok {
			t.Errorf("expected statement[%d] *InstructionStmt, got %T", i, program.Statements[i])
		}
	}
}

func TestParse_Integration_UseAndNamespace(t *testing.T) {
	source := `use mymodule
namespace voorbeeld`

	x86Profile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, x86Profile).Start()
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 2)

	useStmt, ok := program.Statements[0].(*kasm.UseStmt)
	if !ok {
		t.Fatalf("expected *UseStmt, got %T", program.Statements[0])
	}
	if useStmt.ModuleName != "mymodule" {
		t.Errorf("expected module %q, got %q", "mymodule", useStmt.ModuleName)
	}

	nsStmt, ok := program.Statements[1].(*kasm.NamespaceStmt)
	if !ok {
		t.Fatalf("expected *NamespaceStmt, got %T", program.Statements[1])
	}
	if nsStmt.Name != "voorbeeld" {
		t.Errorf("expected namespace %q, got %q", "voorbeeld", nsStmt.Name)
	}
}

func TestParse_Integration_MemoryOperand(t *testing.T) {
	source := `mov [rbp], rax`

	x86Profile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, x86Profile).Start()
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)
	requireStatementCount(t, program, 1)

	stmt := program.Statements[0].(*kasm.InstructionStmt)
	if len(stmt.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(stmt.Operands))
	}
	if _, ok := stmt.Operands[0].(*kasm.MemoryOperand); !ok {
		t.Errorf("expected operand[0] *MemoryOperand, got %T", stmt.Operands[0])
	}
	if _, ok := stmt.Operands[1].(*kasm.RegisterOperand); !ok {
		t.Errorf("expected operand[1] *RegisterOperand, got %T", stmt.Operands[1])
	}
}

func TestParse_Integration_DirectiveWithString(t *testing.T) {
	source := `%include "hulp.kasm"
%define DEBUG
_start:
    mov rax, 1`

	x86Profile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, x86Profile).Start()
	program, errors := kasm.ParserNew(tokens).Parse()
	requireNoErrors(t, errors)

	// %include "hulp.kasm" + %define DEBUG + _start: + mov rax, 1
	requireStatementCount(t, program, 4)

	dir := program.Statements[0].(*kasm.DirectiveStmt)
	if dir.Literal != "%include" {
		t.Errorf("expected directive %q, got %q", "%include", dir.Literal)
	}
	if len(dir.Args) != 1 || dir.Args[0].Literal != "hulp.kasm" {
		t.Errorf("expected 1 string arg 'hulp.kasm', got %v", dir.Args)
	}
}

// ---------------------------------------------------------------------------
// ParseError.String()
// ---------------------------------------------------------------------------

func TestParseError_String(t *testing.T) {
	e := kasm.ParseError{Message: "unexpected token", Line: 3, Column: 7}
	expected := "3:7: unexpected token"
	if e.String() != expected {
		t.Errorf("expected %q, got %q", expected, e.String())
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkParse_SmallProgram(b *testing.B) {
	tokens := []kasm.Token{
		tok(kasm.TokenIdentifier, "_start:", 1, 1),
		tok(kasm.TokenInstruction, "mov", 2, 5),
		tok(kasm.TokenRegister, "rax", 2, 9),
		tok(kasm.TokenIdentifier, ",", 2, 12),
		tok(kasm.TokenImmediate, "60", 2, 14),
		tok(kasm.TokenInstruction, "xor", 3, 5),
		tok(kasm.TokenRegister, "rdi", 3, 9),
		tok(kasm.TokenIdentifier, ",", 3, 12),
		tok(kasm.TokenRegister, "rdi", 3, 14),
		tok(kasm.TokenInstruction, "syscall", 4, 5),
	}
	for b.Loop() {
		kasm.ParserNew(tokens).Parse()
	}
}

func BenchmarkParse_LargeProgram(b *testing.B) {
	tokens := make([]kasm.Token, 0, 2500)
	tokens = append(tokens, tok(kasm.TokenIdentifier, "_start:", 1, 1))
	for i := range 500 {
		line := i*2 + 2
		tokens = append(tokens,
			tok(kasm.TokenInstruction, "mov", line, 5),
			tok(kasm.TokenRegister, "rax", line, 9),
			tok(kasm.TokenIdentifier, ",", line, 12),
			tok(kasm.TokenImmediate, "1", line, 14),
		)
	}
	tokens = append(tokens, tok(kasm.TokenInstruction, "syscall", 1002, 5))

	b.ResetTimer()
	for b.Loop() {
		kasm.ParserNew(tokens).Parse()
	}
}
