package kasm

type Lexer struct {
	Input        string // The input source code to be tokenized.
	Position     int    // Current position in the input (points to the current character).
	ReadPosition int    // Current reading position in the input (after the current character).
	Ch           byte   // Current character being examined.

	Line   int // Current line number (for error reporting).
	Column int // Current column number (for error reporting).

	Tokens map[string]Token
}

func LexerNew(input string) *Lexer {
	l := &Lexer{
		Input:        input,
		Position:     0,
		ReadPosition: 0,
		Ch:           0,
		Line:         1,
		Column:       0,
		Tokens:       make(map[string]Token),
	}
	l.readChar()
	return l
}

// readChar - reads the next character from the input and advances the positions accordingly.
func (l *Lexer) readChar() {
	if l.ReadPosition >= len(l.Input) {
		l.Ch = 0 // ASCII code for NUL, indicates end of input.
	} else {
		l.Ch = l.Input[l.ReadPosition]
	}
	l.Position = l.ReadPosition
	l.ReadPosition++

	// Update line and column numbers for error reporting.
	if l.Ch == '\n' {
		l.Line++
		l.Column = 0
	} else {
		l.Column++
	}
}

// Start - begins the tokenization process and returns a map of tokens
// found in the input source.
func (l *Lexer) Start() map[string]Token {

	return l.Tokens
}
