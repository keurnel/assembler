package kasm

const (
	// TokenWhitespace - represents a sequence of whitespace characters (spaces, tabs, newlines).
	TokenWhitespace TokenType = iota
	// TokenComment - represents a comment in the source code (e.g., lines starting with ';').
	TokenComment
	// TokenIdentifier - represents an identifier (e.g., variable names, labels).
	TokenIdentifier
	// TokenDirective - represents an assembler directive (e.g., %define, %include).
	TokenDirective
	// TokenInstruction - represents an assembly instruction (e.g., MOV, ADD).
	TokenInstruction
	// TokenRegister - represents a CPU register (e.g., EAX, R1).
	TokenRegister
	// TokenImmediate - represents an immediate value (e.g., numeric literals).
	TokenImmediate
	// TokenString - represents a string literal (e.g., "Hello, World!").
	TokenString
)

type TokenType int

// Ignored - used to determine if a token should be ignored during parsing (e.g., whitespace, comments).
func (tT TokenType) Ignored() bool {
	switch tT {
	default:
		return false
	case TokenWhitespace, TokenComment:
		return true
	}
}

// Whitespace - used to determine if a token is a sequence of whitespace characters.
func (tT TokenType) Whitespace() bool {
	return tT == TokenWhitespace
}

// Comment - used to determine if a token is a comment.
func (tT TokenType) Comment() bool {
	return tT == TokenComment
}

// Identifier - used to determine if a token is an identifier.
func (tT TokenType) Identifier() bool {
	return tT == TokenIdentifier
}

// TokenDirective - used to determine if a token is an assembler directive.
func (tT TokenType) Directive() bool {
	return tT == TokenDirective
}

// TokenInstruction - used to determine if a token is an assembly instruction.
func (tT TokenType) Instruction() bool {
	return tT == TokenInstruction
}

// TokenRegister - used to determine if a token is a CPU register.
func (tT TokenType) Register() bool {
	return tT == TokenRegister
}

// TokenImmediate - used to determine if a token is an immediate value.
func (tT TokenType) Immediate() bool {
	return tT == TokenImmediate
}

// TokenString - used to determine if a token is a string literal.
func (tT TokenType) StringLiteral() bool {
	return tT == TokenString
}
