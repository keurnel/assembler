package x86_64

import (
	"github.com/keurnel/assembler/internal/asm"
)

type Assembler struct {
	asm.Architecture
	rawSource string
}

// AssemblerNew - returns a new instance of the x86_64 assembler
func AssemblerNew(rawSource string) *Assembler {
	return &Assembler{
		rawSource: rawSource,
	}
}

// ArchitectureName - returns the name of the architecture
func (a *Assembler) ArchitectureName() string {
	return "x86_64"
}

// Instructions - returns the parsed assembly instructions
func (a *Assembler) Instructions() map[string]asm.Instruction {
	return map[string]asm.Instruction{}
}

// IsInstruction - checks if a given line of assembly code is a valid instruction for the architecture
func (a *Assembler) IsInstruction(line string) bool {
	return false
}

// RegisterSet - returns a list of supported registers for the architecture
func (a *Assembler) RegisterSet() []string {
	return []string{}
}

// IsRegister - checks if a given string is a valid register for the architecture
func (a *Assembler) IsRegister(name string) bool {
	return false
}

// OperandTypes - returns a list of supported operand types for the architecture
func (a *Assembler) OperandTypes() []asm.OperandType {
	return []asm.OperandType{
		OperandNone,
		OperandReg8,
		OperandReg16,
		OperandReg32,
		OperandReg64,
		OperandImm8,
		OperandImm16,
		OperandImm32,
		OperandImm64,
		OperandMem,
		OperandMem8,
		OperandMem16,
		OperandMem32,
		OperandMem64,
		OperandRel8,
		OperandRel32,
		OperandRegMem8,
		OperandRegMem16,
		OperandRegMem32,
		OperandRegMem64,
	}
}

// OperandCounts - returns a list of valid operand counts for the architecture
func (a *Assembler) OperandCounts() []int {
	return []int{OperandCountOne, OperandCountTwo, OperandCountThree}
}

// IsValidOperandCount - checks if a given operand count is valid for the architecture
func (a *Assembler) IsValidOperandCount(count int) bool {
	return count >= OperandCountOne && count <= OperandCountThree
}

// SourceOperandSupportsDestination - checks if a given source operand type can be used with a given destination operand type in an instruction
func (a *Assembler) SourceOperandSupportsDestination(sourceType, destType asm.OperandType) bool {
	// todo: implement this function based on the rules of operand compatibility for x86_64 instructions
	return false
}

// Is8BitInstruction - checks if a given instruction is an 8-bit instruction based on its operand types
func (a *Assembler) Is8BitInstruction(instr asm.Instruction) bool {
	// todo: implement this function based on the instruction's operand types
	return false
}

// RawSource - returns the raw assembly source code
func (a *Assembler) RawSource() string {
	return a.rawSource
}
