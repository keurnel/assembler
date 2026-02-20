package _64

import "github.com/keurnel/assembler/v0/architecture"

var providers = []architecture.InstructionProvider{
	dataTransferProvider{},
	arithmeticianAndLogicProvider{},
	controlFlowProvider{},
}

// Instructions - returns all x86_64 instructions across all providers.
func Instructions() map[string][]architecture.Instruction {
	var result = make(map[string][]architecture.Instruction)
	for _, p := range providers {
		result[p.Group()] = p.Provide()
	}
	return result
}
