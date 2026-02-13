package x86_64

import "github.com/keurnel/assembler/internal/asm"

var (
	// OperandNone - represents no operand
	OperandNone asm.OperandType = asm.OperandType{
		Identifier: "none",
		Type:       "none",
		Size:       0,
	}
	// OperandReg8 - 8-bit register
	OperandReg8 asm.OperandType = asm.OperandType{
		Identifier: "reg8",
		Type:       "register",
		Size:       8,
	}
	// OperandReg16 - 16-bit register
	OperandReg16 asm.OperandType = asm.OperandType{
		Identifier: "reg16",
		Type:       "register",
		Size:       16,
	}
	// OperandReg32 - 32-bit register
	OperandReg32 asm.OperandType = asm.OperandType{
		Identifier: "reg32",
		Type:       "register",
		Size:       32,
	}
	// OperandReg64 - 64-bit register
	OperandReg64 asm.OperandType = asm.OperandType{
		Identifier: "reg64",
		Type:       "register",
		Size:       64,
	}
	// OperandImm8 - 8-bit immediate value
	OperandImm8 asm.OperandType = asm.OperandType{
		Identifier: "imm8",
		Type:       "immediate",
		Size:       8,
	}
	// OperandImm16 - 16-bit immediate value
	OperandImm16 asm.OperandType = asm.OperandType{
		Identifier: "imm16",
		Type:       "immediate",
		Size:       16,
	}
	// OperandImm32 - 32-bit immediate value
	OperandImm32 asm.OperandType = asm.OperandType{
		Identifier: "imm32",
		Type:       "immediate",
		Size:       32,
	}
	// OperandImm64 - 64-bit immediate value
	OperandImm64 asm.OperandType = asm.OperandType{
		Identifier: "imm64",
		Type:       "immediate",
		Size:       64,
	}
	// OperandMem - memory operand (size determined by ModR/M and SIB bytes)
	OperandMem asm.OperandType = asm.OperandType{
		Identifier: "mem",
		Type:       "memory",
		Size:       0, // Size determined by ModR/M and SIB bytes
	}
	// OperandMem8 - 8-bit memory operand
	OperandMem8 asm.OperandType = asm.OperandType{
		Identifier: "mem8",
		Type:       "memory",
		Size:       8,
	}
	// OperandMem16 - 16-bit memory operand
	OperandMem16 asm.OperandType = asm.OperandType{
		Identifier: "mem16",
		Type:       "memory",
		Size:       16,
	}
	// OperandMem32 - 32-bit memory operand
	OperandMem32 asm.OperandType = asm.OperandType{
		Identifier: "mem32",
		Type:       "memory",
		Size:       32,
	}
	// OperandMem64 - 64-bit memory operand
	OperandMem64 asm.OperandType = asm.OperandType{
		Identifier: "mem64",
		Type:       "memory",
		Size:       64,
	}
	// OperandRel8 - 8-bit relative offset
	OperandRel8 asm.OperandType = asm.OperandType{
		Identifier: "rel8",
		Type:       "relative",
		Size:       8,
	}
	// OperandRel32 - 32-bit relative offset
	OperandRel32 asm.OperandType = asm.OperandType{
		Identifier: "rel32",
		Type:       "relative",
		Size:       32,
	}
	// OperandRegMem8 - register or memory operand (size determined by ModR/M and SIB bytes)
	OperandRegMem8 asm.OperandType = asm.OperandType{
		Identifier: "regmem8",
		Type:       "register/memory",
		Size:       8, // Size determined by ModR/M and SIB bytes
	}
	// OperandRegMem16 - register or memory operand (size determined by ModR/M and SIB bytes)
	OperandRegMem16 asm.OperandType = asm.OperandType{
		Identifier: "regmem16",
		Type:       "register/memory",
		Size:       16, // Size determined by ModR/M and SIB bytes
	}
	// OperandRegMem32 - register or memory operand (size determined by ModR/M and SIB bytes)
	OperandRegMem32 asm.OperandType = asm.OperandType{
		Identifier: "regmem32",
		Type:       "register/memory",
		Size:       32, // Size determined by ModR/M and SIB bytes
	}
	// OperandRegMem64 - register or memory operand (size determined by ModR/M and SIB bytes)
	OperandRegMem64 asm.OperandType = asm.OperandType{
		Identifier: "regmem64",
		Type:       "register/memory",
		Size:       64, // Size determined by ModR/M and SIB bytes
	}
)

const (
	// OperandCountOne - represents instructions that take one operand
	OperandCountOne = 1
	// OperandCountTwo - represents instructions that take two operands
	OperandCountTwo = 2
	// OperandCountThree - represents instructions that take three operands
	OperandCountThree = 3
)
