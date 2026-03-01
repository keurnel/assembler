package ast

// MemoryComponent represents a single element inside a memory operand bracket
// expression. It holds the raw token (register, immediate, identifier, or
// operator +/-).
type MemoryComponent struct {
	Token Token
}

// MemoryOperand represents a memory reference enclosed in [ and ]. The
// Components slice holds the inner tokens in order, preserving operators (+ / -).
type MemoryOperand struct {
	Components []MemoryComponent
	Line       int
	Column     int
}

func (o *MemoryOperand) operandNode()       {}
func (o *MemoryOperand) OperandLine() int   { return o.Line }
func (o *MemoryOperand) OperandColumn() int { return o.Column }
