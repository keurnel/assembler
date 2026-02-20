package main

import (
	"github.com/keurnel/assembler/v0/internal/architecture"
)

type Validator struct {
	architecture.InstructionValidator
}

func (v *Validator) Validate(instr *architecture.Instruction) error {

	// Print instruction details for debugging
	println("Validating instruction:", instr.Mnemonic)
	println("Opcode:", instr.Opcode)
	println("Operands:", instr.Operands)
	println("Description:", instr.Description)
	println("Flags:", instr.Flags)
	println("Number of variants:", len(instr.Variants))

	// Print variant details for debugging
	for i, variant := range instr.Variants {
		println("Variant", i)
		println("  Encoding:", variant.Encoding)
		println("  Operands:", variant.Operands)
		println("  Opcode:", variant.Opcode)
		println("  Size:", variant.Size)
	}

	return nil
}

func main() {
	validator := &Validator{}
	instruction := architecture.ArchitectureInstructionNew("MOV", 0x01, []string{"register", "immediate"}, "Move immediate value to register", []string{"ZF"}, []architecture.InstructionVariant{})

	if err := instruction.Validate(validator); err != nil {
		println("Validation failed:", err.Error())
	}
}
