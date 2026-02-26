package kasm

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

// ---------------------------------------------------------------------------
// Program (root AST node)
// ---------------------------------------------------------------------------

// Program is the root AST node. It holds an ordered slice of Statement nodes
// representing every top-level construct in the source. Statements appear in
// the same order as the corresponding tokens in the input slice.
type Program struct {
	Statements []Statement
}

// ---------------------------------------------------------------------------
// Statement types
// ---------------------------------------------------------------------------

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

// NamespaceStmt represents a `namespace` keyword followed by a name
// identifier.
type NamespaceStmt struct {
	Name   string
	Line   int
	Column int
}

func (s *NamespaceStmt) statementNode()       {}
func (s *NamespaceStmt) StatementLine() int   { return s.Line }
func (s *NamespaceStmt) StatementColumn() int { return s.Column }

// UseStmt represents a `use` instruction followed by a module name
// identifier.
type UseStmt struct {
	ModuleName string
	Line       int
	Column     int
}

func (s *UseStmt) statementNode()       {}
func (s *UseStmt) StatementLine() int   { return s.Line }
func (s *UseStmt) StatementColumn() int { return s.Column }

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

// SectionStmt represents a `section` keyword followed by a section name
// identifier. The Name field stores the section name with any trailing colon
// stripped — consistent with label-name handling (FR-3.5.2).
type SectionStmt struct {
	Name   string
	Line   int
	Column int
}

func (s *SectionStmt) statementNode()       {}
func (s *SectionStmt) StatementLine() int   { return s.Line }
func (s *SectionStmt) StatementColumn() int { return s.Column }

// ---------------------------------------------------------------------------
// Operand types
// ---------------------------------------------------------------------------

// RegisterOperand wraps a TokenRegister. The Name preserves original casing.
type RegisterOperand struct {
	Name   string
	Line   int
	Column int
}

func (o *RegisterOperand) operandNode()       {}
func (o *RegisterOperand) OperandLine() int   { return o.Line }
func (o *RegisterOperand) OperandColumn() int { return o.Column }

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

// IdentifierOperand wraps a TokenIdentifier that is not a label, comma, or
// bracket. This covers symbolic references such as label names and data
// symbols.
type IdentifierOperand struct {
	Name   string
	Line   int
	Column int
}

func (o *IdentifierOperand) operandNode()       {}
func (o *IdentifierOperand) OperandLine() int   { return o.Line }
func (o *IdentifierOperand) OperandColumn() int { return o.Column }

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

// MemoryComponent represents a single element inside a memory operand
// bracket expression. It holds the raw token (register, immediate,
// identifier, or operator +/-).
type MemoryComponent struct {
	Token Token
}

// MemoryOperand represents a memory reference enclosed in [ and ]. The
// Components slice holds the inner tokens in order, preserving operators
// (+ / -).
type MemoryOperand struct {
	Components []MemoryComponent
	Line       int
	Column     int
}

func (o *MemoryOperand) operandNode()       {}
func (o *MemoryOperand) OperandLine() int   { return o.Line }
func (o *MemoryOperand) OperandColumn() int { return o.Column }
