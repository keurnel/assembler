package asm

// Architecture - defines the interface for an assembly architecture implementation. It provides
// us with a way to interact with different assembly architectures in a consistent manner.
type Architecture interface {
	// ArchitectureName - returns the name of the architecture (e.g., "x86_64", "arm64", etc.).
	ArchitectureName() string
	// Directives - returns a list of supported directives for the architecture.
	Directives() []string
	// IsDirective - checks if a given line of assembly code is a valid directive for the architecture.
	IsDirective(line string) bool
	// Instructions - returns a list of supported instructions for the architecture.
	Instructions() map[string]Instruction
	// IsInstruction - checks if a given line of assembly code is a valid instruction for the architecture.
	IsInstruction(line string) bool
	// RegisterSet - returns a list of supported registers for the architecture.
	RegisterSet() []string
	// IsRegister - checks if a given string is a valid register for the architecture.
	IsRegister(name string) bool
	// OperandTypes - returns a list of supported operand types for the architecture.
	OperandTypes() []OperandType
	// OperandCounts - returns a list of valid operand counts for the architecture.
	OperandCounts() []int
	// IsValidOperandCount - checks if a given operand count is valid for the architecture.
	IsValidOperandCount(count int) bool
	// SourceOperandSupportsDestination - checks if a given source operand type can be used with
	// a given destination operand type in an instruction.
	SourceOperandSupportsDestination(sourceType, destType OperandType) bool
	// Is8BitInstruction - checks if a given instruction is an 8-bit instruction based on its operand types.
	Is8BitInstruction(instr Instruction) bool
}
