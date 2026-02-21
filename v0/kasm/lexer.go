package kasm

import "strings"

// knownRegisters - set of recognised x86_64 register names (lower-case).
var knownRegisters = map[string]bool{
	// 64-bit general-purpose
	"rax": true, "rbx": true, "rcx": true, "rdx": true,
	"rsi": true, "rdi": true, "rbp": true, "rsp": true,
	"r8": true, "r9": true, "r10": true, "r11": true,
	"r12": true, "r13": true, "r14": true, "r15": true,
	// 32-bit general-purpose
	"eax": true, "ebx": true, "ecx": true, "edx": true,
	"esi": true, "edi": true, "ebp": true, "esp": true,
	"r8d": true, "r9d": true, "r10d": true, "r11d": true,
	"r12d": true, "r13d": true, "r14d": true, "r15d": true,
	// 16-bit general-purpose
	"ax": true, "bx": true, "cx": true, "dx": true,
	"si": true, "di": true, "bp": true, "sp": true,
	// 8-bit
	"al": true, "bl": true, "cl": true, "dl": true,
	"ah": true, "bh": true, "ch": true, "dh": true,
	"sil": true, "dil": true, "bpl": true, "spl": true,
	// Segment registers
	"cs": true, "ds": true, "es": true, "fs": true, "gs": true, "ss": true,
	// Instruction pointer / flags
	"rip": true, "eip": true, "rflags": true, "eflags": true,
}

// knownInstructions - set of recognised x86_64 instruction mnemonics (lower-case).
var knownInstructions = map[string]bool{
	"mov": true, "movzx": true, "movsx": true, "lea": true,
	"push": true, "pop": true, "xchg": true,
	"add": true, "sub": true, "mul": true, "imul": true,
	"div": true, "idiv": true, "inc": true, "dec": true, "neg": true,
	"and": true, "or": true, "xor": true, "not": true,
	"shl": true, "shr": true, "sal": true, "sar": true,
	"rol": true, "ror": true,
	"cmp": true, "test": true,
	"jmp": true, "je": true, "jne": true, "jz": true, "jnz": true,
	"jg": true, "jge": true, "jl": true, "jle": true,
	"ja": true, "jae": true, "jb": true, "jbe": true,
	"call": true, "ret": true, "syscall": true, "int": true,
	"nop": true, "hlt": true, "cli": true, "sti": true,
	"loop": true, "loope": true, "loopne": true,
	"cmove": true, "cmovne": true, "cmovg": true, "cmovl": true,
	"sete": true, "setne": true, "setg": true, "setl": true,
	"rep": true, "movsb": true, "stosb": true,
	"cbw": true, "cwd": true, "cdq": true, "cqo": true,
	"use": true,
}

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

// peekChar - returns the next character without advancing the position.
func (l *Lexer) peekChar() byte {
	if l.ReadPosition >= len(l.Input) {
		return 0
	}
	return l.Input[l.ReadPosition]
}

// skipWhitespace - advances past spaces, tabs, carriage returns and newlines,
// collecting them into a single whitespace token.
func (l *Lexer) skipWhitespace() {
	for l.Ch == ' ' || l.Ch == '\t' || l.Ch == '\r' || l.Ch == '\n' {
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
	l.readChar() // skip opening '"'
	start := l.Position
	for l.Ch != '"' && l.Ch != 0 {
		l.readChar()
	}
	str := l.Input[start:l.Position]
	if l.Ch == '"' {
		l.readChar() // skip closing '"'
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

// classifyWord - determines whether a word is a register, instruction or identifier.
func classifyWord(word string) TokenType {
	lower := strings.ToLower(word)
	if knownRegisters[lower] {
		return TokenRegister
	}
	if knownInstructions[lower] {
		return TokenInstruction
	}
	return TokenIdentifier
}

// addToken - stores a token in the Tokens map, keyed by "line:column" to
// preserve every occurrence (including duplicate literals).
func (l *Lexer) addToken(tokenType TokenType, literal string, line, column int) {
	key := strings.Join([]string{
		strings.Repeat("0", 6-len(itoa(line))) + itoa(line),
		strings.Repeat("0", 4-len(itoa(column))) + itoa(column),
	}, ":")
	l.Tokens[key] = Token{
		Type:    tokenType,
		Literal: literal,
		Line:    line,
		Column:  column,
	}
}

// Start - begins the tokenization process and returns a map of tokens
// found in the input source. Tokens are keyed by "line:column" so every
// occurrence is preserved even when the same literal appears multiple times.
func (l *Lexer) Start() map[string]Token {
	for l.Ch != 0 {
		line := l.Line
		col := l.Column

		switch {
		// Whitespace — skip without emitting a token.
		case l.Ch == ' ' || l.Ch == '\t' || l.Ch == '\r' || l.Ch == '\n':
			l.skipWhitespace()

		// Comment — everything from ';' to end of line.
		case l.Ch == ';':
			comment := l.readComment()
			l.addToken(TokenComment, comment, line, col)

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
			l.addToken(classifyWord(word), word, line, col)

		// Any other single character (commas, brackets, etc.) — emit as identifier.
		default:
			l.addToken(TokenIdentifier, string(l.Ch), line, col)
			l.readChar()
		}
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

// itoa - minimal int-to-string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
