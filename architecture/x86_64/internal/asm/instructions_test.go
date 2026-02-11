package asm_test

import (
	"testing"

	"github.com/keurnel/assembler/architecture/x86_64/internal/asm"
)

// TestMOVInstruction tests the MOV instruction forms
func TestMOVInstruction(t *testing.T) {
	if asm.MOV.Mnemonic != "MOV" {
		t.Errorf("MOV mnemonic = %v, want MOV", asm.MOV.Mnemonic)
	}

	if len(asm.MOV.Forms) != 7 {
		t.Errorf("MOV has %d forms, want 7", len(asm.MOV.Forms))
	}

	// Test MOV r64, r64 form
	form := asm.MOV.Forms[3]
	if len(form.Operands) != 2 {
		t.Errorf("MOV r64, r64 has %d operands, want 2", len(form.Operands))
	}
	if form.Operands[0] != asm.OperandReg64 || form.Operands[1] != asm.OperandReg64 {
		t.Errorf("MOV r64, r64 operand types incorrect")
	}
	if len(form.Opcode) != 1 || form.Opcode[0] != 0x89 {
		t.Errorf("MOV r64, r64 opcode = %v, want [0x89]", form.Opcode)
	}
	if !form.ModRM {
		t.Error("MOV r64, r64 should require ModRM byte")
	}
	if form.REXPrefix != 0x48 {
		t.Errorf("MOV r64, r64 REXPrefix = 0x%02X, want 0x48", form.REXPrefix)
	}
}

// TestADDInstruction tests the ADD instruction forms
func TestADDInstruction(t *testing.T) {
	if asm.ADD.Mnemonic != "ADD" {
		t.Errorf("ADD mnemonic = %v, want ADD", asm.ADD.Mnemonic)
	}

	if len(asm.ADD.Forms) != 5 {
		t.Errorf("ADD has %d forms, want 5", len(asm.ADD.Forms))
	}

	// Test ADD r64, r64 form
	form := asm.ADD.Forms[2]
	if form.Operands[0] != asm.OperandReg64 || form.Operands[1] != asm.OperandReg64 {
		t.Errorf("ADD r64, r64 operand types incorrect")
	}
	if len(form.Opcode) != 1 || form.Opcode[0] != 0x01 {
		t.Errorf("ADD r64, r64 opcode = %v, want [0x01]", form.Opcode)
	}
}

// TestJumpInstructions tests conditional and unconditional jump instructions
func TestJumpInstructions(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		mnemonic  string
		wantForms int
	}{
		{"JMP", asm.JMP, "JMP", 3},
		{"JE", asm.JE, "JE", 2},
		{"JNE", asm.JNE, "JNE", 2},
		{"JG", asm.JG, "JG", 2},
		{"JGE", asm.JGE, "JGE", 2},
		{"JL", asm.JL, "JL", 2},
		{"JLE", asm.JLE, "JLE", 2},
		{"JA", asm.JA, "JA", 2},
		{"JAE", asm.JAE, "JAE", 2},
		{"JB", asm.JB, "JB", 2},
		{"JBE", asm.JBE, "JBE", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.instr.Mnemonic != tt.mnemonic {
				t.Errorf("%s mnemonic = %v, want %v", tt.name, tt.instr.Mnemonic, tt.mnemonic)
			}
			if len(tt.instr.Forms) != tt.wantForms {
				t.Errorf("%s has %d forms, want %d", tt.name, len(tt.instr.Forms), tt.wantForms)
			}
		})
	}
}

// TestArithmeticInstructions tests arithmetic instructions
func TestArithmeticInstructions(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		mnemonic  string
		wantForms int
	}{
		{"ADD", asm.ADD, "ADD", 5},
		{"SUB", asm.SUB, "SUB", 5},
		{"MUL", asm.MUL, "MUL", 3},
		{"IMUL", asm.IMUL, "IMUL", 3},
		{"DIV", asm.DIV, "DIV", 3},
		{"IDIV", asm.IDIV, "IDIV", 3},
		{"INC", asm.INC, "INC", 3},
		{"DEC", asm.DEC, "DEC", 3},
		{"NEG", asm.NEG, "NEG", 3},
		{"CMP", asm.CMP, "CMP", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.instr.Mnemonic != tt.mnemonic {
				t.Errorf("%s mnemonic = %v, want %v", tt.name, tt.instr.Mnemonic, tt.mnemonic)
			}
			if len(tt.instr.Forms) != tt.wantForms {
				t.Errorf("%s has %d forms, want %d", tt.name, len(tt.instr.Forms), tt.wantForms)
			}
		})
	}
}

// TestLogicalInstructions tests logical instructions
func TestLogicalInstructions(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		mnemonic  string
		wantForms int
	}{
		{"AND", asm.AND, "AND", 4},
		{"OR", asm.OR, "OR", 4},
		{"XOR", asm.XOR, "XOR", 4},
		{"NOT", asm.NOT, "NOT", 3},
		{"TEST", asm.TEST, "TEST", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.instr.Mnemonic != tt.mnemonic {
				t.Errorf("%s mnemonic = %v, want %v", tt.name, tt.instr.Mnemonic, tt.mnemonic)
			}
			if len(tt.instr.Forms) != tt.wantForms {
				t.Errorf("%s has %d forms, want %d", tt.name, len(tt.instr.Forms), tt.wantForms)
			}
		})
	}
}

// TestShiftRotateInstructions tests shift and rotate instructions
func TestShiftRotateInstructions(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		mnemonic  string
		wantForms int
	}{
		{"SHL", asm.SHL, "SHL", 3},
		{"SHR", asm.SHR, "SHR", 3},
		{"SAR", asm.SAR, "SAR", 3},
		{"ROL", asm.ROL, "ROL", 2},
		{"ROR", asm.ROR, "ROR", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.instr.Mnemonic != tt.mnemonic {
				t.Errorf("%s mnemonic = %v, want %v", tt.name, tt.instr.Mnemonic, tt.mnemonic)
			}
			if len(tt.instr.Forms) != tt.wantForms {
				t.Errorf("%s has %d forms, want %d", tt.name, len(tt.instr.Forms), tt.wantForms)
			}
		})
	}
}

// TestDataMovementInstructions tests data movement instructions
func TestDataMovementInstructions(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		mnemonic  string
		wantForms int
	}{
		{"MOV", asm.MOV, "MOV", 7},
		{"MOVZX", asm.MOVZX, "MOVZX", 2},
		{"MOVSX", asm.MOVSX, "MOVSX", 2},
		{"LEA", asm.LEA, "LEA", 2},
		{"PUSH", asm.PUSH, "PUSH", 3},
		{"POP", asm.POP, "POP", 1},
		{"XCHG", asm.XCHG, "XCHG", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.instr.Mnemonic != tt.mnemonic {
				t.Errorf("%s mnemonic = %v, want %v", tt.name, tt.instr.Mnemonic, tt.mnemonic)
			}
			if len(tt.instr.Forms) != tt.wantForms {
				t.Errorf("%s has %d forms, want %d", tt.name, len(tt.instr.Forms), tt.wantForms)
			}
		})
	}
}

// TestMiscInstructions tests miscellaneous instructions
func TestMiscInstructions(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		mnemonic  string
		wantForms int
	}{
		{"NOP", asm.NOP, "NOP", 1},
		{"HLT", asm.HLT, "HLT", 1},
		{"SYSCALL", asm.SYSCALL, "SYSCALL", 1},
		{"SYSRET", asm.SYSRET, "SYSRET", 1},
		{"INT", asm.INT, "INT", 1},
		{"IRET", asm.IRET, "IRET", 1},
		{"CPUID", asm.CPUID, "CPUID", 1},
		{"RDTSC", asm.RDTSC, "RDTSC", 1},
		{"CALL", asm.CALL, "CALL", 2},
		{"RET", asm.RET, "RET", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.instr.Mnemonic != tt.mnemonic {
				t.Errorf("%s mnemonic = %v, want %v", tt.name, tt.instr.Mnemonic, tt.mnemonic)
			}
			if len(tt.instr.Forms) != tt.wantForms {
				t.Errorf("%s has %d forms, want %d", tt.name, len(tt.instr.Forms), tt.wantForms)
			}
		})
	}
}

// TestInstructionOpcodes tests specific opcode values
func TestInstructionOpcodes(t *testing.T) {
	tests := []struct {
		name       string
		instr      asm.Instruction
		formIndex  int
		wantOpcode []byte
	}{
		{"NOP", asm.NOP, 0, []byte{0x90}},
		{"HLT", asm.HLT, 0, []byte{0xF4}},
		{"SYSCALL", asm.SYSCALL, 0, []byte{0x0F, 0x05}},
		{"RET", asm.RET, 0, []byte{0xC3}},
		{"JMP rel8", asm.JMP, 0, []byte{0xEB}},
		{"JMP rel32", asm.JMP, 1, []byte{0xE9}},
		{"CALL rel32", asm.CALL, 0, []byte{0xE8}},
		{"INT", asm.INT, 0, []byte{0xCD}},
		{"CPUID", asm.CPUID, 0, []byte{0x0F, 0xA2}},
		{"RDTSC", asm.RDTSC, 0, []byte{0x0F, 0x31}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.formIndex >= len(tt.instr.Forms) {
				t.Fatalf("Form index %d out of range for %s", tt.formIndex, tt.name)
			}
			form := tt.instr.Forms[tt.formIndex]
			if len(form.Opcode) != len(tt.wantOpcode) {
				t.Errorf("%s opcode length = %d, want %d", tt.name, len(form.Opcode), len(tt.wantOpcode))
			}
			for i, b := range tt.wantOpcode {
				if i < len(form.Opcode) && form.Opcode[i] != b {
					t.Errorf("%s opcode[%d] = 0x%02X, want 0x%02X", tt.name, i, form.Opcode[i], b)
				}
			}
		})
	}
}

// TestInstructionsByMnemonic tests the mnemonic lookup map
func TestInstructionsByMnemonic(t *testing.T) {
	tests := []struct {
		mnemonic    string
		shouldExist bool
		wantForms   int
	}{
		// Data Movement
		{"MOV", true, 7},
		{"MOVZX", true, 2},
		{"MOVSX", true, 2},
		{"LEA", true, 2},
		{"PUSH", true, 3},
		{"POP", true, 1},
		{"XCHG", true, 3},

		// Arithmetic
		{"ADD", true, 5},
		{"SUB", true, 5},
		{"MUL", true, 3},
		{"IMUL", true, 3},
		{"DIV", true, 3},
		{"IDIV", true, 3},
		{"INC", true, 3},
		{"DEC", true, 3},
		{"NEG", true, 3},
		{"CMP", true, 4},

		// Logical
		{"AND", true, 4},
		{"OR", true, 4},
		{"XOR", true, 4},
		{"NOT", true, 3},
		{"TEST", true, 3},

		// Shift/Rotate
		{"SHL", true, 3},
		{"SHR", true, 3},
		{"SAR", true, 3},
		{"ROL", true, 2},
		{"ROR", true, 2},

		// Control Flow
		{"JMP", true, 3},
		{"JE", true, 2},
		{"JNE", true, 2},
		{"JG", true, 2},
		{"JGE", true, 2},
		{"JL", true, 2},
		{"JLE", true, 2},
		{"JA", true, 2},
		{"JAE", true, 2},
		{"JB", true, 2},
		{"JBE", true, 2},
		{"CALL", true, 2},
		{"RET", true, 2},

		// Miscellaneous
		{"NOP", true, 1},
		{"HLT", true, 1},
		{"SYSCALL", true, 1},
		{"SYSRET", true, 1},
		{"INT", true, 1},
		{"IRET", true, 1},
		{"CPUID", true, 1},
		{"RDTSC", true, 1},

		// Invalid instructions
		{"INVALID", false, 0},
		{"FOOBAR", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.mnemonic, func(t *testing.T) {
			instr, exists := asm.InstructionsByMnemonic[tt.mnemonic]
			if exists != tt.shouldExist {
				t.Errorf("InstructionsByMnemonic[%q] exists = %v, want %v", tt.mnemonic, exists, tt.shouldExist)
			}
			if tt.shouldExist {
				if instr.Mnemonic != tt.mnemonic {
					t.Errorf("InstructionsByMnemonic[%q].Mnemonic = %v, want %v", tt.mnemonic, instr.Mnemonic, tt.mnemonic)
				}
				if len(instr.Forms) != tt.wantForms {
					t.Errorf("InstructionsByMnemonic[%q] has %d forms, want %d", tt.mnemonic, len(instr.Forms), tt.wantForms)
				}
			}
		})
	}
}

// TestInstructionFormOperands tests operand types for various instruction forms
func TestInstructionFormOperands(t *testing.T) {
	tests := []struct {
		name         string
		instr        asm.Instruction
		formIndex    int
		wantOperands []asm.OperandType
		wantModRM    bool
		wantImm      bool
		wantEncoding asm.InstructionEncoding
	}{
		{
			name:         "MOV r8, r8",
			instr:        asm.MOV,
			formIndex:    0,
			wantOperands: []asm.OperandType{asm.OperandReg8, asm.OperandReg8},
			wantModRM:    true,
			wantImm:      false,
			wantEncoding: asm.EncodingLegacy,
		},
		{
			name:         "ADD r32, imm32",
			instr:        asm.ADD,
			formIndex:    3,
			wantOperands: []asm.OperandType{asm.OperandReg32, asm.OperandImm32},
			wantModRM:    true,
			wantImm:      true,
			wantEncoding: asm.EncodingLegacy,
		},
		{
			name:         "PUSH imm8",
			instr:        asm.PUSH,
			formIndex:    1,
			wantOperands: []asm.OperandType{asm.OperandImm8},
			wantModRM:    false,
			wantImm:      true,
			wantEncoding: asm.EncodingLegacy,
		},
		{
			name:         "NOP",
			instr:        asm.NOP,
			formIndex:    0,
			wantOperands: []asm.OperandType{asm.OperandNone},
			wantModRM:    false,
			wantImm:      false,
			wantEncoding: asm.EncodingLegacy,
		},
		{
			name:         "JE rel8",
			instr:        asm.JE,
			formIndex:    0,
			wantOperands: []asm.OperandType{asm.OperandRel8},
			wantModRM:    false,
			wantImm:      true,
			wantEncoding: asm.EncodingLegacy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.formIndex >= len(tt.instr.Forms) {
				t.Fatalf("Form index %d out of range", tt.formIndex)
			}
			form := tt.instr.Forms[tt.formIndex]

			if len(form.Operands) != len(tt.wantOperands) {
				t.Errorf("Operand count = %d, want %d", len(form.Operands), len(tt.wantOperands))
			}

			for i, want := range tt.wantOperands {
				if i < len(form.Operands) && form.Operands[i] != want {
					t.Errorf("Operand[%d] = %v, want %v", i, form.Operands[i], want)
				}
			}

			if form.ModRM != tt.wantModRM {
				t.Errorf("ModRM = %v, want %v", form.ModRM, tt.wantModRM)
			}

			if form.Imm != tt.wantImm {
				t.Errorf("Imm = %v, want %v", form.Imm, tt.wantImm)
			}

			if form.Encoding != tt.wantEncoding {
				t.Errorf("Encoding = %v, want %v", form.Encoding, tt.wantEncoding)
			}
		})
	}
}

// TestREXPrefixRequirements tests REX prefix requirements for 64-bit operations
func TestREXPrefixRequirements(t *testing.T) {
	tests := []struct {
		name          string
		instr         asm.Instruction
		formIndex     int
		wantREXPrefix byte
	}{
		{"MOV r64, r64", asm.MOV, 3, 0x48},
		{"ADD r64, r64", asm.ADD, 2, 0x48},
		{"SUB r64, r64", asm.SUB, 2, 0x48},
		{"XOR r64, r64", asm.XOR, 2, 0x48},
		{"LEA r64, m", asm.LEA, 1, 0x48},
		{"CMP r64, r64", asm.CMP, 2, 0x48},
		{"MUL r64", asm.MUL, 2, 0x48},
		{"DIV r64", asm.DIV, 2, 0x48},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.formIndex >= len(tt.instr.Forms) {
				t.Fatalf("Form index %d out of range", tt.formIndex)
			}
			form := tt.instr.Forms[tt.formIndex]
			if form.REXPrefix != tt.wantREXPrefix {
				t.Errorf("REXPrefix = 0x%02X, want 0x%02X", form.REXPrefix, tt.wantREXPrefix)
			}
		})
	}
}

// TestOperandTypeConstants tests that operand type constants are unique
func TestOperandTypeConstants(t *testing.T) {
	types := []asm.OperandType{
		asm.OperandNone,
		asm.OperandReg8,
		asm.OperandReg16,
		asm.OperandReg32,
		asm.OperandReg64,
		asm.OperandImm8,
		asm.OperandImm16,
		asm.OperandImm32,
		asm.OperandImm64,
		asm.OperandMem,
		asm.OperandMem8,
		asm.OperandMem16,
		asm.OperandMem32,
		asm.OperandMem64,
		asm.OperandRel8,
		asm.OperandRel32,
		asm.OperandRegMem8,
		asm.OperandRegMem16,
		asm.OperandRegMem32,
		asm.OperandRegMem64,
	}

	seen := make(map[asm.OperandType]bool)
	for _, ot := range types {
		if seen[ot] {
			t.Errorf("Duplicate OperandType value: %v", ot)
		}
		seen[ot] = true
	}

	if len(seen) != len(types) {
		t.Errorf("Expected %d unique OperandType values, got %d", len(types), len(seen))
	}
}

// TestInstructionEncodingConstants tests encoding type constants
func TestInstructionEncodingConstants(t *testing.T) {
	encodings := []asm.InstructionEncoding{
		asm.EncodingLegacy,
		asm.EncodingVEX,
		asm.EncodingEVEX,
		asm.EncodingXOP,
	}

	seen := make(map[asm.InstructionEncoding]bool)
	for _, enc := range encodings {
		if seen[enc] {
			t.Errorf("Duplicate InstructionEncoding value: %v", enc)
		}
		seen[enc] = true
	}

	if len(seen) != len(encodings) {
		t.Errorf("Expected %d unique InstructionEncoding values, got %d", len(encodings), len(seen))
	}
}

// TestConditionalJumpOpcodes tests that conditional jumps have correct opcodes
func TestConditionalJumpOpcodes(t *testing.T) {
	tests := []struct {
		name        string
		instr       asm.Instruction
		shortOpcode byte
		longOpcode  []byte
	}{
		{"JE", asm.JE, 0x74, []byte{0x0F, 0x84}},
		{"JNE", asm.JNE, 0x75, []byte{0x0F, 0x85}},
		{"JG", asm.JG, 0x7F, []byte{0x0F, 0x8F}},
		{"JGE", asm.JGE, 0x7D, []byte{0x0F, 0x8D}},
		{"JL", asm.JL, 0x7C, []byte{0x0F, 0x8C}},
		{"JLE", asm.JLE, 0x7E, []byte{0x0F, 0x8E}},
		{"JA", asm.JA, 0x77, []byte{0x0F, 0x87}},
		{"JAE", asm.JAE, 0x73, []byte{0x0F, 0x83}},
		{"JB", asm.JB, 0x72, []byte{0x0F, 0x82}},
		{"JBE", asm.JBE, 0x76, []byte{0x0F, 0x86}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test short form (rel8)
			if len(tt.instr.Forms) < 1 {
				t.Fatal("Expected at least 1 form")
			}
			shortForm := tt.instr.Forms[0]
			if len(shortForm.Opcode) != 1 || shortForm.Opcode[0] != tt.shortOpcode {
				t.Errorf("%s short form opcode = %v, want [0x%02X]", tt.name, shortForm.Opcode, tt.shortOpcode)
			}

			// Test long form (rel32)
			if len(tt.instr.Forms) < 2 {
				t.Fatal("Expected at least 2 forms")
			}
			longForm := tt.instr.Forms[1]
			if len(longForm.Opcode) != len(tt.longOpcode) {
				t.Errorf("%s long form opcode length = %d, want %d", tt.name, len(longForm.Opcode), len(tt.longOpcode))
			}
			for i, b := range tt.longOpcode {
				if i < len(longForm.Opcode) && longForm.Opcode[i] != b {
					t.Errorf("%s long form opcode[%d] = 0x%02X, want 0x%02X", tt.name, i, longForm.Opcode[i], b)
				}
			}
		})
	}
}

// TestInstructionsByMnemonicCompleteness verifies all defined instructions are in the map
func TestInstructionsByMnemonicCompleteness(t *testing.T) {
	allInstructions := []asm.Instruction{
		// Data Movement
		asm.MOV, asm.MOVZX, asm.MOVSX, asm.LEA, asm.PUSH, asm.POP, asm.XCHG,
		// Arithmetic
		asm.ADD, asm.SUB, asm.MUL, asm.IMUL, asm.DIV, asm.IDIV,
		asm.INC, asm.DEC, asm.NEG, asm.CMP,
		// Logical
		asm.AND, asm.OR, asm.XOR, asm.NOT, asm.TEST,
		// Shift/Rotate
		asm.SHL, asm.SHR, asm.SAR, asm.ROL, asm.ROR,
		// Control Flow
		asm.JMP, asm.JE, asm.JNE, asm.JG, asm.JGE, asm.JL, asm.JLE,
		asm.JA, asm.JAE, asm.JB, asm.JBE, asm.CALL, asm.RET,
		// Miscellaneous
		asm.NOP, asm.HLT, asm.SYSCALL, asm.SYSRET, asm.INT, asm.IRET,
		asm.CPUID, asm.RDTSC,
	}

	for _, instr := range allInstructions {
		t.Run(instr.Mnemonic, func(t *testing.T) {
			found, exists := asm.InstructionsByMnemonic[instr.Mnemonic]
			if !exists {
				t.Errorf("Instruction %q not found in InstructionsByMnemonic", instr.Mnemonic)
				return
			}
			if found.Mnemonic != instr.Mnemonic {
				t.Errorf("InstructionsByMnemonic[%q].Mnemonic = %v, want %v", instr.Mnemonic, found.Mnemonic, instr.Mnemonic)
			}
			if len(found.Forms) != len(instr.Forms) {
				t.Errorf("InstructionsByMnemonic[%q] has %d forms, want %d", instr.Mnemonic, len(found.Forms), len(instr.Forms))
			}
		})
	}
}

// TestTwoByteOpcodes tests instructions with two-byte opcodes
func TestTwoByteOpcodes(t *testing.T) {
	tests := []struct {
		name      string
		instr     asm.Instruction
		formIndex int
		want      []byte
	}{
		{"MOVZX r32, r8", asm.MOVZX, 0, []byte{0x0F, 0xB6}},
		{"MOVZX r32, r16", asm.MOVZX, 1, []byte{0x0F, 0xB7}},
		{"MOVSX r32, r8", asm.MOVSX, 0, []byte{0x0F, 0xBE}},
		{"MOVSX r32, r16", asm.MOVSX, 1, []byte{0x0F, 0xBF}},
		{"IMUL r32, r32", asm.IMUL, 1, []byte{0x0F, 0xAF}},
		{"SYSCALL", asm.SYSCALL, 0, []byte{0x0F, 0x05}},
		{"SYSRET", asm.SYSRET, 0, []byte{0x0F, 0x07}},
		{"CPUID", asm.CPUID, 0, []byte{0x0F, 0xA2}},
		{"RDTSC", asm.RDTSC, 0, []byte{0x0F, 0x31}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.formIndex >= len(tt.instr.Forms) {
				t.Fatalf("Form index %d out of range", tt.formIndex)
			}
			form := tt.instr.Forms[tt.formIndex]
			if len(form.Opcode) != len(tt.want) {
				t.Errorf("Opcode length = %d, want %d", len(form.Opcode), len(tt.want))
			}
			for i, b := range tt.want {
				if i < len(form.Opcode) && form.Opcode[i] != b {
					t.Errorf("Opcode[%d] = 0x%02X, want 0x%02X", i, form.Opcode[i], b)
				}
			}
		})
	}
}
