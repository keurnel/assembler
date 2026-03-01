package ast

// ImmediateOperand wraps a TokenImmediate. The Value is stored verbatim —
// numeric conversion is deferred to semantic analysis or code generation.
type ImmediateOperand struct {
	Value  string
	Line   int
	Column int
}

func (o *ImmediateOperand) operandNode()       {}
func (o *ImmediateOperand) OperandLine() int   { return o.Line }
func (o *ImmediateOperand) OperandColumn() int { return o.Column }
