package lexer_test

import (
	"testing"

	"github.com/keurnel/assembler/architecture/x86_64"
	"github.com/keurnel/assembler/internal/asm"
	"github.com/keurnel/assembler/internal/keurnel_asm/lexer"
)

func TestTokenNew(t *testing.T) {

	// Fails to create token when giving illegal token type.
	//
	t.Run("illegal token type", func(t *testing.T) {
		tokenType := lexer.TokenType("ILLEGAL-TOKEN-TYPE")
		_, err := lexer.TokenNew("MOV", &tokenType)
		if err == nil {
			t.Errorf("Expected error for illegal token type, got nil")
		}

		if err.Value != tokenType {
			t.Errorf("Expected error value to be '%s', got '%s'", tokenType, err.Value)
		}

		if err.Message != "Illegal token type received." {
			t.Errorf("Expected error message to be 'Illegal token type received.', got '%s'", err.Message)
		}
	})

	// Successfully creates token when giving valid token type.
	//
	t.Run("valid token type", func(t *testing.T) {
		tokenType := lexer.INSTRUCTION
		token, err := lexer.TokenNew("MOV", &tokenType)
		if err != nil {
			t.Errorf("Expected no error for valid token type, got %v", err)
		}

		if token.Type != &tokenType {
			t.Errorf("Expected token type to be '%s', got '%s'", tokenType, *token.Type)
		}

		if token.Literal != "MOV" {
			t.Errorf("Expected token literal to be 'MOV', got '%s'", token.Literal)
		}
	})
}

func TestTokenTypeDetermine(t *testing.T) {
	scenariosX86_64 := []struct {
		name         string
		literal      string
		expected     lexer.TokenType
		architecture asm.Architecture
	}{
		// Directives
		{"Directive .data", ".data", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .data with colon", ".data:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .text", ".text", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .text with colon", ".text:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .section", ".section", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .section with colon", ".section:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .global", ".global", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .global with colon", ".global:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .globl", ".globl", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .globl with colon", ".globl:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .bss", ".bss", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .bss with colon", ".bss:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .rodata", ".rodata", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .rodata with colon", ".rodata:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .extern", ".extern", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .extern with colon", ".extern:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .byte", ".byte", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .byte with colon", ".byte:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .word", ".word", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .word with colon", ".word:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .long", ".long", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .long with colon", ".long:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .quad", ".quad", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .quad with colon", ".quad:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .ascii", ".ascii", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .ascii with colon", ".ascii:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .asciz", ".asciz", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .asciz with colon", ".asciz:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .string", ".string", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .string with colon", ".string:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .align", ".align", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .align with colon", ".align:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .balign", ".balign", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .balign with colon", ".balign:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .p2align", ".p2align", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .p2align with colon", ".p2align:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .comm", ".comm", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .comm with colon", ".comm:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .local", ".local", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .local with colon", ".local:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .type", ".type", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .type with colon", ".type:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .size", ".size", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .size with colon", ".size:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .set", ".set", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .set with colon", ".set:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .equ", ".equ", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .equ with colon", ".equ:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .equiv", ".equiv", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .equiv with colon", ".equiv:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .intel_syntax", ".intel_syntax", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .intel_syntax with colon", ".intel_syntax:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .att_syntax", ".att_syntax", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .att_syntax with colon", ".att_syntax:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .file", ".file", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .file with colon", ".file:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .ident", ".ident", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive .ident with colon", ".ident:", lexer.DIRECTIVE, x86_64.New("")},
		{"Directive custom", ".custom:", lexer.DIRECTIVE, x86_64.New("")},

		// Label tokens
		{"Label token with valid format", "main:", lexer.LABEL, x86_64.New("")},
		{"Label token with valid format and leading whitespace", "loop:", lexer.LABEL, x86_64.New("")},
		{"Label token with valid format and trailing whitespace", "function:", lexer.LABEL, x86_64.New("")},
		{"Label token with valid format and surrounding whitespace", "start:", lexer.LABEL, x86_64.New("")},

		// Identifier tokens
		{"Identifier token with valid format", "my_variable", lexer.IDENT, x86_64.New("")},

		// Int tokens
		{"Int token with decimal format", "12345", lexer.INT, x86_64.New("")},
		{"Int token with hexadecimal format", "0x1A2B3C", lexer.INT, x86_64.New("")},
		{"Int token with octal format", "0o755", lexer.INT, x86_64.New("")},
		{"Int token with binary format", "0b11010101", lexer.INT, x86_64.New("")},

		// Float tokens
		{"Float token with decimal format", "3.14", lexer.FLOAT, x86_64.New("")},
		{"Float token with scientific notation", "1.23e-4", lexer.FLOAT, x86_64.New("")},
		{"Float token with leading decimal point", ".5", lexer.FLOAT, x86_64.New("")},
		{"Float token with trailing decimal point", "2.", lexer.FLOAT, x86_64.New("")},

		// Char tokens
		{"Char token with single character", "'A'", lexer.CHAR, x86_64.New("")},
		{"Char token with escaped character", "'\\n'", lexer.CHAR, x86_64.New("")},

		// String tokens
		{"String token with double quotes", "\"Hello, World!\"", lexer.STRING, x86_64.New("")},
		{"String token with single quotes", "'Hello, World!'", lexer.STRING, x86_64.New("")},
		{"String token with escaped characters", "\"Line 1\\nLine 2\"", lexer.STRING, x86_64.New("")},

		// Instruction tokens
		{"Instruction token with valid mnemonic", "MOV", lexer.INSTRUCTION, x86_64.New("")},
		{"Instruction token with valid mnemonic and lowercase", "ADD", lexer.INSTRUCTION, x86_64.New("")},
		{"Instruction token with valid mnemonic and mixed case", "LEA", lexer.INSTRUCTION, x86_64.New("")},

		// Register tokens
		{"Register token with valid format and uppercase", "RAX", lexer.REGISTER, x86_64.New("")},

		// Illegal tokens
		{"Illegal token without dot", "data", lexer.ILLEGAL, x86_64.New("")},
		{"Illegal token with dot but no directive name", ".:", lexer.ILLEGAL, x86_64.New("")},
		{"Illegal token starting with 'kasm'", ".kasmDirective:", lexer.ILLEGAL_RESERVED, x86_64.New("")},
	}

	t.Run("Architecture: x86_64", func(t *testing.T) {
		for _, scenario := range scenariosX86_64 {
			t.Run(scenario.name, func(t *testing.T) {
				tokenType := lexer.TokenTypeDetermine(scenario.literal, &scenario.architecture)
				if tokenType != scenario.expected {
					t.Errorf("Expected token type to be '%s', got '%s'", scenario.expected, tokenType)
				}
			})
		}
	})
}

func TestIsOperand(t *testing.T) {

	scenarios := []struct {
		name         string
		operand      string
		expected     bool
		architecture asm.Architecture
	}{
		{"Valid register operand", "RAX", true, x86_64.New("")},
		{"Invalid operand format", "invalid_operand", false, x86_64.New("")},

		// Base displacement operands
		//
		{"Valid memory operand", "[RBP-8]", true, x86_64.New("")},

		// Base displacement operands with positive offsets
		{"Valid memory operand with positive offset", "[RBP+8]", true, x86_64.New("")},
		{"Valid memory operand with larger positive offset", "[RSP+16]", true, x86_64.New("")},
		{"Valid memory operand with hex offset", "[RBX+0x10]", true, x86_64.New("")},

		// Base displacement operands with negative offsets
		{"Valid memory operand with negative offset", "[RBP-16]", true, x86_64.New("")},
		{"Valid memory operand with larger negative offset", "[RSP-32]", true, x86_64.New("")},

		// Base displacement operands without offset (register indirect)
		{"Valid memory operand register indirect", "[RAX]", true, x86_64.New("")},
		{"Valid memory operand register indirect RDI", "[RDI]", true, x86_64.New("")},

		// Base + index operands
		{"Valid memory operand base plus index", "[RBX+RCX]", true, x86_64.New("")},
		{"Valid memory operand base plus index with offset", "[RBX+RCX+8]", true, x86_64.New("")},

		// Scaled index operands
		{"Valid memory operand with scaled index", "[RBX+RCX*4]", true, x86_64.New("")},
		{"Valid memory operand with scaled index and offset", "[RBX+RCX*8+16]", true, x86_64.New("")},

		// Direct memory addressing
		{"Valid direct memory address", "[0x400000]", true, x86_64.New("")},

		// Invalid operands
		{"Invalid memory operand missing bracket", "RBP-8", false, x86_64.New("")},
		{"Invalid memory operand extra bracket", "[[RBP-8]]", false, x86_64.New("")},

		// Immediate values
		//
		{"Valid immediate operand", "123", true, x86_64.New("")},
	}

	t.Run("Architecture: x86_64", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				isOperand := lexer.IsOperand(scenario.operand, scenario.architecture)
				if isOperand != scenario.expected {
					t.Errorf("Expected IsOperand to return %v for operand '%s', got %v", scenario.expected, scenario.operand, isOperand)
				}
			})
		}
	})
}
