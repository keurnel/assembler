package kasm

import (
	"fmt"
	"strings"

	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/v0/kasm/profile"
)

// Lexer holds the input, position state, architecture profile, and the
// accumulating token slice. If a Lexer value exists, it is guaranteed to hold
// a valid input string, a valid profile, initialised position state, and an
// empty token slice. There is no uninitialised or partially-constructed state.
type Lexer struct {
	Input        string // The input source code to be tokenized.
	Position     int    // Current position in the input (points to the current character).
	ReadPosition int    // Current reading position in the input (after the current character).
	Ch           byte   // Current character being examined.

	Line   int // Current line number (for error reporting).
	Column int // Current column number (for error reporting).

	Tokens   []Token
	profile  profile.ArchitectureProfile // Architecture-specific vocabulary for classification.
	debugCtx *debugcontext.DebugContext  // Optional debug context for diagnostic recording. May be nil.
}

// LexerNew is the sole constructor. It accepts the pre-processed source string
// and an ArchitectureProfile, and returns a *Lexer that is ready for Start()
// to be called. There is no separate Init() or SetProfile() step.
//
// LexerNew is infallible — it cannot fail. Any valid string (including the
// empty string) is accepted. The profile must not be nil; passing nil may
// panic — this is a programming error, not a runtime error.
func LexerNew(input string, p profile.ArchitectureProfile) *Lexer {
	l := &Lexer{
		Input:        input,
		Position:     0,
		ReadPosition: 0,
		Ch:           0,
		Line:         1,
		Column:       0,
		Tokens:       make([]Token, 0),
		profile:      p,
	}
	l.readChar()
	return l
}

// WithDebugContext attaches a debug context to the lexer for diagnostic
// recording. When set, the lexer records trace entries (e.g. token count)
// and warnings (e.g. unterminated strings) into the context. When nil,
// the lexer operates silently. Returns the lexer for chaining.
func (l *Lexer) WithDebugContext(ctx *debugcontext.DebugContext) *Lexer {
	l.debugCtx = ctx
	return l
}

// previousTokenType - returns the type of the most recently emitted token, or
// -1 if no tokens have been emitted yet. Because the Tokens slice is
// initialised as non-nil and empty, a length check is sufficient — no nil
// guard is needed.
func (l *Lexer) previousTokenType() TokenType {
	if len(l.Tokens) == 0 {
		return -1 // No tokens emitted yet
	}
	return l.Tokens[len(l.Tokens)-1].Type
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

// peekChar - returns the next character without advancing the position.
func (l *Lexer) peekChar() byte {
	if l.ReadPosition >= len(l.Input) {
		return 0
	}
	return l.Input[l.ReadPosition]
}

// skipWhitespace - advances past spaces, tabs, carriage returns and newlines.
// No TokenWhitespace token is emitted. Because skipWhitespace() loops until a
// non-whitespace character is found, a single invocation handles any run.
func (l *Lexer) skipWhitespace() {
	for l.Ch == ' ' || l.Ch == '\t' || l.Ch == '\r' || l.Ch == '\n' {
		l.readChar()
	}
}

// skipComment - advances past a comment starting with ';' to the end of the line.
// No TokenComment token is emitted.
func (l *Lexer) skipComment() {
	for l.Ch != '\n' && l.Ch != 0 {
		l.readChar()
	}
}

// readWord - reads a contiguous word composed of letters, digits, underscores and dots.
func (l *Lexer) readWord() string {
	start := l.Position
	for isWordChar(l.Ch) {
		l.readChar()
	}
	return l.Input[start:l.Position]
}

// readNumber - reads a numeric literal (decimal or hexadecimal 0x...).
func (l *Lexer) readNumber() string {
	start := l.Position
	if l.Ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar() // '0'
		l.readChar() // 'x'
		for isHexDigit(l.Ch) {
			l.readChar()
		}
	} else {
		for isDigit(l.Ch) {
			l.readChar()
		}
	}
	return l.Input[start:l.Position]
}

// readString - reads a string literal enclosed in double quotes.
func (l *Lexer) readString() string {
	startLine := l.Line
	startCol := l.Column
	l.readChar() // skip opening '"'
	start := l.Position
	for l.Ch != '"' && l.Ch != 0 {
		l.readChar()
	}
	str := l.Input[start:l.Position]
	if l.Ch == '"' {
		l.readChar() // skip closing '"'
	} else if l.debugCtx != nil {
		l.debugCtx.Warning(
			l.debugCtx.Loc(startLine, startCol),
			"unterminated string literal",
		)
	}
	return str
}

// readComment - reads from ';' to the end of the line.
func (l *Lexer) readComment() string {
	start := l.Position
	for l.Ch != '\n' && l.Ch != 0 {
		l.readChar()
	}
	return l.Input[start:l.Position]
}

// readDirective - reads a '%'-prefixed directive word (e.g. %define, %include).
func (l *Lexer) readDirective() string {
	start := l.Position
	l.readChar() // skip '%'
	for isWordChar(l.Ch) {
		l.readChar()
	}
	return l.Input[start:l.Position]
}

// classifyWord determines whether a word is a section keyword, register,
// instruction, keyword, or identifier by consulting the lexer's
// ArchitectureProfile. Because the profile supplies the vocabulary (FR-1),
// the lexer core has no hardcoded knowledge of any specific register or
// instruction name.
//
// When the previous token is a TokenKeyword or TokenSection, the current word
// is always classified as TokenIdentifier regardless of lookup results
// (FR-11.2, FR-11.3). This rule takes precedence, so keywords and section
// headers can introduce arbitrary names without the name being misclassified.
//
// The word "section" (case-insensitive) is a lexer-level language construct
// (FR-4.8.1). It is checked before profile lookups and takes precedence over
// any profile entry with the same name (FR-4.8.4).
func classifyWord(word string, lexer *Lexer) TokenType {
	// Context-sensitive override: keyword and section arguments are always identifiers.
	prev := lexer.previousTokenType()
	if prev == TokenKeyword || prev == TokenSection {
		return TokenIdentifier
	}

	lower := strings.ToLower(word)

	// Section keyword — lexer-level construct, checked before profile lookups.
	if lower == "section" {
		return TokenSection
	}

	if lexer.profile.Registers()[lower] {
		return TokenRegister
	}
	if lexer.profile.Instructions()[lower] {
		return TokenInstruction
	}
	if lexer.profile.Keywords()[lower] {
		return TokenKeyword
	}

	return TokenIdentifier
}

// addToken - appends a token to the Tokens slice.
func (l *Lexer) addToken(tokenType TokenType, literal string, line, column int) {
	l.Tokens = append(l.Tokens, Token{
		Type:    tokenType,
		Literal: literal,
		Line:    line,
		Column:  column,
	})
}

// Start - begins the tokenization process and returns a slice of tokens
// found in the input source, in the order they appear.
func (l *Lexer) Start() []Token {
	if l.debugCtx != nil {
		l.debugCtx.SetPhase("lexer")
	}

	for l.Ch != 0 {
		line := l.Line
		col := l.Column

		switch {
		// Whitespace — skip without emitting a token.
		case l.Ch == ' ' || l.Ch == '\t' || l.Ch == '\r' || l.Ch == '\n':
			l.skipWhitespace()

		// Comment — ';' to end of line.
		case l.Ch == ';':
			l.skipComment()

		// Directive — '%' followed by a word.
		case l.Ch == '%':
			directive := l.readDirective()
			l.addToken(TokenDirective, directive, line, col)

		// String literal — "...".
		case l.Ch == '"':
			str := l.readString()
			l.addToken(TokenString, str, line, col)

		// Numeric literal (immediate value).
		case isDigit(l.Ch):
			num := l.readNumber()
			l.addToken(TokenImmediate, num, line, col)

		// Word — could be an instruction, register or identifier.
		case isLetter(l.Ch) || l.Ch == '_' || l.Ch == '.':
			word := l.readWord()
			// A trailing ':' marks a label — consume it and treat the word as an identifier.
			if l.Ch == ':' {
				word += ":"
				l.readChar()
			}
			l.addToken(classifyWord(word, l), word, line, col)

		// Any other single character (commas, brackets, etc.) — emit as identifier.
		default:
			l.addToken(TokenIdentifier, string(l.Ch), line, col)
			l.readChar()
		}
	}

	if l.debugCtx != nil {
		l.debugCtx.Trace(
			l.debugCtx.Loc(0, 0),
			fmt.Sprintf("tokenised %d token(s) from %d byte(s) of input", len(l.Tokens), len(l.Input)),
		)
	}

	return l.Tokens
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isWordChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '_' || ch == '.'
}
