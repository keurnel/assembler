package asm

// InstructionForm represents a specific form/variant of an instruction
type InstructionForm struct {
	Operands  []OperandType       // Operand types
	Opcode    []byte              // Opcode bytes
	ModRM     bool                // Whether ModR/M byte is required
	Imm       bool                // Whether immediate value follows
	Encoding  InstructionEncoding // Encoding type
	REXPrefix byte                // REX prefix requirements (0 if none)
}
