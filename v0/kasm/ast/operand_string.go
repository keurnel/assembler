package ast

// StringOperand wraps a TokenString. The Value contains the content between
// the quotes (delimiters already stripped by the lexer).
type StringOperand struct {
	Value  string
	Line   int
	Column int
}

func (o *StringOperand) operandNode()       {}
func (o *StringOperand) OperandLine() int   { return o.Line }
func (o *StringOperand) OperandColumn() int { return o.Column }
