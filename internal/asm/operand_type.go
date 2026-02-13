package asm

// OperandType - represents the type of an operand in an assembly instruction (e.g., register, immediate value, memory address, etc.).
type OperandType struct {
	Identifier string // Human understandable identifier for the operand type (e.g., "reg8", "imm32", "mem64")
	Type       string // Type of the operand. (e.g., "register", "immediate", "memory")
	Size       int    // Size of bits for the operand (e.g., 8, 16, 32, 64)
}

// Identifier - returns the identifier of the operand
func (ot OperandType) IdentifierOf() string {
	return ot.Identifier
}

// Type - returns the type of the operand
func (ot OperandType) TypeOf() string {
	return ot.Type
}

// IsType - checks if the operand is of a specific type
func (ot OperandType) IsType(t string) bool {
	return ot.Type == t
}

// Bits - returns the bit size of the operand
func (ot OperandType) Bits() int {
	return ot.Size
}
