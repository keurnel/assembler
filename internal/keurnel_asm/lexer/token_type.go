package lexer

import (
	"errors"
	"regexp"
)

const (
	ILLEGAL          TokenType = "ILLEGAL"
	ILLEGAL_RESERVED TokenType = "ILLEGAL_RESERVED"
	ILLEGAL_PATTERN  TokenType = "ILLEGAL_PATTERN"
	EOF              TokenType = "EOF"

	// Identifiers and literals
	IDENT  TokenType = "IDENT"  // variable/label names
	INT    TokenType = "INT"    // integer literals: 42, 0x2A, 0b101010
	FLOAT  TokenType = "FLOAT"  // floating point: 3.14
	STRING TokenType = "STRING" // string literals: "hello"
	CHAR   TokenType = "CHAR"   // character literals: 'a'

	// Architecture-specific tokens
	INSTRUCTION TokenType = "INSTRUCTION" // MOV, ADD, SUB, etc.
	REGISTER    TokenType = "REGISTER"    // rax, rbx, eax, etc.

	// Memory and addressing
	MEMORY_REF TokenType = "MEMORY_REF" // [rbx], [rsp+8]
	IMMEDIATE  TokenType = "IMMEDIATE"  // immediate values with prefix

	// Directives
	DIRECTIVE TokenType = "DIRECTIVE" // .data, .text, .section, etc.
	LABEL     TokenType = "LABEL"     // main:, loop:

	// Keurnel-specific
	NAMESPACE TokenType = "NAMESPACE"
	USE       TokenType = "USE"

	// Operators and delimiters
	COMMA    TokenType = "COMMA"    // ,
	COLON    TokenType = "COLON"    // :
	LBRACKET TokenType = "LBRACKET" // [
	RBRACKET TokenType = "RBRACKET" // ]
	LBRACE   TokenType = "LBRACE"   // {
	RBRACE   TokenType = "RBRACE"   // }
	PLUS     TokenType = "PLUS"     // +
	MINUS    TokenType = "MINUS"    // -
	ASTERISK TokenType = "ASTERISK" // *

	// Comments
	COMMENT TokenType = "COMMENT" // ; comment or // comment

	// Special
	NEWLINE TokenType = "NEWLINE" // line breaks (if significant)
	MACRO   TokenType = "MACRO"   // %define, %macro, etc.

	// Errors
	ErrIllegalTokenPrefix  = "illegal-token-prefix"
	ErrIllegalTokenPattern = "illegal-token-pattern"
)

type TokenType string

type InvalidTokenTypeError struct {
	Value   TokenType
	Message string
}

// Valid - verifies if the value of the TokenType is valid. Returns nil if
// valid, otherwise returns an error.
func (t *TokenType) Valid() *InvalidTokenTypeError {
	switch *t {
	default:
		return &InvalidTokenTypeError{
			Value:   *t,
			Message: "Illegal token type received.",
		}
	case ILLEGAL, EOF, IDENT, INT, FLOAT, STRING, CHAR, INSTRUCTION, REGISTER,
		MEMORY_REF, IMMEDIATE, DIRECTIVE, LABEL, NAMESPACE, USE,
		COMMA, COLON, LBRACKET, RBRACKET, LBRACE, RBRACE,
		PLUS, MINUS, ASTERISK,
		COMMENT, NEWLINE, MACRO:
		return nil
	}
}

// TokenTypeDetermine - determines the token type of given literal string. The
// literal should already be trimmed of whitespace and comments before being passed
// to this function.
func TokenTypeDetermine(literal string) TokenType {
	// =========================================================
	//
	// Handling of directives
	//
	// =========================================================
	isDirective, err := isDirective(literal)
	if err != nil {
		switch err.Error() {
		default:
			return ILLEGAL
		case ErrIllegalTokenPattern:
			return ILLEGAL_PATTERN
		case ErrIllegalTokenPrefix:
			return ILLEGAL_RESERVED
		}
	}

	if isDirective {
		return DIRECTIVE
	}

	// =========================================================
	//
	// Handling of labels (e.g., main:, loop:, etc.)
	//
	// =========================================================
	isLabel, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*:$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isLabel {
		return LABEL
	}

	// =========================================================
	//
	// Handling of identifiers (e.g., variable names, label names without colon, etc.)
	//
	// =========================================================
	isIdent, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isIdent {

		// Cannot be reserved keywords such as CPU opcodes or machine instructions.
		//
		if isCPUOpcode(literal) {
			return INSTRUCTION
		}

		if isMachineInstruction(literal) {
			return REGISTER
		}

		// When the identifier matches any known blacklisted keyword
		// (e.g., `data`), it cannot be an identifier as this would
		// cause ambiguity with directives. In such case we continue
		// checking other rules.
		identifierBlacklist := map[string]bool{
			"data": true,
		}
		if _, exists := identifierBlacklist[literal]; !exists {
			return IDENT
		}
	}

	// =========================================================
	//
	// Handle int literals (e.g., 42, 0x2A, 0b101010)
	//
	// =========================================================
	isInt, err := regexp.MatchString(`^(0x[0-9a-fA-F]+|0b[01]+|0o[0-7]+|0O[0-7]+|\d+)$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isInt {
		return INT
	}

	// =========================================================
	//
	// Handle float literals (e.g., 3.14, 1.23e-4, .5, 2.)
	//
	// =========================================================
	isFloat, err := regexp.MatchString(`^(\d+\.\d*|\.\d+)([eE][+-]?\d+)?$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isFloat {
		return FLOAT
	}

	// =========================================================
	//
	// Handle char literals (e.g., 'a', '\n')
	//
	// =========================================================
	isChar, err := regexp.MatchString(`^'(\\.|[^\\'])'$`, literal)
	if err != nil {
		return ILLEGAL
	}
	if isChar {
		return CHAR
	}

	// =========================================================
	//
	// Handle string literals (e.g., "hello", 'world', "Line 1\nLine 2")
	//
	// =========================================================
	isString, err := regexp.MatchString(`^("([^"\\]|\\.)*"|'([^'\\]|\\.)*')$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isString {
		return STRING
	}

	// =========================================================
	//
	// Instruction mnemonics
	//
	// =========================================================
	if isCPUOpcode(literal) {
		return INSTRUCTION
	}

	if isMachineInstruction(literal) {
		return REGISTER
	}

	// No match found, return ILLEGAL token type to indicate
	// an unrecognized token was encountered.
	//
	return ILLEGAL
}

// isDirective - checks if the given literal matches a known directive pattern.
func isDirective(literal string) (bool, error) {

	// 1. Check if literal corresponds to any CPU opcode or machine instruction. If
	// it does, then it cannot be a directive.
	//
	if isCPUOpcode(literal) || isMachineInstruction(literal) {
		return false, nil
	}

	// 2. Check if the literal matches known directive patterns. If it exists
	// in the known directives map, then it is a directive. Otherwise, continue
	// checking other rules.
	//
	knownDirectives := map[string]bool{
		".data":         true,
		".data:":        true,
		".text":         true,
		".text:":        true,
		".section":      true,
		".section:":     true,
		".global":       true,
		".global:":      true,
		".globl":        true, // alternative spelling
		".globl:":       true,
		".bss":          true, // uninitialized data section
		".bss:":         true,
		".rodata":       true, // read-only data section
		".rodata:":      true,
		".extern":       true, // external symbol declaration
		".extern:":      true,
		".byte":         true, // define byte
		".word":         true, // define word (2 bytes)
		".long":         true, // define long (4 bytes)
		".quad":         true, // define quad (8 bytes)
		".ascii":        true, // ASCII string
		".asciz":        true, // null-terminated ASCII string
		".string":       true, // string data
		".align":        true, // alignment directive
		".balign":       true, // byte alignment
		".p2align":      true, // power-of-2 alignment
		".comm":         true, // common symbol
		".local":        true, // local symbol
		".type":         true, // symbol type
		".size":         true, // symbol size
		".set":          true, // set symbol value
		".equ":          true, // equate symbol
		".equiv":        true, // equivalent symbol
		".intel_syntax": true, // Intel syntax mode
		".att_syntax":   true, // AT&T syntax mode
		".file":         true, // source file info
		".ident":        true, // identification string
	}
	if _, exists := knownDirectives[literal]; exists {
		return true, nil
	}

	// 3. Check if the literal starts with a single dot followed by characters (e.g., my_directive) and does
	// end with a colon (`:`). If it does, then it is a directive.
	//
	matched, err := regexp.MatchString(`^\.[a-zA-Z_][a-zA-Z0-9_]*:$`, literal)
	if err != nil {
		return false, errors.New(ErrIllegalTokenPattern)
	}

	if matched {

		// Directives cannot start with `.kasm` as this is reserved for Keurnel-specific directives.
		//
		if regexp.MustCompile(`^\.kasm`).MatchString(literal) {
			return false, errors.New(ErrIllegalTokenPrefix)
		}

		return true, nil
	}

	return matched, nil
}

// isCPUOpcode - checks if the given literal matches a known CPU opcode.
func isCPUOpcode(literal string) bool {
	if _, exists := X86_64Opcodes[literal]; exists {
		return true
	}

	return false
}

// isMachineInstruction - checks if the given literal matches a known machine instruction.
func isMachineInstruction(literal string) bool {
	// This is a simplified check. In a real implementation, you would have a comprehensive list of machine instructions.
	instructions := []string{"RAX", "RBX", "EAX", "EBX"}
	for _, instr := range instructions {
		if literal == instr {
			return true
		}
	}
	return false
}
