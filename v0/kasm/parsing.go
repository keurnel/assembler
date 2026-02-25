package kasm

import (
	"fmt"
	"strings"

	"github.com/keurnel/assembler/internal/debugcontext"
)

// Parser holds the token slice, current position, and accumulated errors.
// If a Parser value exists, it is guaranteed to hold a valid (possibly empty)
// token slice and initialised position state. There is no uninitialised or
// partially-constructed state.
type Parser struct {
	// Position is the current index into the Tokens slice.
	Position int

	// Tokens is the input token slice from the lexer.
	Tokens []Token

	// errors accumulates parse errors encountered during Parse().
	errors []ParseError

	// debugCtx is an optional debug context for diagnostic recording. May be nil.
	debugCtx *debugcontext.DebugContext
}

// ParserNew is the sole constructor. It accepts the []Token slice produced by
// Lexer.Start() and returns a *Parser that is ready for Parse() to be called.
// ParserNew is infallible — it cannot fail. An empty or nil slice is valid.
func ParserNew(tokens []Token) *Parser {
	if tokens == nil {
		tokens = []Token{}
	}
	return &Parser{
		Position: 0,
		Tokens:   tokens,
		errors:   make([]ParseError, 0),
	}
}

// WithDebugContext attaches a debug context to the parser for diagnostic
// recording. When set, the parser records errors and trace entries into the
// context. When nil, the parser operates silently using only the internal
// error slice. Returns the parser for chaining.
func (p *Parser) WithDebugContext(ctx *debugcontext.DebugContext) *Parser {
	p.debugCtx = ctx
	return p
}

// ---------------------------------------------------------------------------
// Token consumption helpers (FR-4)
// ---------------------------------------------------------------------------

// current returns the token at Position, or a sentinel zero-value Token if
// Position is at or past the end.
func (p *Parser) current() Token {
	if p.Position >= len(p.Tokens) {
		return Token{}
	}
	return p.Tokens[p.Position]
}

// peek returns the token at Position+1 without advancing, or a sentinel
// zero-value Token if no next token exists.
func (p *Parser) peek() Token {
	if p.Position+1 >= len(p.Tokens) {
		return Token{}
	}
	return p.Tokens[p.Position+1]
}

// advance increments Position by one and returns the token that was at the
// previous position. If already at the end, it returns the sentinel zero-value
// Token without advancing further.
func (p *Parser) advance() Token {
	if p.Position >= len(p.Tokens) {
		return Token{}
	}
	tok := p.Tokens[p.Position]
	p.Position++
	return tok
}

// expect checks that the current token matches the expected type. If it
// matches, the token is consumed and returned. If it does not match, a
// ParseError is recorded and the parser does not advance.
func (p *Parser) expect(tokenType TokenType) (Token, bool) {
	tok := p.current()
	if tok.Type == tokenType {
		p.advance()
		return tok, true
	}
	return tok, false
}

// isAtEnd returns true when Position is at or past the length of the token
// slice.
func (p *Parser) isAtEnd() bool {
	return p.Position >= len(p.Tokens)
}

// addError records a parse error at the given position. If a debug context
// is attached, the error is also recorded there.
func (p *Parser) addError(message string, line, column int) {
	p.errors = append(p.errors, ParseError{
		Message: message,
		Line:    line,
		Column:  column,
	})
	if p.debugCtx != nil {
		p.debugCtx.Error(
			p.debugCtx.Loc(line, column),
			message,
		)
	}
}

// addErrorAtCurrent records a parse error at the current token's position.
func (p *Parser) addErrorAtCurrent(message string) {
	tok := p.current()
	p.addError(message, tok.Line, tok.Column)
}

// ---------------------------------------------------------------------------
// Recovery (FR-5)
// ---------------------------------------------------------------------------

// isStatementStart returns true if the given token can begin a new statement.
func isStatementStart(tok Token) bool {
	switch tok.Type {
	case TokenInstruction, TokenKeyword, TokenDirective:
		return true
	case TokenIdentifier:
		// Labels (trailing ':') start a statement.
		return strings.HasSuffix(tok.Literal, ":")
	}
	return false
}

// recover advances past tokens until the start of a recognisable statement is
// found or the end of input is reached. At least one token is consumed to
// guarantee progress.
func (p *Parser) recover() {
	// Consume at least one token to guarantee progress.
	p.advance()
	for !p.isAtEnd() && !isStatementStart(p.current()) {
		p.advance()
	}
}

// ---------------------------------------------------------------------------
// Parse (FR-2)
// ---------------------------------------------------------------------------

// Parse performs a single left-to-right pass over the token slice and returns
// a *Program AST and a slice of ParseError values. It is the sole public
// method that drives parsing.
func (p *Parser) Parse() (*Program, []ParseError) {
	if p.debugCtx != nil {
		p.debugCtx.SetPhase("parser")
	}

	program := &Program{
		Statements: make([]Statement, 0),
	}

	for !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	if p.debugCtx != nil {
		p.debugCtx.Trace(
			p.debugCtx.Loc(0, 0),
			fmt.Sprintf("parsed %d statement(s) with %d error(s) from %d token(s)",
				len(program.Statements), len(p.errors), len(p.Tokens)),
		)
	}

	return program, p.errors
}

// ---------------------------------------------------------------------------
// Statement dispatch (FR-6)
// ---------------------------------------------------------------------------

// parseStatement inspects the current token's type to determine which parsing
// method to invoke. Returns nil if recovery consumed the token instead.
func (p *Parser) parseStatement() Statement {
	tok := p.current()

	switch tok.Type {
	case TokenInstruction:
		return p.parseInstruction()

	case TokenIdentifier:
		if strings.HasSuffix(tok.Literal, ":") {
			return p.parseLabel()
		}
		// Identifier without trailing ':' outside instruction context — parse error.
		p.addErrorAtCurrent("unexpected identifier outside instruction context: " + tok.Literal)
		p.recover()
		return nil

	case TokenKeyword:
		return p.parseKeyword()

	case TokenDirective:
		return p.parseDirective()

	case TokenRegister:
		p.addErrorAtCurrent("unexpected register outside instruction context: " + tok.Literal)
		p.recover()
		return nil

	case TokenImmediate:
		p.addErrorAtCurrent("unexpected immediate value outside instruction context: " + tok.Literal)
		p.recover()
		return nil

	case TokenString:
		p.addErrorAtCurrent("unexpected string literal outside instruction context")
		p.recover()
		return nil

	default:
		p.addErrorAtCurrent("unexpected token: " + tok.Literal)
		p.recover()
		return nil
	}
}

// ---------------------------------------------------------------------------
// Instruction parsing (FR-7)
// ---------------------------------------------------------------------------

// parseInstruction parses a TokenInstruction followed by zero or more operands
// separated by commas. If the mnemonic is "use", it delegates to parseUse.
func (p *Parser) parseInstruction() Statement {
	tok := p.advance()

	// FR-7.6: delegate "use" to UseStmt parsing.
	if strings.EqualFold(tok.Literal, "use") {
		return p.parseUse(tok)
	}

	operands := p.parseOperandList()

	return &InstructionStmt{
		Mnemonic: tok.Literal,
		Operands: operands,
		Line:     tok.Line,
		Column:   tok.Column,
	}
}

// parseOperandList collects zero or more operands separated by commas.
// Parsing continues until a token is encountered that cannot be an operand
// or a comma (i.e. the start of the next statement or end of input).
func (p *Parser) parseOperandList() []Operand {
	operands := make([]Operand, 0)

	for !p.isAtEnd() {
		tok := p.current()

		// Stop if this token starts a new statement.
		if isStatementStart(tok) {
			break
		}

		operand := p.parseOperand()
		if operand == nil {
			break
		}
		operands = append(operands, operand)

		// Check for comma separator.
		if !p.isAtEnd() && p.current().Type == TokenIdentifier && p.current().Literal == "," {
			p.advance() // consume the comma
		}
	}

	return operands
}

// parseOperand parses a single operand based on the current token type.
// Returns nil if the current token cannot be an operand.
func (p *Parser) parseOperand() Operand {
	tok := p.current()

	switch tok.Type {
	case TokenRegister:
		p.advance()
		return &RegisterOperand{Name: tok.Literal, Line: tok.Line, Column: tok.Column}

	case TokenImmediate:
		p.advance()
		return &ImmediateOperand{Value: tok.Literal, Line: tok.Line, Column: tok.Column}

	case TokenString:
		p.advance()
		return &StringOperand{Value: tok.Literal, Line: tok.Line, Column: tok.Column}

	case TokenIdentifier:
		// Opening bracket → memory operand.
		if tok.Literal == "[" {
			return p.parseMemoryOperand()
		}
		// Closing bracket or comma → not an operand, stop.
		if tok.Literal == "]" || tok.Literal == "," {
			return nil
		}
		p.advance()
		return &IdentifierOperand{Name: tok.Literal, Line: tok.Line, Column: tok.Column}

	default:
		return nil
	}
}

// parseMemoryOperand parses a memory reference enclosed in [ and ].
// The opening '[' must be the current token.
func (p *Parser) parseMemoryOperand() Operand {
	openBracket := p.advance() // consume '['
	components := make([]MemoryComponent, 0)

	for !p.isAtEnd() {
		tok := p.current()

		// Closing bracket — consume and return.
		if tok.Type == TokenIdentifier && tok.Literal == "]" {
			p.advance()
			return &MemoryOperand{
				Components: components,
				Line:       openBracket.Line,
				Column:     openBracket.Column,
			}
		}

		// Stop if we hit a statement boundary (unterminated bracket).
		if isStatementStart(tok) {
			break
		}

		// Collect the component.
		components = append(components, MemoryComponent{Token: tok})
		p.advance()
	}

	// Unterminated '[' — record error and return what we have.
	p.addError("unterminated memory operand, expected ']'", openBracket.Line, openBracket.Column)
	return &MemoryOperand{
		Components: components,
		Line:       openBracket.Line,
		Column:     openBracket.Column,
	}
}

// ---------------------------------------------------------------------------
// Label parsing (FR-8)
// ---------------------------------------------------------------------------

// parseLabel parses a TokenIdentifier whose literal ends with ':'.
func (p *Parser) parseLabel() Statement {
	tok := p.advance()
	// Strip the trailing ':' to produce the semantic label name.
	name := tok.Literal[:len(tok.Literal)-1]
	return &LabelStmt{
		Name:   name,
		Line:   tok.Line,
		Column: tok.Column,
	}
}

// ---------------------------------------------------------------------------
// Keyword dispatch
// ---------------------------------------------------------------------------

// parseKeyword dispatches by keyword literal.
func (p *Parser) parseKeyword() Statement {
	tok := p.current()

	if strings.EqualFold(tok.Literal, "namespace") {
		return p.parseNamespace()
	}

	// Unknown keyword.
	p.addErrorAtCurrent("unknown keyword: " + tok.Literal)
	p.recover()
	return nil
}

// ---------------------------------------------------------------------------
// Namespace parsing (FR-9)
// ---------------------------------------------------------------------------

// parseNamespace parses a `namespace` keyword followed by an identifier.
func (p *Parser) parseNamespace() Statement {
	kwTok := p.advance() // consume 'namespace' keyword

	if p.isAtEnd() {
		p.addError("expected namespace name, got end of input", kwTok.Line, kwTok.Column)
		return nil
	}

	nameTok, ok := p.expect(TokenIdentifier)
	if !ok {
		p.addError("expected namespace name, got "+nameTok.Literal, nameTok.Line, nameTok.Column)
		return nil
	}

	return &NamespaceStmt{
		Name:   nameTok.Literal,
		Line:   kwTok.Line,
		Column: kwTok.Column,
	}
}

// ---------------------------------------------------------------------------
// Use parsing (FR-10)
// ---------------------------------------------------------------------------

// parseUse parses a `use` instruction followed by a module name identifier.
// The `use` token has already been consumed by the caller.
func (p *Parser) parseUse(useTok Token) Statement {
	if p.isAtEnd() {
		p.addError("expected module name after 'use', got end of input", useTok.Line, useTok.Column)
		return nil
	}

	nameTok, ok := p.expect(TokenIdentifier)
	if !ok {
		p.addError("expected module name after 'use', got "+nameTok.Literal, nameTok.Line, nameTok.Column)
		return nil
	}

	return &UseStmt{
		ModuleName: nameTok.Literal,
		Line:       useTok.Line,
		Column:     useTok.Column,
	}
}

// ---------------------------------------------------------------------------
// Directive parsing (FR-11)
// ---------------------------------------------------------------------------

// parseDirective parses a TokenDirective and collects any argument tokens
// that follow on the same logical statement.
func (p *Parser) parseDirective() Statement {
	dirTok := p.advance() // consume the directive

	args := make([]Token, 0)
	for !p.isAtEnd() {
		tok := p.current()
		// Stop if this token starts a new statement.
		if isStatementStart(tok) {
			break
		}
		// Stop on another directive — that starts a new directive statement.
		if tok.Type == TokenDirective {
			break
		}
		args = append(args, tok)
		p.advance()
	}

	return &DirectiveStmt{
		Literal: dirTok.Literal,
		Args:    args,
		Line:    dirTok.Line,
		Column:  dirTok.Column,
	}
}
