package ast

// Statement is a sum type representing one top-level construct in the .kasm
// language. Every statement carries Line and Column of its first token for
// diagnostic purposes. The marker method statementNode() prevents unrelated
// types from satisfying the interface.
type Statement interface {
	statementNode()
	// StatementLine returns the 1-based line number where the statement starts.
	StatementLine() int
	// StatementColumn returns the 1-based column number where the statement starts.
	StatementColumn() int
}

// Operand is a sum type representing a single argument to an instruction.
// Operands are not recursive — there are no sub-expressions. The marker
// method operandNode() prevents unrelated types from satisfying the interface.
type Operand interface {
	operandNode()
	// OperandLine returns the 1-based line number where the operand starts.
	OperandLine() int
	// OperandColumn returns the 1-based column number where the operand starts.
	OperandColumn() int
}

// Program is the root AST node. It holds an ordered slice of Statement nodes
// representing every top-level construct in the source. Statements appear in
// the same order as the corresponding tokens in the input slice.
type Program struct {
	Statements []Statement
}
