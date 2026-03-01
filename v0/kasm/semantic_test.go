package kasm_test

import (
	"strings"
	"testing"

	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/kasm"
	"github.com/keurnel/assembler/v0/kasm/ast"
	"github.com/keurnel/assembler/v0/kasm/profile"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// minimalInstructions returns a small instruction table for isolated tests.
// Contains MOV (register,register and register,immediate) and JMP (relative, far).
func minimalInstructions() map[string]architecture.Instruction {
	return map[string]architecture.Instruction{
		"MOV": {
			Mnemonic: "MOV",
			Variants: []architecture.InstructionVariant{
				{Encoding: "RM", Operands: []string{"register", "register"}, Opcode: 0x89, Size: 2},
				{Encoding: "RI", Operands: []string{"register", "immediate"}, Opcode: 0xB8, Size: 5},
			},
		},
		"JMP": {
			Mnemonic: "JMP",
			Variants: []architecture.InstructionVariant{
				{Encoding: "R", Operands: []string{"relative"}, Opcode: 0xE9, Size: 5},
				{Encoding: "F", Operands: []string{"far"}, Opcode: 0xEA, Size: 5},
			},
		},
		"PUSH": {
			Mnemonic: "PUSH",
			Variants: []architecture.InstructionVariant{
				{Encoding: "R", Operands: []string{"register"}, Opcode: 0x50, Size: 1},
			},
		},
		"RET": {
			Mnemonic: "RET",
			Variants: []architecture.InstructionVariant{
				{Encoding: "N", Operands: []string{}, Opcode: 0xC3, Size: 1},
			},
		},
		"SYSCALL": {
			Mnemonic:    "SYSCALL",
			Description: "System call",
			// No variants defined — variant validation should be skipped.
		},
	}
}

func requireSemanticErrorCount(t *testing.T, errors []kasm.SemanticError, expected int) {
	t.Helper()
	if len(errors) != expected {
		msgs := make([]string, len(errors))
		for i, e := range errors {
			msgs[i] = e.String()
		}
		t.Fatalf("expected %d semantic error(s), got %d: [%s]", expected, len(errors), strings.Join(msgs, "; "))
	}
}

func requireNoSemanticErrors(t *testing.T, errors []kasm.SemanticError) {
	t.Helper()
	requireSemanticErrorCount(t, errors, 0)
}

func requireErrorContains(t *testing.T, errors []kasm.SemanticError, index int, substr string) {
	t.Helper()
	if index >= len(errors) {
		t.Fatalf("error index %d out of range (have %d errors)", index, len(errors))
	}
	if !strings.Contains(errors[index].Message, substr) {
		t.Errorf("expected error[%d] to contain %q, got %q", index, substr, errors[index].Message)
	}
}

// ---------------------------------------------------------------------------
// FR-1: Construction
// ---------------------------------------------------------------------------

func TestAnalyserNew_NilProgram(t *testing.T) {
	errors := kasm.AnalyserNew(nil, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyserNew_EmptyProgram(t *testing.T) {
	program := &ast.Program{Statements: []ast.Statement{}}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyserNew_NilInstructions(t *testing.T) {
	program := &ast.Program{Statements: []ast.Statement{}}
	errors := kasm.AnalyserNew(program, nil).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-3.1: Mnemonic validation
// ---------------------------------------------------------------------------

func TestAnalyse_KnownInstruction(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.RegisterOperand{Name: "rbx", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_UnknownInstruction(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "foobar",
				Operands: []ast.Operand{},
				Line:     1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "unknown instruction 'foobar'")
}

func TestAnalyse_CaseInsensitiveMnemonic(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "Mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.RegisterOperand{Name: "rbx", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-3.2: Operand count validation
// ---------------------------------------------------------------------------

func TestAnalyse_OperandCountMismatch(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.RegisterOperand{Name: "rbx", Line: 1, Column: 10},
					&ast.RegisterOperand{Name: "rcx", Line: 1, Column: 15},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "expects")
	requireErrorContains(t, errors, 0, "got 3")
}

func TestAnalyse_ZeroOperands_WhenVariantExpectsNone(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "ret",
				Operands: []ast.Operand{},
				Line:     1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_OperandCountMismatch_ZeroGiven(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{},
				Line:     1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "expects")
	requireErrorContains(t, errors, 0, "got 0")
}

// FR-3.2.2: No variants — skip count validation.
func TestAnalyse_InstructionWithoutVariants_SkipsValidation(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "syscall",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 9},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-3.3: Operand type validation
// ---------------------------------------------------------------------------

func TestAnalyse_OperandTypeMismatch(t *testing.T) {
	// mov takes (register,register) or (register,immediate), not (immediate,immediate)
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.ImmediateOperand{Value: "1", Line: 1, Column: 5},
					&ast.ImmediateOperand{Value: "2", Line: 1, Column: 8},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "no variant of 'mov' accepts operands")
}

// FR-3.3.3: Identifier compatible with relative/far.
func TestAnalyse_IdentifierAsJmpTarget(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStmt{Name: "target", Line: 1, Column: 1},
			&ast.InstructionStmt{
				Mnemonic: "jmp",
				Operands: []ast.Operand{
					&ast.IdentifierOperand{Name: "target", Line: 2, Column: 5},
				},
				Line: 2, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_ValidMovRegImm(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.ImmediateOperand{Value: "60", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-4: Label validation
// ---------------------------------------------------------------------------

func TestAnalyse_DuplicateLabel(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStmt{Name: "_start", Line: 1, Column: 1},
			&ast.LabelStmt{Name: "_start", Line: 5, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "duplicate label '_start'")
	requireErrorContains(t, errors, 0, "previously declared at 1:1")
}

func TestAnalyse_UniqueLabels(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStmt{Name: "_start", Line: 1, Column: 1},
			&ast.LabelStmt{Name: ".loop", Line: 3, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// FR-4.2: Undefined label reference.
func TestAnalyse_UndefinedReference(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "jmp",
				Operands: []ast.Operand{
					&ast.IdentifierOperand{Name: "nonexistent", Line: 1, Column: 5},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	// Expect both "undefined reference" and the variant mismatch won't fire
	// because identifier substitution to "relative" works for JMP.
	// Actually JMP expects "relative" or "far", and identifier -> relative matches,
	// so the variant check passes. Only the undefined reference error fires.
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "undefined reference to 'nonexistent'")
}

// FR-4.2.2: Forward references must resolve.
func TestAnalyse_ForwardReference(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "jmp",
				Operands: []ast.Operand{
					&ast.IdentifierOperand{Name: "later", Line: 1, Column: 5},
				},
				Line: 1, Column: 1,
			},
			&ast.LabelStmt{Name: "later", Line: 3, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-5: Namespace validation
// ---------------------------------------------------------------------------

func TestAnalyse_DuplicateNamespace(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.NamespaceStmt{Name: "myns", Line: 1, Column: 1},
			&ast.NamespaceStmt{Name: "myns", Line: 5, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "duplicate namespace 'myns'")
	requireErrorContains(t, errors, 0, "previously declared at 1:1")
}

func TestAnalyse_UniqueNamespaces(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.NamespaceStmt{Name: "ns1", Line: 1, Column: 1},
			&ast.NamespaceStmt{Name: "ns2", Line: 2, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_NamespaceStartsWithDigit(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.NamespaceStmt{Name: "9invalid", Line: 1, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "must not start with a digit")
}

// ---------------------------------------------------------------------------
// FR-6: Use statement validation
// ---------------------------------------------------------------------------

func TestAnalyse_DuplicateUse(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.UseStmt{ModuleName: "mymod", Line: 1, Column: 1},
			&ast.UseStmt{ModuleName: "mymod", Line: 3, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "duplicate use of module 'mymod'")
	requireErrorContains(t, errors, 0, "previously imported at 1:1")
}

func TestAnalyse_UniqueUses(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.UseStmt{ModuleName: "mod1", Line: 1, Column: 1},
			&ast.UseStmt{ModuleName: "mod2", Line: 2, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-7: Directive validation
// ---------------------------------------------------------------------------

func TestAnalyse_UnrecognisedDirective(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DirectiveStmt{Literal: "%foobar", Line: 1, Column: 1},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "unrecognised directive '%foobar'")
}

// ---------------------------------------------------------------------------
// FR-8: Immediate value validation
// ---------------------------------------------------------------------------

func TestAnalyse_ValidDecimalImmediate(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.ImmediateOperand{Value: "42", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_ValidHexImmediate(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.ImmediateOperand{Value: "0xFF", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_InvalidImmediate(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.ImmediateOperand{Value: "12abc", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	// 1 error for invalid immediate value.
	// Also will get a variant mismatch since "immediate" is still the type.
	// Actually no — FindVariant("register","immediate") will match RI. So only 1 error.
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "invalid immediate value '12abc'")
}

func TestAnalyse_InvalidHexImmediate_NoDigits(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 5},
					&ast.ImmediateOperand{Value: "0x", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireSemanticErrorCount(t, errors, 1)
	requireErrorContains(t, errors, 0, "invalid immediate value '0x'")
}

// ---------------------------------------------------------------------------
// FR-9: Memory operand validation
// ---------------------------------------------------------------------------

func TestAnalyse_EmptyMemoryOperand(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.MemoryOperand{Components: []ast.MemoryComponent{}, Line: 1, Column: 5},
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 10},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	// Expect "empty memory operand" + potentially a variant mismatch.
	found := false
	for _, e := range errors {
		if strings.Contains(e.Message, "empty memory operand") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'empty memory operand' error, got: %v", errors)
	}
}

func TestAnalyse_MemoryOperandImmediateBase(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.MemoryOperand{
						Components: []ast.MemoryComponent{
							{Token: kasm.Token{Type: kasm.TokenImmediate, Literal: "42", Line: 1, Column: 6}},
						},
						Line: 1, Column: 5,
					},
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 12},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	found := false
	for _, e := range errors {
		if strings.Contains(e.Message, "memory operand base must be a register or identifier") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'memory operand base' error, got: %v", errors)
	}
}

func TestAnalyse_MemoryOperandInvalidOperator(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.MemoryOperand{
						Components: []ast.MemoryComponent{
							{Token: kasm.Token{Type: kasm.TokenRegister, Literal: "rbp", Line: 1, Column: 6}},
							{Token: kasm.Token{Type: kasm.TokenIdentifier, Literal: "*", Line: 1, Column: 10}},
							{Token: kasm.Token{Type: kasm.TokenImmediate, Literal: "8", Line: 1, Column: 12}},
						},
						Line: 1, Column: 5,
					},
					&ast.RegisterOperand{Name: "rax", Line: 1, Column: 16},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	found := false
	for _, e := range errors {
		if strings.Contains(e.Message, "invalid operator '*' in memory operand") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'invalid operator' error, got: %v", errors)
	}
}

func TestAnalyse_MemoryOperandValidOperators(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "syscall", // Use syscall (no variants) to avoid variant errors.
				Operands: []ast.Operand{
					&ast.MemoryOperand{
						Components: []ast.MemoryComponent{
							{Token: kasm.Token{Type: kasm.TokenRegister, Literal: "rbp", Line: 1, Column: 6}},
							{Token: kasm.Token{Type: kasm.TokenIdentifier, Literal: "+", Line: 1, Column: 10}},
							{Token: kasm.Token{Type: kasm.TokenImmediate, Literal: "8", Line: 1, Column: 12}},
						},
						Line: 1, Column: 5,
					},
				},
				Line: 1, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// FR-2.4: Multiple errors — no early abort
// ---------------------------------------------------------------------------

func TestAnalyse_MultipleErrors(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStmt{Name: "_start", Line: 1, Column: 1},
			&ast.LabelStmt{Name: "_start", Line: 2, Column: 1}, // duplicate
			&ast.InstructionStmt{
				Mnemonic: "foobar", // unknown
				Operands: []ast.Operand{},
				Line:     3, Column: 1,
			},
			&ast.DirectiveStmt{Literal: "%bogus", Line: 4, Column: 1}, // unrecognised
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	if len(errors) < 3 {
		t.Fatalf("expected at least 3 errors, got %d: %v", len(errors), errors)
	}
}

// ---------------------------------------------------------------------------
// NFR-2.4: Forward reference resolution
// ---------------------------------------------------------------------------

func TestAnalyse_ForwardReference_FullProgram(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InstructionStmt{
				Mnemonic: "jmp",
				Operands: []ast.Operand{
					&ast.IdentifierOperand{Name: "_start", Line: 1, Column: 5},
				},
				Line: 1, Column: 1,
			},
			&ast.InstructionStmt{
				Mnemonic: "ret",
				Operands: []ast.Operand{},
				Line:     2, Column: 1,
			},
			&ast.LabelStmt{Name: "_start", Line: 3, Column: 1},
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 4, Column: 5},
					&ast.ImmediateOperand{Value: "60", Line: 4, Column: 10},
				},
				Line: 4, Column: 1,
			},
		},
	}
	errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
	requireNoSemanticErrors(t, errors)
}

// ---------------------------------------------------------------------------
// SemanticError.String()
// ---------------------------------------------------------------------------

func TestSemanticError_String(t *testing.T) {
	e := kasm.SemanticError{Message: "unknown instruction 'foo'", Line: 3, Column: 7}
	expected := "3:7: unknown instruction 'foo'"
	if e.String() != expected {
		t.Errorf("expected %q, got %q", expected, e.String())
	}
}

// ---------------------------------------------------------------------------
// Integration: lexer → parser → analyser
// ---------------------------------------------------------------------------

func TestAnalyse_Integration_FullPipeline(t *testing.T) {
	source := `_start:
    mov rax, 60
    ret`

	archProfile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, archProfile).Start()
	program, parseErrors := kasm.ParserNew(tokens).Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("unexpected parse errors: %v", parseErrors)
	}

	// Build instruction table from architecture data.
	instrTable := map[string]architecture.Instruction{
		"MOV": {
			Mnemonic: "MOV",
			Variants: []architecture.InstructionVariant{
				{Encoding: "RM", Operands: []string{"register", "register"}, Opcode: 0x89, Size: 2},
				{Encoding: "RI", Operands: []string{"register", "immediate"}, Opcode: 0xB8, Size: 5},
			},
		},
		"RET": {
			Mnemonic: "RET",
			Variants: []architecture.InstructionVariant{
				{Encoding: "N", Operands: []string{}, Opcode: 0xC3, Size: 1},
			},
		},
	}

	errors := kasm.AnalyserNew(program, instrTable).Analyse()
	requireNoSemanticErrors(t, errors)
}

func TestAnalyse_Integration_WithErrors(t *testing.T) {
	// 'nop' is a valid instruction token (known to the lexer profile), but
	// we deliberately exclude it from the instruction table, so the analyser
	// will report it as unknown. 'mov rax' has the wrong operand count.
	source := `_start:
    nop
    mov rax`

	archProfile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, archProfile).Start()
	program, parseErrors := kasm.ParserNew(tokens).Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("unexpected parse errors: %v", parseErrors)
	}

	instrTable := map[string]architecture.Instruction{
		"MOV": {
			Mnemonic: "MOV",
			Variants: []architecture.InstructionVariant{
				{Encoding: "RM", Operands: []string{"register", "register"}, Opcode: 0x89, Size: 2},
				{Encoding: "RI", Operands: []string{"register", "immediate"}, Opcode: 0xB8, Size: 5},
			},
		},
		// NOP deliberately omitted — will be reported as unknown.
	}

	errors := kasm.AnalyserNew(program, instrTable).Analyse()
	// 'nop' is unknown, 'mov rax' has wrong operand count.
	if len(errors) < 2 {
		t.Fatalf("expected at least 2 errors, got %d: %v", len(errors), errors)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkAnalyse_SmallProgram(b *testing.B) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStmt{Name: "_start", Line: 1, Column: 1},
			&ast.InstructionStmt{
				Mnemonic: "mov",
				Operands: []ast.Operand{
					&ast.RegisterOperand{Name: "rax", Line: 2, Column: 5},
					&ast.ImmediateOperand{Value: "60", Line: 2, Column: 10},
				},
				Line: 2, Column: 1,
			},
			&ast.InstructionStmt{
				Mnemonic: "ret",
				Operands: []ast.Operand{},
				Line:     3, Column: 1,
			},
		},
	}
	instrs := minimalInstructions()
	for b.Loop() {
		kasm.AnalyserNew(program, instrs).Analyse()
	}
}
