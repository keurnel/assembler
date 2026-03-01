package ast

// TokenType identifies the syntactic category of a token.
type TokenType int

const (
	// TokenWhitespace represents a sequence of whitespace characters (spaces, tabs, newlines).
	TokenWhitespace TokenType = iota
	// TokenComment represents a comment in the source code (e.g., lines starting with ';').
	TokenComment
	// TokenIdentifier represents an identifier (e.g., variable names, labels).
	TokenIdentifier
	// TokenDirective represents an assembler directive (e.g., %define, %include).
	TokenDirective
	// TokenInstruction represents an assembly instruction (e.g., MOV, ADD).
	TokenInstruction
	// TokenRegister represents a CPU register (e.g., EAX, R1).
	TokenRegister
	// TokenImmediate represents an immediate value (e.g., numeric literals).
	TokenImmediate
	// TokenString represents a string literal (e.g., "Hello, World!").
	TokenString
	// TokenKeyword represents a reserved keyword in the assembly language (e.g., namespace).
	TokenKeyword
	// TokenSection represents the section keyword (e.g., section .data:).
	TokenSection
)

// ToInt converts the TokenType to its underlying integer value.
func (tT TokenType) ToInt() int {
	return int(tT)
}

// Ignored returns true if the token should be skipped during parsing
// (whitespace and comments).
func (tT TokenType) Ignored() bool {
	switch tT {
	case TokenWhitespace, TokenComment:
		return true
	default:
		return false
	}
}

// Whitespace returns true if the token is a whitespace sequence.
func (tT TokenType) Whitespace() bool { return tT == TokenWhitespace }

// Comment returns true if the token is a comment.
func (tT TokenType) Comment() bool { return tT == TokenComment }

// Identifier returns true if the token is an identifier.
func (tT TokenType) Identifier() bool { return tT == TokenIdentifier }

// Directive returns true if the token is an assembler directive.
func (tT TokenType) Directive() bool { return tT == TokenDirective }

// Instruction returns true if the token is an assembly instruction.
func (tT TokenType) Instruction() bool { return tT == TokenInstruction }

// Register returns true if the token is a CPU register.
func (tT TokenType) Register() bool { return tT == TokenRegister }

// Immediate returns true if the token is an immediate value.
func (tT TokenType) Immediate() bool { return tT == TokenImmediate }

// StringLiteral returns true if the token is a string literal.
func (tT TokenType) StringLiteral() bool { return tT == TokenString }

// Section returns true if the token is a section keyword.
func (tT TokenType) Section() bool { return tT == TokenSection }

// Token carries the syntactic category, raw text, and source location of a
// single lexical unit.
type Token struct {
	Type    TokenType // The syntactic category of the token.
	Literal string    // The raw text from the source code.
	Line    int       // 1-based line number where the token was found.
	Column  int       // 1-based column number where the token starts.
}
