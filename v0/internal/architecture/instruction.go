package architecture

// Instruction - represents a single instruction in a CPU architecture.
type Instruction struct {
	// Mnemonic - is the human-readable name of the instruction (e.g., "MOV", "ADD", "JMP")
	Mnemonic string
	// Opcode - is the binary representation of the instruction
	Opcode uint8
	// Operands - defines the types/kinds of operands the instruction accepts (e.g., "register", "immediate", "memory")
	Operands []string
	// Description - is a human-readable description of what the instruction does
	Description string
	// Flags - indicates which CPU flags are affected by this instruction (e.g., "ZF", "CF", "OF")
	Flags []string
	// Variants - the different encoding forms this instruction can take
	Variants []InstructionVariant
}

// ArchitectureInstructionNew - creates a new instruction with the given properties
func ArchitectureInstructionNew(mnemonic string, opcode uint8, operands []string, description string, flags []string, variants []InstructionVariant) *Instruction {
	return &Instruction{
		Mnemonic:    mnemonic,
		Opcode:      opcode,
		Operands:    operands,
		Description: description,
		Flags:       flags,
		Variants:    variants,
	}
}

// Validate - performs validation checks against the instruction's properties and variants to ensure they conform
// to expected formats and constraints. Returns an error if any validation fails.
func (instr *Instruction) Validate(validators ...InstructionValidator) error {
	for _, validator := range validators {
		if err := validator.Validate(instr); err != nil {
			return err
		}
	}
	return nil
}

// HasVariants - returns whether the instruction has multiple encoding variants.
func (instr *Instruction) HasVariants() bool {
	return len(instr.Variants) > 0
}

// FindVariant - returns the first variant matching the given operand types, or nil if not found.
func (instr *Instruction) FindVariant(operands ...string) *InstructionVariant {
	for i := range instr.Variants {
		v := &instr.Variants[i]
		if len(v.Operands) != len(operands) {
			continue
		}
		match := true
		for j, op := range operands {
			if v.Operands[j] != op {
				match = false
				break
			}
		}
		if match {
			return v
		}
	}
	return nil
}
