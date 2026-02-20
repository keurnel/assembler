package x86_64_test

import (
	"testing"

	"github.com/keurnel/assembler/architecture/x86_64"
)

func TestAssembler_IsInstruction(t *testing.T) {
	scenarios := []struct {
		name        string
		instruction string
		expected    bool
	}{
		// Data Movement Instructions
		{"Valid instruction MOV", "MOV", true},
		{"Valid instruction MOVZX", "MOVZX", true},
		{"Valid instruction MOVSX", "MOVSX", true},
		{"Valid instruction LEA", "LEA", true},
		{"Valid instruction PUSH", "PUSH", true},
		{"Valid instruction POP", "POP", true},
		{"Valid instruction XCHG", "XCHG", true},

		// Arithmetic Instructions
		{"Valid instruction ADD", "ADD", true},
		{"Valid instruction SUB", "SUB", true},
		{"Valid instruction MUL", "MUL", true},
		{"Valid instruction IMUL", "IMUL", true},
		{"Valid instruction DIV", "DIV", true},
		{"Valid instruction IDIV", "IDIV", true},
		{"Valid instruction INC", "INC", true},
		{"Valid instruction DEC", "DEC", true},
		{"Valid instruction NEG", "NEG", true},
		{"Valid instruction CMP", "CMP", true},

		// Logical Instructions
		{"Valid instruction AND", "AND", true},
		{"Valid instruction OR", "OR", true},
		{"Valid instruction XOR", "XOR", true},
		{"Valid instruction NOT", "NOT", true},
		{"Valid instruction TEST", "TEST", true},

		// Shift and Rotate Instructions
		{"Valid instruction SHL", "SHL", true},
		{"Valid instruction SHR", "SHR", true},
		{"Valid instruction SAR", "SAR", true},
		{"Valid instruction ROL", "ROL", true},
		{"Valid instruction ROR", "ROR", true},

		// Control Flow Instructions
		{"Valid instruction JMP", "JMP", true},
		{"Valid instruction JE", "JE", true},
		{"Valid instruction JNE", "JNE", true},
		{"Valid instruction JG", "JG", true},
		{"Valid instruction JGE", "JGE", true},
		{"Valid instruction JL", "JL", true},
		{"Valid instruction JLE", "JLE", true},
		{"Valid instruction JA", "JA", true},
		{"Valid instruction JAE", "JAE", true},
		{"Valid instruction JB", "JB", true},
		{"Valid instruction JBE", "JBE", true},
		{"Valid instruction CALL", "CALL", true},
		{"Valid instruction RET", "RET", true},

		// Miscellaneous Instructions
		{"Valid instruction NOP", "NOP", true},
		{"Valid instruction HLT", "HLT", true},
		{"Valid instruction SYSCALL", "SYSCALL", true},
		{"Valid instruction SYSRET", "SYSRET", true},
		{"Valid instruction INT", "INT", true},
		{"Valid instruction IRET", "IRET", true},
		{"Valid instruction CPUID", "CPUID", true},
		{"Valid instruction RDTSC", "RDTSC", true},

		// Invalid Instructions
		{"Invalid instruction lowercase", "mov", false},
		{"Invalid instruction empty", "", false},
		{"Invalid instruction random", "INVALID_INSTR", false},
		{"Invalid instruction typo", "MOVA", false},
		{"Invalid instruction partial", "MO", false},
	}

	architecture := x86_64.New("")

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := architecture.IsInstruction(scenario.instruction)
			if result != scenario.expected {
				t.Errorf("Expected IsInstruction(%q) to be %v, got %v", scenario.instruction, scenario.expected, result)
			}
		})
	}
}

func TestAssembler_IsOperand(t *testing.T) {
	scenarios := []struct {
		name     string
		operand  string
		expected bool
	}{
		{"Valid register operand", "RAX", true},
		{"Invalid operand format", "invalid_operand", false},

		// Base displacement operands
		//
		{"Valid memory operand", "[RBP-8]", true},

		// Base displacement operands with positive offsets
		{"Valid memory operand with positive offset", "[RBP+8]", true},
		{"Valid memory operand with larger positive offset", "[RSP+16]", true},
		{"Valid memory operand with hex offset", "[RBX+0x10]", true},

		// Base displacement operands with negative offsets
		{"Valid memory operand with negative offset", "[RBP-16]", true},
		{"Valid memory operand with larger negative offset", "[RSP-32]", true},

		// Base displacement operands without offset (register indirect)
		{"Valid memory operand register indirect", "[RAX]", true},
		{"Valid memory operand register indirect RDI", "[RDI]", true},

		// Base + index operands
		{"Valid memory operand base plus index", "[RBX+RCX]", true},
		{"Valid memory operand base plus index with offset", "[RBX+RCX+8]", true},

		// Scaled index operands
		{"Valid memory operand with scaled index", "[RBX+RCX*4]", true},
		{"Valid memory operand with scaled index and offset", "[RBX+RCX*8+16]", true},

		// Direct memory addressing
		{"Valid direct memory address", "[0x400000]", true},

		// Invalid operands
		{"Invalid memory operand missing bracket", "RBP-8", false},
		{"Invalid memory operand extra bracket", "[[RBP-8]]", false},

		// Immediate values
		//
		{"Valid immediate operand", "123", true},
	}

	architecture := x86_64.New("")

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := architecture.IsOperand(scenario.operand)
			if result != scenario.expected {
				t.Errorf("Expected IsOperand(%q) to be %v, got %v", scenario.operand, scenario.expected, result)
			}
		})
	}
}
