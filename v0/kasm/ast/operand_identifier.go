package ast

// IdentifierOperand wraps a TokenIdentifier that is not a label, comma, or
// bracket. This covers symbolic references such as label names and data symbols.
type IdentifierOperand struct {
	Name   string
	Line   int
	Column int
}

func (o *IdentifierOperand) operandNode()       {}
func (o *IdentifierOperand) OperandLine() int   { return o.Line }
func (o *IdentifierOperand) OperandColumn() int { return o.Column }
