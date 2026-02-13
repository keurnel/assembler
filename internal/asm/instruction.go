package asm

// Instruction - represents a singular assembly instruction, including its mnemonic and operands.
type Instruction struct {
	Mnemonic           string                       // Instruction mnemonic (e.g., "MOV", "ADD")
	Forms              []InstructionForm            // Different forms of the instruction
	FormsByOperandType map[string][]InstructionForm // Cached forms by operand type identifier
}

// formsByOperandType - helper function to find instruction forms by operand type identifier
func (instr *Instruction) formsByOperandType(operandType OperandType) []InstructionForm {
	var matchedForms []InstructionForm
	for _, form := range instr.Forms {
		for _, operand := range form.Operands {
			if operand.Identifier == operandType.Identifier {
				matchedForms = append(matchedForms, form)
				break
			}
		}
	}
	return matchedForms
}

// getCachedFormsByOperandType - retrieves cached instruction forms by operand type if available. When
// not available, it returns (nil, false) indicating the cache miss.
func (instr *Instruction) getCachedFormsByOperandType(operandType OperandType) ([]InstructionForm, bool) {
	forms, exists := instr.FormsByOperandType[operandType.Identifier]
	if !exists {
		return nil, false
	}
	return forms, true
}

// cacheFormsByOperandType - caches instruction forms by operand type for faster future retrievals.
func (instr *Instruction) cacheFormsByOperandType(operandType OperandType, forms []InstructionForm) {
	if instr.FormsByOperandType == nil {
		instr.FormsByOperandType = make(map[string][]InstructionForm)
	}
	instr.FormsByOperandType[operandType.Identifier] = forms
}

// Form - retrieves the appropriate instruction form based on the provided operand type.
// When no matching form is found an empty slice is returned.
func (instr *Instruction) Form(operandType OperandType) []InstructionForm {
	// Try to get cached forms first.
	//
	cachedForms, cacheExits := instr.getCachedFormsByOperandType(operandType)
	if cacheExits {
		return cachedForms
	}

	// Find matching forms based on operand type.
	//
	matchedForms := instr.formsByOperandType(operandType)
	instr.cacheFormsByOperandType(operandType, matchedForms)

	return matchedForms
}
