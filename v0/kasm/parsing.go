package kasm

type Parser struct {
	// Position of the parser
	//
	Position int

	// Input tokens
	//
	Tokens []Token
}

// ParserNew - creates a new Parser instance with the given input source code.
func ParserNew(tokens []Token) *Parser {
	return &Parser{
		Tokens: tokens,
	}
}

// Parse - processes the input tokens and constructs an abstract syntax tree (AST) as
// intermediate representation of the assembly code.
func (p *Parser) Parse() error {

	// Print each token
	for _, token := range p.Tokens {
		println("Token Type:", token.Type.ToInt(), "Literal:", token.Literal, "Line:", token.Line, "Column:", token.Column)
	}

	return nil
}
