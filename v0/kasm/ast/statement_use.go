package ast

// UseStmt represents a `use` instruction followed by a module name identifier.
type UseStmt struct {
	ModuleName string
	Line       int
	Column     int
}

func (s *UseStmt) statementNode()       {}
func (s *UseStmt) StatementLine() int   { return s.Line }
func (s *UseStmt) StatementColumn() int { return s.Column }
