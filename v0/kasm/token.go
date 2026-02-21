package kasm

type Token struct {
	Type    TokenType // The type of the token (e.g., identifier, instruction, register).
	Literal string    // The literal value of the token (e.g., the actual text from the source code).
	Line    int       // The line number where the token was found (for error reporting).
	Column  int       // The column number where the token starts (for error reporting).
}
