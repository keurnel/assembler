package keurnel_asm

import (
	"testing"
)

func TestLexerNamespace(t *testing.T) {
	input := `namespace my_namespace
MOV AX, 0x1234
ADD AX, BX`

	lexer := LexerNew(input)
	tokens := lexer.Process()

	// Expected tokens: NAMESPACE, INSTRUCTION, OPERAND, OPERAND, INSTRUCTION, OPERAND, OPERAND, EOF
	expectedTokens := []struct {
		tokenType TokenType
		literal   string
	}{
		{NAMESPACE, "my_namespace"},
		{INSTRUCTION, "MOV"},
		{OPERAND, "AX"},
		{OPERAND, "0x1234"},
		{INSTRUCTION, "ADD"},
		{OPERAND, "AX"},
		{OPERAND, "BX"},
		{EOF, ""},
	}

	if len(tokens) != len(expectedTokens) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	for i, expected := range expectedTokens {
		if tokens[i].Type != expected.tokenType {
			t.Errorf("Token %d: expected type %s, got %s", i, expected.tokenType, tokens[i].Type)
		}
		if tokens[i].Literal != expected.literal {
			t.Errorf("Token %d: expected literal '%s', got '%s'", i, expected.literal, tokens[i].Literal)
		}
	}
}

func TestLexerMultipleNamespaces(t *testing.T) {
	input := `namespace kernel
MOV AX, 0x1234

namespace drivers
ADD BX, CX`

	lexer := LexerNew(input)
	tokens := lexer.Process()

	// Count namespace tokens
	namespaceCount := 0
	for _, token := range tokens {
		if token.Type == NAMESPACE {
			namespaceCount++
		}
	}

	if namespaceCount != 2 {
		t.Errorf("Expected 2 namespace tokens, got %d", namespaceCount)
	}
}

func TestLexerNamespaceWithLabel(t *testing.T) {
	input := `namespace bootloader
start:
MOV AX, 0x1234
ret`

	lexer := LexerNew(input)
	tokens := lexer.Process()

	// Expected: NAMESPACE, LABEL, INSTRUCTION, OPERAND, OPERAND, INSTRUCTION, EOF
	expectedTypes := []TokenType{NAMESPACE, LABEL, INSTRUCTION, OPERAND, OPERAND, INSTRUCTION, EOF}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, expected := range expectedTypes {
		if tokens[i].Type != expected {
			t.Errorf("Token %d: expected type %s, got %s", i, expected, tokens[i].Type)
		}
	}
}
