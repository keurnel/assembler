package architecture

// InstructionGroup - represents a group of related instructions in a CPU architecture (e.g., "Data Transfer", "Arithmetic", "Control Flow")
type InstructionGroup struct {
	// Name - is the name of the instruction group (e.g., "Data Transfer", "Arithmetic", "Control Flow")
	Name string
	// Instructions - is a map of instruction mnemonics to their corresponding ArchitectureInstruction structs, representing the instructions that belong to this group
	Instructions map[string]Instruction
}

// FromSlice - creates a new instruction group from a slice of instructions.
func FromSlice(name string, instructions []Instruction) *InstructionGroup {
	instrMap := make(map[string]Instruction)
	for _, instr := range instructions {
		instrMap[instr.Mnemonic] = instr
	}
	return &InstructionGroup{
		Name:         name,
		Instructions: instrMap,
	}
}

// Has - checks if an instruction with the given mnemonic exists in the group.
//
//	Returns `true` if it exists, otherwise `false`.
func (group *InstructionGroup) Has(mnemonic string) bool {
	_, exists := group.Instructions[mnemonic]
	return exists
}

// Get - retrieves an instruction from the group by its mnemonic. Returns `nil` if the instruction is not found.
func (group *InstructionGroup) Get(mnemonic string) *Instruction {
	if instr, exists := group.Instructions[mnemonic]; exists {
		return &instr
	}
	return nil
}

// Put - adds a new instruction to the group. If an instruction with the same mnemonic already exists,
// it will be overwritten.
func (group *InstructionGroup) Put(instruction Instruction) {
	group.Instructions[instruction.Mnemonic] = instruction
}

// Remove - removes an instruction from the group by its mnemonic. If the instruction does not exist,
// this method does nothing.
func (group *InstructionGroup) Remove(mnemonic string) {
	delete(group.Instructions, mnemonic)
}

// Count - returns the number of instructions in the group.
func (group *InstructionGroup) Count() int {
	return len(group.Instructions)
}

// List - returns a slice of all instructions in the group.
func (group *InstructionGroup) List() []Instruction {
	instructions := make([]Instruction, 0, len(group.Instructions))
	for _, instr := range group.Instructions {
		instructions = append(instructions, instr)
	}
	return instructions
}

// Mnemonics - returns a slice of all instruction mnemonics in the group.
func (group *InstructionGroup) Mnemonics() []string {
	mnemonics := make([]string, 0, len(group.Instructions))
	for mnemonic := range group.Instructions {
		mnemonics = append(mnemonics, mnemonic)
	}
	return mnemonics
}

// Merge - merges another instruction group into the current group. If there are duplicate mnemonics, the instructions
// from the other group will overwrite those in the current group.
func (group *InstructionGroup) Merge(other *InstructionGroup) {
	for mnemonic, instr := range other.Instructions {
		group.Instructions[mnemonic] = instr
	}
}
