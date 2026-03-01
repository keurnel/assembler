package ast

// LabelStmt represents a label declaration. The Name field stores the label
// name without the trailing colon — stripping the colon is the parser's
// responsibility.
type LabelStmt struct {
	Name   string
	Line   int
	Column int
}

func (s *LabelStmt) statementNode()       {}
func (s *LabelStmt) StatementLine() int   { return s.Line }
func (s *LabelStmt) StatementColumn() int { return s.Column }
