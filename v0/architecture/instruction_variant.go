package architecture

// InstructionVariant - represents a specific encoding variant of an instruction.
type InstructionVariant struct {
	// Encoding - defines how this variant is encoded (e.g., "RM", "MR", "RI")
	Encoding string
	// Operands - the operand types for this specific variant (e.g., ["register", "register"])
	Operands []string
	// Opcode - the opcode for this specific variant (may differ per variant)
	Opcode uint8
	// Size - the size in bytes for this specific variant
	Size uint8
}

// InstructionVariantNew - creates a new instruction variant with the given properties
func InstructionVariantNew(encoding string, operands []string, opcode uint8, size uint8) *InstructionVariant {
	return &InstructionVariant{
		Encoding: encoding,
		Operands: operands,
		Opcode:   opcode,
		Size:     size,
	}
}
