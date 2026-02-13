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

	//
	// Logical Instructions
	//

	//
	// Shift and Rotate Instructions
	//

	//
	// Control Flow Instructions
	//

	//
	// Miscellaneous Instructions
	//
)

//// Arithmetic Instructions
//var (
//	ADD = Instruction{
//		Mnemonic: "ADD",
//		Forms: []InstructionForm{
//			// ADD r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x00}, ModRM: true, Encoding: EncodingLegacy},
//			// ADD r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x01}, ModRM: true, Encoding: EncodingLegacy},
//			// ADD r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x01}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//			// ADD r32, imm32
//			{Operands: []OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// ADD r64, imm32
//			{Operands: []OperandType{OperandReg64, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	SUB = Instruction{
//		Mnemonic: "SUB",
//		Forms: []InstructionForm{
//			// SUB r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x28}, ModRM: true, Encoding: EncodingLegacy},
//			// SUB r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x29}, ModRM: true, Encoding: EncodingLegacy},
//			// SUB r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x29}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//			// SUB r32, imm32
//			{Operands: []OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// SUB r64, imm32
//			{Operands: []OperandType{OperandReg64, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	MUL = Instruction{
//		Mnemonic: "MUL",
//		Forms: []InstructionForm{
//			// MUL r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
//			// MUL r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
//			// MUL r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	IMUL = Instruction{
//		Mnemonic: "IMUL",
//		Forms: []InstructionForm{
//			// IMUL r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
//			// IMUL r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x0F, 0xAF}, ModRM: true, Encoding: EncodingLegacy},
//			// IMUL r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x0F, 0xAF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	DIV = Instruction{
//		Mnemonic: "DIV",
//		Forms: []InstructionForm{
//			// DIV r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
//			// DIV r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
//			// DIV r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	IDIV = Instruction{
//		Mnemonic: "IDIV",
//		Forms: []InstructionForm{
//			// IDIV r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
//			// IDIV r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
//			// IDIV r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	INC = Instruction{
//		Mnemonic: "INC",
//		Forms: []InstructionForm{
//			// INC r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xFE}, ModRM: true, Encoding: EncodingLegacy},
//			// INC r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
//			// INC r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	DEC = Instruction{
//		Mnemonic: "DEC",
//		Forms: []InstructionForm{
//			// DEC r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xFE}, ModRM: true, Encoding: EncodingLegacy},
//			// DEC r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
//			// DEC r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	NEG = Instruction{
//		Mnemonic: "NEG",
//		Forms: []InstructionForm{
//			// NEG r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
//			// NEG r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
//			// NEG r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	CMP = Instruction{
//		Mnemonic: "CMP",
//		Forms: []InstructionForm{
//			// CMP r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x38}, ModRM: true, Encoding: EncodingLegacy},
//			// CMP r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x39}, ModRM: true, Encoding: EncodingLegacy},
//			// CMP r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x39}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//			// CMP r32, imm32
//			{Operands: []OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//)
//
//// Logical Instructions
//var (
//	AND = Instruction{
//		Mnemonic: "AND",
//		Forms: []InstructionForm{
//			// AND r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x20}, ModRM: true, Encoding: EncodingLegacy},
//			// AND r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x21}, ModRM: true, Encoding: EncodingLegacy},
//			// AND r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x21}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//			// AND r32, imm32
//			{Operands: []OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	OR = Instruction{
//		Mnemonic: "OR",
//		Forms: []InstructionForm{
//			// OR r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x08}, ModRM: true, Encoding: EncodingLegacy},
//			// OR r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x09}, ModRM: true, Encoding: EncodingLegacy},
//			// OR r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x09}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//			// OR r32, imm32
//			{Operands: []OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	XOR = Instruction{
//		Mnemonic: "XOR",
//		Forms: []InstructionForm{
//			// XOR r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x30}, ModRM: true, Encoding: EncodingLegacy},
//			// XOR r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x31}, ModRM: true, Encoding: EncodingLegacy},
//			// XOR r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x31}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//			// XOR r32, imm32
//			{Operands: []OperandType{OperandReg32, OperandImm32}, Opcode: []byte{0x81}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	NOT = Instruction{
//		Mnemonic: "NOT",
//		Forms: []InstructionForm{
//			// NOT r8
//			{Operands: []OperandType{OperandReg8}, Opcode: []byte{0xF6}, ModRM: true, Encoding: EncodingLegacy},
//			// NOT r32
//			{Operands: []OperandType{OperandReg32}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy},
//			// NOT r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xF7}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	TEST = Instruction{
//		Mnemonic: "TEST",
//		Forms: []InstructionForm{
//			// TEST r8, r8
//			{Operands: []OperandType{OperandReg8, OperandReg8}, Opcode: []byte{0x84}, ModRM: true, Encoding: EncodingLegacy},
//			// TEST r32, r32
//			{Operands: []OperandType{OperandReg32, OperandReg32}, Opcode: []byte{0x85}, ModRM: true, Encoding: EncodingLegacy},
//			// TEST r64, r64
//			{Operands: []OperandType{OperandReg64, OperandReg64}, Opcode: []byte{0x85}, ModRM: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//)
//
//// Shift and Rotate Instructions
//var (
//	SHL = Instruction{
//		Mnemonic: "SHL",
//		Forms: []InstructionForm{
//			// SHL r8, 1
//			{Operands: []OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xD0}, ModRM: true, Encoding: EncodingLegacy},
//			// SHL r32, imm8
//			{Operands: []OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// SHL r64, imm8
//			{Operands: []OperandType{OperandReg64, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	SHR = Instruction{
//		Mnemonic: "SHR",
//		Forms: []InstructionForm{
//			// SHR r8, 1
//			{Operands: []OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xD0}, ModRM: true, Encoding: EncodingLegacy},
//			// SHR r32, imm8
//			{Operands: []OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// SHR r64, imm8
//			{Operands: []OperandType{OperandReg64, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	SAR = Instruction{
//		Mnemonic: "SAR",
//		Forms: []InstructionForm{
//			// SAR r8, 1
//			{Operands: []OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xD0}, ModRM: true, Encoding: EncodingLegacy},
//			// SAR r32, imm8
//			{Operands: []OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// SAR r64, imm8
//			{Operands: []OperandType{OperandReg64, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy, REXPrefix: 0x48},
//		},
//	}
//
//	ROL = Instruction{
//		Mnemonic: "ROL",
//		Forms: []InstructionForm{
//			// ROL r8, imm8
//			{Operands: []OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xC0}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// ROL r32, imm8
//			{Operands: []OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	ROR = Instruction{
//		Mnemonic: "ROR",
//		Forms: []InstructionForm{
//			// ROR r8, imm8
//			{Operands: []OperandType{OperandReg8, OperandImm8}, Opcode: []byte{0xC0}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//			// ROR r32, imm8
//			{Operands: []OperandType{OperandReg32, OperandImm8}, Opcode: []byte{0xC1}, ModRM: true, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//)
//
//// Control Flow Instructions
//var (
//	JMP = Instruction{
//		Mnemonic: "JMP",
//		Forms: []InstructionForm{
//			// JMP rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0xEB}, Imm: true, Encoding: EncodingLegacy},
//			// JMP rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0xE9}, Imm: true, Encoding: EncodingLegacy},
//			// JMP r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JE = Instruction{
//		Mnemonic: "JE",
//		Forms: []InstructionForm{
//			// JE rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x74}, Imm: true, Encoding: EncodingLegacy},
//			// JE rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x84}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JNE = Instruction{
//		Mnemonic: "JNE",
//		Forms: []InstructionForm{
//			// JNE rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x75}, Imm: true, Encoding: EncodingLegacy},
//			// JNE rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x85}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JG = Instruction{
//		Mnemonic: "JG",
//		Forms: []InstructionForm{
//			// JG rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x7F}, Imm: true, Encoding: EncodingLegacy},
//			// JG rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8F}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JGE = Instruction{
//		Mnemonic: "JGE",
//		Forms: []InstructionForm{
//			// JGE rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x7D}, Imm: true, Encoding: EncodingLegacy},
//			// JGE rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8D}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JL = Instruction{
//		Mnemonic: "JL",
//		Forms: []InstructionForm{
//			// JL rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x7C}, Imm: true, Encoding: EncodingLegacy},
//			// JL rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8C}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JLE = Instruction{
//		Mnemonic: "JLE",
//		Forms: []InstructionForm{
//			// JLE rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x7E}, Imm: true, Encoding: EncodingLegacy},
//			// JLE rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x8E}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JA = Instruction{
//		Mnemonic: "JA",
//		Forms: []InstructionForm{
//			// JA rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x77}, Imm: true, Encoding: EncodingLegacy},
//			// JA rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x87}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JAE = Instruction{
//		Mnemonic: "JAE",
//		Forms: []InstructionForm{
//			// JAE rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x73}, Imm: true, Encoding: EncodingLegacy},
//			// JAE rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x83}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JB = Instruction{
//		Mnemonic: "JB",
//		Forms: []InstructionForm{
//			// JB rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x72}, Imm: true, Encoding: EncodingLegacy},
//			// JB rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x82}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	JBE = Instruction{
//		Mnemonic: "JBE",
//		Forms: []InstructionForm{
//			// JBE rel8
//			{Operands: []OperandType{OperandRel8}, Opcode: []byte{0x76}, Imm: true, Encoding: EncodingLegacy},
//			// JBE rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0x0F, 0x86}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	CALL = Instruction{
//		Mnemonic: "CALL",
//		Forms: []InstructionForm{
//			// CALL rel32
//			{Operands: []OperandType{OperandRel32}, Opcode: []byte{0xE8}, Imm: true, Encoding: EncodingLegacy},
//			// CALL r64
//			{Operands: []OperandType{OperandReg64}, Opcode: []byte{0xFF}, ModRM: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	RET = Instruction{
//		Mnemonic: "RET",
//		Forms: []InstructionForm{
//			// RET
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0xC3}, Encoding: EncodingLegacy},
//			// RET imm16
//			{Operands: []OperandType{OperandImm16}, Opcode: []byte{0xC2}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//)
//
//// Miscellaneous Instructions
//var (
//	NOP = Instruction{
//		Mnemonic: "NOP",
//		Forms: []InstructionForm{
//			// NOP
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0x90}, Encoding: EncodingLegacy},
//		},
//	}
//
//	HLT = Instruction{
//		Mnemonic: "HLT",
//		Forms: []InstructionForm{
//			// HLT
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0xF4}, Encoding: EncodingLegacy},
//		},
//	}
//
//	SYSCALL = Instruction{
//		Mnemonic: "SYSCALL",
//		Forms: []InstructionForm{
//			// SYSCALL
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0x0F, 0x05}, Encoding: EncodingLegacy},
//		},
//	}
//
//	SYSRET = Instruction{
//		Mnemonic: "SYSRET",
//		Forms: []InstructionForm{
//			// SYSRET
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0x0F, 0x07}, Encoding: EncodingLegacy},
//		},
//	}
//
//	INT = Instruction{
//		Mnemonic: "INT",
//		Forms: []InstructionForm{
//			// INT imm8
//			{Operands: []OperandType{OperandImm8}, Opcode: []byte{0xCD}, Imm: true, Encoding: EncodingLegacy},
//		},
//	}
//
//	IRET = Instruction{
//		Mnemonic: "IRET",
//		Forms: []InstructionForm{
//			// IRET
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0xCF}, Encoding: EncodingLegacy},
//		},
//	}
//
//	CPUID = Instruction{
//		Mnemonic: "CPUID",
//		Forms: []InstructionForm{
//			// CPUID
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0x0F, 0xA2}, Encoding: EncodingLegacy},
//		},
//	}
//
//	RDTSC = Instruction{
//		Mnemonic: "RDTSC",
//		Forms: []InstructionForm{
//			// RDTSC
//			{Operands: []OperandType{OperandNone}, Opcode: []byte{0x0F, 0x31}, Encoding: EncodingLegacy},
//		},
//	}
//)
//
//// InstructionsByMnemonic is a map for looking up instructions by their mnemonic
//var InstructionsByMnemonic = map[string]Instruction{
//	// Data Movement
//	"MOV":   MOV,
//	"MOVZX": MOVZX,
//	"MOVSX": MOVSX,
//	"LEA":   LEA,
//	"PUSH":  PUSH,
//	"POP":   POP,
//	"XCHG":  XCHG,
//
//	// Arithmetic
//	"ADD":  ADD,
//	"SUB":  SUB,
//	"MUL":  MUL,
//	"IMUL": IMUL,
//	"DIV":  DIV,
//	"IDIV": IDIV,
//	"INC":  INC,
//	"DEC":  DEC,
//	"NEG":  NEG,
//	"CMP":  CMP,
//
//	// Logical
//	"AND":  AND,
//	"OR":   OR,
//	"XOR":  XOR,
//	"NOT":  NOT,
//	"TEST": TEST,
//
//	// Shift/Rotate
//	"SHL": SHL,
//	"SHR": SHR,
//	"SAR": SAR,
//	"ROL": ROL,
//	"ROR": ROR,
//
//	// Control Flow
//	"JMP":  JMP,
//	"JE":   JE,
//	"JNE":  JNE,
//	"JG":   JG,
//	"JGE":  JGE,
//	"JL":   JL,
//	"JLE":  JLE,
//	"JA":   JA,
//	"JAE":  JAE,
//	"JB":   JB,
//	"JBE":  JBE,
//	"CALL": CALL,
//	"RET":  RET,
//
//	// Miscellaneous
//	"NOP":     NOP,
//	"HLT":     HLT,
//	"SYSCALL": SYSCALL,
//	"SYSRET":  SYSRET,
//	"INT":     INT,
//	"IRET":    IRET,
//	"CPUID":   CPUID,
//	"RDTSC":   RDTSC,
//}
