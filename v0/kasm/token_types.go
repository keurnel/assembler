package kasm

import "github.com/keurnel/assembler/v0/kasm/ast"

// Token type constants — re-exported from ast for backward compatibility.
const (
	TokenWhitespace  = ast.TokenWhitespace
	TokenComment     = ast.TokenComment
	TokenIdentifier  = ast.TokenIdentifier
	TokenDirective   = ast.TokenDirective
	TokenInstruction = ast.TokenInstruction
	TokenRegister    = ast.TokenRegister
	TokenImmediate   = ast.TokenImmediate
	TokenString      = ast.TokenString
	TokenKeyword     = ast.TokenKeyword
	TokenSection     = ast.TokenSection
)
