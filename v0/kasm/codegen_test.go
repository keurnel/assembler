package kasm_test

import (
	"testing"

	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/kasm"
)

// ---------------------------------------------------------------------------
// FR-1: Construction
// ---------------------------------------------------------------------------

func TestGeneratorNew_NilProgram(t *testing.T) {
	gen := kasm.GeneratorNew(nil, nil)
	output, errors := gen.Generate()
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errors))
	}
	if len(output) != 0 {
		t.Fatalf("expected empty output, got %d bytes", len(output))
	}
}

func TestGeneratorNew_EmptyProgram(t *testing.T) {
	program := &kasm.Program{Statements: []kasm.Statement{}}
	gen := kasm.GeneratorNew(program, nil)
	output, errors := gen.Generate()
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errors))
	}
	if len(output) != 0 {
		t.Fatalf("expected empty output, got %d bytes", len(output))
	}
}

// ---------------------------------------------------------------------------
// FR-3: Section Handling
// ---------------------------------------------------------------------------

func TestGenerate_DefaultTextSection(t *testing.T) {
	// An instruction without an explicit section should go into .text (FR-3.2).
	instrTable := map[string]architecture.Instruction{
		"NOP": {
			Mnemonic: "NOP",
			Variants: []architecture.InstructionVariant{
				{Encoding: "R", Operands: []string{}, Opcode: 0x90, Size: 1},
			},
		},
	}
	// NOP has zero operands, so we need a variant with empty operands.
	// But our classifyOperand won't match — let's use a simple MOV reg, imm.
	instrTable = movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "42", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	output, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestGenerate_ExplicitSection(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.SectionStmt{Type: ".text", Name: "code", Line: 1, Column: 1},
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 2, Column: 5},
					&kasm.ImmediateOperand{Value: "1", Line: 2, Column: 10},
				},
				Line:   2,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	output, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ---------------------------------------------------------------------------
// FR-4: Label Resolution
// ---------------------------------------------------------------------------

func TestGenerate_DuplicateLabel(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.SectionStmt{Type: ".text", Name: "code", Line: 1, Column: 1},
			&kasm.LabelStmt{Name: "start", Line: 2, Column: 1},
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 3, Column: 5},
					&kasm.ImmediateOperand{Value: "1", Line: 3, Column: 10},
				},
				Line:   3,
				Column: 1,
			},
			&kasm.LabelStmt{Name: "start", Line: 4, Column: 1}, // duplicate
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	_, errors := gen.Generate()

	if len(errors) != 1 {
		t.Fatalf("expected 1 error for duplicate label, got %d: %v", len(errors), errors)
	}
	if msg := errors[0].Message; msg == "" {
		t.Fatal("expected error message for duplicate label")
	}
}

func TestGenerate_UnresolvedLabel(t *testing.T) {
	instrTable := jmpInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.SectionStmt{Type: ".text", Name: "code", Line: 1, Column: 1},
			&kasm.InstructionStmt{
				Mnemonic: "JMP",
				Operands: []kasm.Operand{
					&kasm.IdentifierOperand{Name: "nonexistent", Line: 2, Column: 5},
				},
				Line:   2,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	_, errors := gen.Generate()

	found := false
	for _, e := range errors {
		if e.Message == "unresolved label 'nonexistent'" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'unresolved label' error, got: %v", errors)
	}
}

// ---------------------------------------------------------------------------
// FR-5: Instruction Encoding
// ---------------------------------------------------------------------------

func TestGenerate_UnknownInstruction(t *testing.T) {
	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "FAKEINSTR",
				Operands: []kasm.Operand{},
				Line:     1,
				Column:   1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, map[string]architecture.Instruction{})
	_, errors := gen.Generate()

	if len(errors) == 0 {
		t.Fatal("expected error for unknown instruction")
	}
}

func TestGenerate_NoMatchingVariant(t *testing.T) {
	instrTable := movInstrTable()

	// MOV with two immediates — no variant should match.
	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.ImmediateOperand{Value: "1", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "2", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	_, errors := gen.Generate()

	if len(errors) == 0 {
		t.Fatal("expected error for no matching variant")
	}
}

func TestGenerate_MOV_RegisterRegister(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.RegisterOperand{Name: "RBX", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	output, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output for MOV RAX, RBX")
	}

	// Output should contain REX prefix (0x48 for REX.W) + opcode (0x89) + ModR/M.
	if output[0] != 0x48 {
		t.Errorf("expected REX.W prefix 0x48, got 0x%02X", output[0])
	}
	if output[1] != 0x89 {
		t.Errorf("expected opcode 0x89, got 0x%02X", output[1])
	}
	// ModR/M: mod=11, reg=RBX(3), r/m=RAX(0) → 0xC0 | (3<<3) | 0 = 0xD8
	if output[2] != 0xD8 {
		t.Errorf("expected ModR/M 0xD8, got 0x%02X", output[2])
	}
}

func TestGenerate_MOV_RegisterImmediate(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "42", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	output, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output for MOV RAX, 42")
	}

	// Should have REX.W prefix.
	if output[0] != 0x48 {
		t.Errorf("expected REX.W prefix 0x48, got 0x%02X", output[0])
	}
	// Opcode 0xB8.
	if output[1] != 0xB8 {
		t.Errorf("expected opcode 0xB8, got 0x%02X", output[1])
	}
}

// ---------------------------------------------------------------------------
// FR-5.6: Immediate Formats
// ---------------------------------------------------------------------------

func TestGenerate_ImmediateHex(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "0xFF", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	_, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for hex immediate, got %d: %v", len(errors), errors)
	}
}

func TestGenerate_ImmediateBinary(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "0b1010", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	_, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for binary immediate, got %d: %v", len(errors), errors)
	}
}

func TestGenerate_ImmediateInvalid(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "notanumber", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	_, errors := gen.Generate()

	if len(errors) == 0 {
		t.Fatal("expected error for invalid immediate")
	}
}

// ---------------------------------------------------------------------------
// FR-6: REX Prefix
// ---------------------------------------------------------------------------

func TestGenerate_REX_ExtendedRegister(t *testing.T) {
	instrTable := movInstrTable()

	// MOV R8, RAX — R8 is an extended register, requires REX.B.
	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "R8", Line: 1, Column: 5},
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 9},
				},
				Line:   1,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	output, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}

	// REX byte should have W=1 and B=1 (R8 in r/m) → 0x49.
	if len(output) < 1 {
		t.Fatal("expected output")
	}
	if output[0] != 0x49 {
		t.Errorf("expected REX prefix 0x49 (W+B), got 0x%02X", output[0])
	}
}

// ---------------------------------------------------------------------------
// FR-7: Output Format
// ---------------------------------------------------------------------------

func TestGenerate_MultipleInstructions(t *testing.T) {
	instrTable := movInstrTable()

	program := &kasm.Program{
		Statements: []kasm.Statement{
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RAX", Line: 1, Column: 5},
					&kasm.ImmediateOperand{Value: "1", Line: 1, Column: 10},
				},
				Line:   1,
				Column: 1,
			},
			&kasm.InstructionStmt{
				Mnemonic: "MOV",
				Operands: []kasm.Operand{
					&kasm.RegisterOperand{Name: "RBX", Line: 2, Column: 5},
					&kasm.ImmediateOperand{Value: "2", Line: 2, Column: 10},
				},
				Line:   2,
				Column: 1,
			},
		},
	}

	gen := kasm.GeneratorNew(program, instrTable)
	output, errors := gen.Generate()

	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
	// Two MOV RI instructions, each with REX+opcode+reg+imm32.
	if len(output) == 0 {
		t.Fatal("expected non-empty output for two instructions")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func movInstrTable() map[string]architecture.Instruction {
	return map[string]architecture.Instruction{
		"MOV": {
			Mnemonic:    "MOV",
			Description: "Move data between registers or memory",
			Flags:       []string{},
			Variants: []architecture.InstructionVariant{
				{Encoding: "RM", Operands: []string{"register", "register"}, Opcode: 0x89, Size: 2},
				{Encoding: "RI", Operands: []string{"register", "immediate"}, Opcode: 0xB8, Size: 5},
			},
		},
	}
}

func jmpInstrTable() map[string]architecture.Instruction {
	return map[string]architecture.Instruction{
		"JMP": {
			Mnemonic:    "JMP",
			Description: "Unconditional jump",
			Flags:       []string{},
			Variants: []architecture.InstructionVariant{
				{Encoding: "R", Operands: []string{"relative"}, Opcode: 0xE9, Size: 5},
			},
		},
	}
}
