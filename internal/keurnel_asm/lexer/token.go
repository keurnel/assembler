package lexer

type Token struct {
	Type    *TokenType
	Literal string
}

// TokenNew - returns a new instance of the Token struct
func TokenNew(literal string, tokenType *TokenType) (*Token, *InvalidTokenTypeError) {
	if err := tokenType.Valid(); err != nil {
		return nil, err
	}

	return &Token{
		Type:    tokenType,
		Literal: literal,
	}, nil
}
