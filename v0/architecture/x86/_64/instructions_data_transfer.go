package _64

import "github.com/keurnel/assembler/v0/architecture"

type dataTransferProvider struct {
	architecture.InstructionProvider
}

func (p dataTransferProvider) Group() string {
	return "Data Transfer"
}

func (p dataTransferProvider) Provide() []architecture.Instruction {
	return []architecture.Instruction{
		{
			Mnemonic:    "MOV",
			Description: "Move data between registers or memory",
			Flags:       []string{},
			Variants: []architecture.InstructionVariant{
				{Encoding: "RM", Operands: []string{"register", "register"}, Opcode: 0x89, Size: 2},
				{Encoding: "RI", Operands: []string{"register", "immediate"}, Opcode: 0xB8, Size: 5},
			},
		},
	}
}
