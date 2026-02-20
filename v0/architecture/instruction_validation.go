package architecture

type InstructionValidator interface {
	// Validate - checks if the given instruction is valid according to specific rules or constraints defined by the validator.
	Validate(instr *Instruction) error
}
