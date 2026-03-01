package ast

// NamespaceStmt represents a `namespace` keyword followed by a name identifier.
type NamespaceStmt struct {
	Name   string
	Line   int
	Column int
}

func (s *NamespaceStmt) statementNode()       {}
func (s *NamespaceStmt) StatementLine() int   { return s.Line }
func (s *NamespaceStmt) StatementColumn() int { return s.Column }
