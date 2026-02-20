package _64

import "github.com/keurnel/assembler/v0/architecture"

type controlFlowProvider struct {
	architecture.InstructionProvider
}

func (p controlFlowProvider) Group() string {
	return "Control Flow"
}

func (p controlFlowProvider) Provide() []architecture.Instruction {
	return []architecture.Instruction{
		{
			Mnemonic:    "JMP",
			Description: "Unconditional jump to a specified address",
			Flags:       []string{},
			Variants: []architecture.InstructionVariant{
				{Encoding: "R", Operands: []string{"relative"}, Opcode: 0xE9, Size: 5},
				{Encoding: "F", Operands: []string{"far"}, Opcode: 0xEA, Size: 5},
			},
		},
	}
}
