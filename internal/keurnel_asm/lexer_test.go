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

func TestLexerUseStatement(t *testing.T) {
	input := `use my_namespace
MOV AX, 0x1234`

	lexer := LexerNew(input)
	tokens := lexer.Process()

	// Expected tokens: USE, INSTRUCTION, OPERAND, OPERAND, EOF
	expectedTokens := []struct {
		tokenType TokenType
		literal   string
	}{
		{USE, "my_namespace"},
		{INSTRUCTION, "MOV"},
		{OPERAND, "AX"},
		{OPERAND, "0x1234"},
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

func TestLexerMultipleUseStatements(t *testing.T) {
	input := `use kernel
use drivers
MOV AX, 0x1234`

	lexer := LexerNew(input)
	tokens := lexer.Process()

	// Count use tokens
	useCount := 0
	for _, token := range tokens {
		if token.Type == USE {
			useCount++
		}
	}

	if useCount != 2 {
		t.Errorf("Expected 2 use tokens, got %d", useCount)
	}

	// Check the literals
	if tokens[0].Type != USE || tokens[0].Literal != "kernel" {
		t.Errorf("Expected first USE token with literal 'kernel', got type %s with literal '%s'", tokens[0].Type, tokens[0].Literal)
	}
	if tokens[1].Type != USE || tokens[1].Literal != "drivers" {
		t.Errorf("Expected second USE token with literal 'drivers', got type %s with literal '%s'", tokens[1].Type, tokens[1].Literal)
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

func TestParserNamespaceHierarchy(t *testing.T) {
	input := `namespace my_namespace

.start:
    MOV AX, 0x1234
    ret

group:
    MOV BX, 0x5678
    ret

MOV CX, 0x9ABC`

	lexer := LexerNew(input)
	lexer.Process()

	parser := ParserNew(lexer)
	parser.Parse()

	// Check that namespace exists
	namespace, exists := parser.GetGroup("namespace:my_namespace")
	if !exists {
		t.Fatal("Namespace 'my_namespace' not found")
	}

	if namespace.Type != INSTRUCTION_GROUP_TYPE_NAMESPACE {
		t.Errorf("Expected namespace type %d, got %d", INSTRUCTION_GROUP_TYPE_NAMESPACE, namespace.Type)
	}

	// Check that .start is a child of the namespace
	startGroup, exists := namespace.GetChild(".start")
	if !exists {
		t.Fatal("Child group '.start' not found in namespace")
	}

	if startGroup.Type != INSTRUCTION_GROUP_TYPE_DIRECTIVE {
		t.Errorf("Expected directive type %d, got %d", INSTRUCTION_GROUP_TYPE_DIRECTIVE, startGroup.Type)
	}

	if len(startGroup.Instructions) != 2 {
		t.Errorf("Expected 2 instructions in .start, got %d", len(startGroup.Instructions))
	}

	// Check that group: is a child of the namespace
	groupLabel, exists := namespace.GetChild("group:")
	if !exists {
		t.Fatal("Child group 'group:' not found in namespace")
	}

	if groupLabel.Type != INSTRUCTION_GROUP_TYPE_LABEL {
		t.Errorf("Expected label type %d, got %d", INSTRUCTION_GROUP_TYPE_LABEL, groupLabel.Type)
	}

	// Check that global instructions exist
	globalGroup, exists := parser.GetGroup("global")
	if !exists {
		t.Fatal("Global group not found")
	}

	if len(globalGroup.Instructions) != 1 {
		t.Errorf("Expected 1 global instruction, got %d", len(globalGroup.Instructions))
	}

	if globalGroup.Instructions[0].Mnemonic != "MOV" {
		t.Errorf("Expected global instruction 'MOV', got '%s'", globalGroup.Instructions[0].Mnemonic)
	}
}

func TestParserUseStatement(t *testing.T) {
	input := `namespace kernel
_add:
    ADD AX, BX
    ret

namespace app
use kernel

_start:
    MOV AX, 0x1234
    ret`

	lexer := LexerNew(input)
	lexer.Process()

	parser := ParserNew(lexer)
	parser.Parse()

	// Check that app namespace exists and has use statement
	appNamespace, exists := parser.GetGroup("namespace:app")
	if !exists {
		t.Fatal("Namespace 'app' not found")
	}

	if len(appNamespace.Uses) != 1 {
		t.Fatalf("Expected 1 use statement, got %d", len(appNamespace.Uses))
	}

	if appNamespace.Uses[0] != "kernel" {
		t.Errorf("Expected use statement for 'kernel', got '%s'", appNamespace.Uses[0])
	}

	// Verify kernel namespace exists
	kernelNamespace, exists := parser.GetGroup("namespace:kernel")
	if !exists {
		t.Fatal("Namespace 'kernel' not found")
	}

	if len(kernelNamespace.Uses) != 0 {
		t.Errorf("Expected 0 use statements in kernel namespace, got %d", len(kernelNamespace.Uses))
	}
}

func TestParserMultipleUseStatements(t *testing.T) {
	input := `namespace math
_add:
    ADD AX, BX
    ret

namespace io
_print:
    MOV AX, 0x1234
    ret

namespace app
use math
use io

_start:
    MOV AX, 0x1234
    ret`

	lexer := LexerNew(input, x)
	lexer.Process()

	parser := ParserNew(lexer)
	parser.Parse()

	// Check that app namespace has both use statements
	appNamespace, exists := parser.GetGroup("namespace:app")
	if !exists {
		t.Fatal("Namespace 'app' not found")
	}

	if len(appNamespace.Uses) != 2 {
		t.Fatalf("Expected 2 use statements, got %d", len(appNamespace.Uses))
	}

	if appNamespace.Uses[0] != "math" {
		t.Errorf("Expected first use statement for 'math', got '%s'", appNamespace.Uses[0])
	}

	if appNamespace.Uses[1] != "io" {
		t.Errorf("Expected second use statement for 'io', got '%s'", appNamespace.Uses[1])
	}
}
