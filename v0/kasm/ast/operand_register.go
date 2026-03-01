package ast

// RegisterOperand wraps a TokenRegister. The Name preserves original casing.
type RegisterOperand struct {
	Name   string
	Line   int
	Column int
}

func (o *RegisterOperand) operandNode()       {}
func (o *RegisterOperand) OperandLine() int   { return o.Line }
func (o *RegisterOperand) OperandColumn() int { return o.Column }
