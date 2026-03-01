package ast

// DirectiveStmt represents a pre-processor directive that survived into the
// token stream. The Literal field includes the `%` prefix. Args holds any
// argument tokens that follow the directive on the same logical statement.
type DirectiveStmt struct {
	Literal string
	Args    []Token
	Line    int
	Column  int
}

func (s *DirectiveStmt) statementNode()       {}
func (s *DirectiveStmt) StatementLine() int   { return s.Line }
func (s *DirectiveStmt) StatementColumn() int { return s.Column }
