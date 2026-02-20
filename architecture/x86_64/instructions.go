package x86_64

import "github.com/keurnel/assembler/internal/asm"

var (
	//
	// Data Movement Instructions
	//
	MOV = asm.Instruction{
		Mnemonic: "MOV",
		Forms: []asm.InstructionForm{
			// MOV r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x88}, ModRM: true, Encoding: EncodingLegacy},
			// MOV r16, r16
			{Operands: []asm.OperandType{OperandReg16, OperandReg16}, Opcode: []byte{0x89}, ModRM: true, Encoding: EncodingLegacy},
			// MOV r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x89}, ModRM: true, Encoding: EncodingLegacy},
			// MOV r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x89}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// MOV r8, imm8
			{Operands: []asm.OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xB0}, Imm: true, Encoding: EncodingLegacy},
			// MOV r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0xB8}, Imm: true, Encoding: EncodingLegacy},
			// MOV r64, imm64
			{Operands: []asm.OperandType{OperandReg64, OperandImm64}, Opcode: []byte{0xB8}, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	MOVZX = asm.Instruction{
		Mnemonic: "MOVZX",
		Forms: []asm.InstructionForm{
			// MOVZX r32, r8
			{Operands: []asm.OperandType{OperandReg32, OperandReg8}, Opcode: []byte{0x0F, 0xB6}, ModRM: true, Encoding: EncodingLegacy},
			// MOVZX r32, r16
			{Operands: []asm.OperandType{OperandReg32, OperandReg16}, Opcode: []byte{0x0F, 0xB7}, ModRM: true, Encoding: EncodingLegacy},
		},
	}

	MOVSX = asm.Instruction{
		Mnemonic: "MOVSX",
		Forms: []asm.InstructionForm{
			// MOVSX r32, r8
			{Operands: []asm.OperandType{OperandReg32, OperandReg8}, Opcode: []byte{0x0F, 0xBE}, ModRM: true, Encoding: EncodingLegacy},
			// MOVSX r32, r16
			{Operands: []asm.OperandType{OperandReg32, OperandReg16}, Opcode: []byte{0x0F, 0xBF}, ModRM: true, Encoding: EncodingLegacy},
		},
	}

	LEA = asm.Instruction{
		Mnemonic: "LEA",
		Forms: []asm.InstructionForm{
			// LEA r32, m
			{Operands: []asm.OperandType{OperandReg32, OperandMem}, Opcode: []byte{0x8D}, ModRM: true, Encoding: EncodingLegacy},
			// LEA r64, m
			{Operands: []asm.OperandType{OperandReg64, OperandMem}, Opcode: []byte{0x8D}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	PUSH = asm.Instruction{
		Mnemonic: "PUSH",
		Forms: []asm.InstructionForm{
			// PUSH r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0x50}, Encoding: EncodingLegacy},
			// PUSH imm8
			{Operands: []asm.OperandType{OperandImm8}, Opcode: []byte{0x6A}, Imm: true, Encoding: EncodingLegacy},
			// PUSH imm32
			{Operands: []asm.OperandType{OperandImm32}, Opcode: []byte{0x68}, Imm: true, Encoding: EncodingLegacy},
			// PUSH r/m64
			{Operands: []asm.OperandType{OperandMem}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	POP = asm.Instruction{
		Mnemonic: "POP",
		Forms: []asm.InstructionForm{
			// POP r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0x58}, Encoding: EncodingLegacy},
		},
	}

	ADD = asm.Instruction{
		Mnemonic: "ADD",
		Forms: []asm.InstructionForm{
			// ADD r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x00}, ModRM: true, Encoding: EncodingLegacy},
			// ADD r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x01}, ModRM: true, Encoding: EncodingLegacy},
			// ADD r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x01}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// ADD r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// ADD r64, imm32
			{Operands: []asm.OperandType{OperandReg64, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	XCHG = asm.Instruction{
		Mnemonic: "XCHG",
		Forms: []asm.InstructionForm{
			// XCHG r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x86}, ModRM: true, Encoding: EncodingLegacy},
			// XCHG r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x87}, ModRM: true, Encoding: EncodingLegacy},
			// XCHG r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x87}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	//
	// Arithmetic Instructions
	//

	SUB = asm.Instruction{
		Mnemonic: "SUB",
		Forms: []asm.InstructionForm{
			// SUB r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x28}, ModRM: true, Encoding: EncodingLegacy},
			// SUB r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x29}, ModRM: true, Encoding: EncodingLegacy},
			// SUB r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x29}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// SUB r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// SUB r64, imm32
			{Operands: []asm.OperandType{OperandReg64, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	MUL = asm.Instruction{
		Mnemonic: "MUL",
		Forms: []asm.InstructionForm{
			// MUL r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
			// MUL r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
			// MUL r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	IMUL = asm.Instruction{
		Mnemonic: "IMUL",
		Forms: []asm.InstructionForm{
			// IMUL r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
			// IMUL r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x0F, 0xAF}, ModRM: true, Encoding: EncodingLegacy},
			// IMUL r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x0F, 0xAF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	DIV = asm.Instruction{
		Mnemonic: "DIV",
		Forms: []asm.InstructionForm{
			// DIV r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
			// DIV r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
			// DIV r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	IDIV = asm.Instruction{
		Mnemonic: "IDIV",
		Forms: []asm.InstructionForm{
			// IDIV r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
			// IDIV r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
			// IDIV r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	INC = asm.Instruction{
		Mnemonic: "INC",
		Forms: []asm.InstructionForm{
			// INC r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xFE}, ModRM: true, Encoding: EncodingLegacy},
			// INC r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
			// INC r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	DEC = asm.Instruction{
		Mnemonic: "DEC",
		Forms: []asm.InstructionForm{
			// DEC r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xFE}, ModRM: true, Encoding: EncodingLegacy},
			// DEC r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
			// DEC r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	NEG = asm.Instruction{
		Mnemonic: "NEG",
		Forms: []asm.InstructionForm{
			// NEG r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
			// NEG r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
			// NEG r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	CMP = asm.Instruction{
		Mnemonic: "CMP",
		Forms: []asm.InstructionForm{
			// CMP r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x38}, ModRM: true, Encoding: EncodingLegacy},
			// CMP r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x39}, ModRM: true, Encoding: EncodingLegacy},
			// CMP r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x39}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// CMP r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
		},
	}

	//
	// Logical Instructions
	//

	AND = asm.Instruction{
		Mnemonic: "AND",
		Forms: []asm.InstructionForm{
			// AND r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x20}, ModRM: true, Encoding: EncodingLegacy},
			// AND r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x21}, ModRM: true, Encoding: EncodingLegacy},
			// AND r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x21}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// AND r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
		},
	}

	OR = asm.Instruction{
		Mnemonic: "OR",
		Forms: []asm.InstructionForm{
			// OR r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x08}, ModRM: true, Encoding: EncodingLegacy},
			// OR r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x09}, ModRM: true, Encoding: EncodingLegacy},
			// OR r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x09}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// OR r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
		},
	}

	XOR = asm.Instruction{
		Mnemonic: "XOR",
		Forms: []asm.InstructionForm{
			// XOR r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x30}, ModRM: true, Encoding: EncodingLegacy},
			// XOR r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x31}, ModRM: true, Encoding: EncodingLegacy},
			// XOR r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x31}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
			// XOR r32, imm32
			{Operands: []asm.OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
		},
	}

	NOT = asm.Instruction{
		Mnemonic: "NOT",
		Forms: []asm.InstructionForm{
			// NOT r8
			{Operands: []asm.OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
			// NOT r32
			{Operands: []asm.OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
			// NOT r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	TEST = asm.Instruction{
		Mnemonic: "TEST",
		Forms: []asm.InstructionForm{
			// TEST r8, r8
			{Operands: []asm.OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x84}, ModRM: true, Encoding: EncodingLegacy},
			// TEST r32, r32
			{Operands: []asm.OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x85}, ModRM: true, Encoding: EncodingLegacy},
			// TEST r64, r64
			{Operands: []asm.OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x85}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	//
	// Shift and Rotate Instructions
	//

	SHL = asm.Instruction{
		Mnemonic: "SHL",
		Forms: []asm.InstructionForm{
			// SHL r8, 1
			{Operands: []asm.OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xD0}, ModRM: true, Encoding: EncodingLegacy},
			// SHL r32, imm8
			{Operands: []asm.OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// SHL r64, imm8
			{Operands: []asm.OperandType{OperandReg64, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	SHR = asm.Instruction{
		Mnemonic: "SHR",
		Forms: []asm.InstructionForm{
			// SHR r8, 1
			{Operands: []asm.OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xD0}, ModRM: true, Encoding: EncodingLegacy},
			// SHR r32, imm8
			{Operands: []asm.OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// SHR r64, imm8
			{Operands: []asm.OperandType{OperandReg64, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	SAR = asm.Instruction{
		Mnemonic: "SAR",
		Forms: []asm.InstructionForm{
			// SAR r8, 1
			{Operands: []asm.OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xD0}, ModRM: true, Encoding: EncodingLegacy},
			// SAR r32, imm8
			{Operands: []asm.OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// SAR r64, imm8
			{Operands: []asm.OperandType{OperandReg64, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
		},
	}

	ROL = asm.Instruction{
		Mnemonic: "ROL",
		Forms: []asm.InstructionForm{
			// ROL r8, imm8
			{Operands: []asm.OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xC0}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// ROL r32, imm8
			{Operands: []asm.OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
		},
	}

	ROR = asm.Instruction{
		Mnemonic: "ROR",
		Forms: []asm.InstructionForm{
			// ROR r8, imm8
			{Operands: []asm.OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xC0}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
			// ROR r32, imm8
			{Operands: []asm.OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
		},
	}

	//
	// Control Flow Instructions
	//

	JMP = asm.Instruction{
		Mnemonic: "JMP",
		Forms: []asm.InstructionForm{
			// JMP rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0xEB}, Imm: true, Encoding: EncodingLegacy},
			// JMP rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0xE9}, Imm: true, Encoding: EncodingLegacy},
			// JMP r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
		},
	}

	JE = asm.Instruction{
		Mnemonic: "JE",
		Forms: []asm.InstructionForm{
			// JE rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x74}, Imm: true, Encoding: EncodingLegacy},
			// JE rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x84}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JNE = asm.Instruction{
		Mnemonic: "JNE",
		Forms: []asm.InstructionForm{
			// JNE rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x75}, Imm: true, Encoding: EncodingLegacy},
			// JNE rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x85}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JG = asm.Instruction{
		Mnemonic: "JG",
		Forms: []asm.InstructionForm{
			// JG rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x7F}, Imm: true, Encoding: EncodingLegacy},
			// JG rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8F}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JGE = asm.Instruction{
		Mnemonic: "JGE",
		Forms: []asm.InstructionForm{
			// JGE rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x7D}, Imm: true, Encoding: EncodingLegacy},
			// JGE rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8D}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JL = asm.Instruction{
		Mnemonic: "JL",
		Forms: []asm.InstructionForm{
			// JL rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x7C}, Imm: true, Encoding: EncodingLegacy},
			// JL rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8C}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JLE = asm.Instruction{
		Mnemonic: "JLE",
		Forms: []asm.InstructionForm{
			// JLE rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x7E}, Imm: true, Encoding: EncodingLegacy},
			// JLE rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8E}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JA = asm.Instruction{
		Mnemonic: "JA",
		Forms: []asm.InstructionForm{
			// JA rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x77}, Imm: true, Encoding: EncodingLegacy},
			// JA rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x87}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JAE = asm.Instruction{
		Mnemonic: "JAE",
		Forms: []asm.InstructionForm{
			// JAE rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x73}, Imm: true, Encoding: EncodingLegacy},
			// JAE rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x83}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JB = asm.Instruction{
		Mnemonic: "JB",
		Forms: []asm.InstructionForm{
			// JB rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x72}, Imm: true, Encoding: EncodingLegacy},
			// JB rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x82}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	JBE = asm.Instruction{
		Mnemonic: "JBE",
		Forms: []asm.InstructionForm{
			// JBE rel8
			{Operands: []asm.OperandType{OperandRel8}, Opcode: []byte{0x76}, Imm: true, Encoding: EncodingLegacy},
			// JBE rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x86}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	CALL = asm.Instruction{
		Mnemonic: "CALL",
		Forms: []asm.InstructionForm{
			// CALL rel32
			{Operands: []asm.OperandType{OperandRel32}, Opcode: []byte{0xE8}, Imm: true, Encoding: EncodingLegacy},
			// CALL r64
			{Operands: []asm.OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
		},
	}

	RET = asm.Instruction{
		Mnemonic: "RET",
		Forms: []asm.InstructionForm{
			// RET
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0xC3}, Encoding: EncodingLegacy},
			// RET imm16
			{Operands: []asm.OperandType{OperandImm16}, Opcode: []byte{0xC2}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	//
	// Miscellaneous Instructions
	//

	NOP = asm.Instruction{
		Mnemonic: "NOP",
		Forms: []asm.InstructionForm{
			// NOP
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0x90}, Encoding: EncodingLegacy},
		},
	}

	HLT = asm.Instruction{
		Mnemonic: "HLT",
		Forms: []asm.InstructionForm{
			// HLT
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0xF4}, Encoding: EncodingLegacy},
		},
	}

	SYSCALL = asm.Instruction{
		Mnemonic: "SYSCALL",
		Forms: []asm.InstructionForm{
			// SYSCALL
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0x0F, 0x05}, Encoding: EncodingLegacy},
		},
	}

	SYSRET = asm.Instruction{
		Mnemonic: "SYSRET",
		Forms: []asm.InstructionForm{
			// SYSRET
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0x0F, 0x07}, Encoding: EncodingLegacy},
		},
	}

	INT = asm.Instruction{
		Mnemonic: "INT",
		Forms: []asm.InstructionForm{
			// INT imm8
			{Operands: []asm.OperandType{OperandImm8}, Opcode: []byte{0xCD}, Imm: true, Encoding: EncodingLegacy},
		},
	}

	IRET = asm.Instruction{
		Mnemonic: "IRET",
		Forms: []asm.InstructionForm{
			// IRET
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0xCF}, Encoding: EncodingLegacy},
		},
	}

	CPUID = asm.Instruction{
		Mnemonic: "CPUID",
		Forms: []asm.InstructionForm{
			// CPUID
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0x0F, 0xA2}, Encoding: EncodingLegacy},
		},
	}

	RDTSC = asm.Instruction{
		Mnemonic: "RDTSC",
		Forms: []asm.InstructionForm{
			// RDTSC
			{Operands: []asm.OperandType{OperandNone}, Opcode: []byte{0x0F, 0x31}, Encoding: EncodingLegacy},
		},
	}
)
