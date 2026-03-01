package ast

// SectionStmt represents a `section` keyword followed by a section type and
// a section name. The Type field stores the section type (e.g. `.data`,
// `.text`) with any trailing colon stripped. The Name field stores the section
// name verbatim.
type SectionStmt struct {
	Type   string
	Name   string
	Line   int
	Column int
}

func (s *SectionStmt) statementNode()       {}
func (s *SectionStmt) StatementLine() int   { return s.Line }
func (s *SectionStmt) StatementColumn() int { return s.Column }
