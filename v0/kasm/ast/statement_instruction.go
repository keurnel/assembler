package ast

// InstructionStmt represents an instruction mnemonic followed by zero or more
// operands. The mnemonic is stored as a string — the parser does not restrict
// which instructions are valid (that is a semantic concern).
type InstructionStmt struct {
	Mnemonic string
	Operands []Operand
	Line     int
	Column   int
}

func (s *InstructionStmt) statementNode()       {}
func (s *InstructionStmt) StatementLine() int   { return s.Line }
func (s *InstructionStmt) StatementColumn() int { return s.Column }
