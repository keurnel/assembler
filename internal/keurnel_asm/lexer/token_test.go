package lexer_test

import (
	"testing"

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
	scenarios := []struct {
		name     string
		literal  string
		expected lexer.TokenType
	}{
		// Directives
		// Directives
		{"Directive .data", ".data", lexer.DIRECTIVE},
		{"Directive .data with colon", ".data:", lexer.DIRECTIVE},
		{"Directive .text", ".text", lexer.DIRECTIVE},
		{"Directive .text with colon", ".text:", lexer.DIRECTIVE},
		{"Directive .section", ".section", lexer.DIRECTIVE},
		{"Directive .section with colon", ".section:", lexer.DIRECTIVE},
		{"Directive .global", ".global", lexer.DIRECTIVE},
		{"Directive .global with colon", ".global:", lexer.DIRECTIVE},
		{"Directive .globl", ".globl", lexer.DIRECTIVE},
		{"Directive .globl with colon", ".globl:", lexer.DIRECTIVE},
		{"Directive .bss", ".bss", lexer.DIRECTIVE},
		{"Directive .bss with colon", ".bss:", lexer.DIRECTIVE},
		{"Directive .rodata", ".rodata", lexer.DIRECTIVE},
		{"Directive .rodata with colon", ".rodata:", lexer.DIRECTIVE},
		{"Directive .extern", ".extern", lexer.DIRECTIVE},
		{"Directive .extern with colon", ".extern:", lexer.DIRECTIVE},
		{"Directive .byte", ".byte", lexer.DIRECTIVE},
		{"Directive .byte with colon", ".byte:", lexer.DIRECTIVE},
		{"Directive .word", ".word", lexer.DIRECTIVE},
		{"Directive .word with colon", ".word:", lexer.DIRECTIVE},
		{"Directive .long", ".long", lexer.DIRECTIVE},
		{"Directive .long with colon", ".long:", lexer.DIRECTIVE},
		{"Directive .quad", ".quad", lexer.DIRECTIVE},
		{"Directive .quad with colon", ".quad:", lexer.DIRECTIVE},
		{"Directive .ascii", ".ascii", lexer.DIRECTIVE},
		{"Directive .ascii with colon", ".ascii:", lexer.DIRECTIVE},
		{"Directive .asciz", ".asciz", lexer.DIRECTIVE},
		{"Directive .asciz with colon", ".asciz:", lexer.DIRECTIVE},
		{"Directive .string", ".string", lexer.DIRECTIVE},
		{"Directive .string with colon", ".string:", lexer.DIRECTIVE},
		{"Directive .align", ".align", lexer.DIRECTIVE},
		{"Directive .align with colon", ".align:", lexer.DIRECTIVE},
		{"Directive .balign", ".balign", lexer.DIRECTIVE},
		{"Directive .balign with colon", ".balign:", lexer.DIRECTIVE},
		{"Directive .p2align", ".p2align", lexer.DIRECTIVE},
		{"Directive .p2align with colon", ".p2align:", lexer.DIRECTIVE},
		{"Directive .comm", ".comm", lexer.DIRECTIVE},
		{"Directive .comm with colon", ".comm:", lexer.DIRECTIVE},
		{"Directive .local", ".local", lexer.DIRECTIVE},
		{"Directive .local with colon", ".local:", lexer.DIRECTIVE},
		{"Directive .type", ".type", lexer.DIRECTIVE},
		{"Directive .type with colon", ".type:", lexer.DIRECTIVE},
		{"Directive .size", ".size", lexer.DIRECTIVE},
		{"Directive .size with colon", ".size:", lexer.DIRECTIVE},
		{"Directive .set", ".set", lexer.DIRECTIVE},
		{"Directive .set with colon", ".set:", lexer.DIRECTIVE},
		{"Directive .equ", ".equ", lexer.DIRECTIVE},
		{"Directive .equ with colon", ".equ:", lexer.DIRECTIVE},
		{"Directive .equiv", ".equiv", lexer.DIRECTIVE},
		{"Directive .equiv with colon", ".equiv:", lexer.DIRECTIVE},
		{"Directive .intel_syntax", ".intel_syntax", lexer.DIRECTIVE},
		{"Directive .intel_syntax with colon", ".intel_syntax:", lexer.DIRECTIVE},
		{"Directive .att_syntax", ".att_syntax", lexer.DIRECTIVE},
		{"Directive .att_syntax with colon", ".att_syntax:", lexer.DIRECTIVE},
		{"Directive .file", ".file", lexer.DIRECTIVE},
		{"Directive .file with colon", ".file:", lexer.DIRECTIVE},
		{"Directive .ident", ".ident", lexer.DIRECTIVE},
		{"Directive .ident with colon", ".ident:", lexer.DIRECTIVE},
		{"Directive custom", ".custom:", lexer.DIRECTIVE},

		// Label tokens
		{"Label token with valid format", "main:", lexer.LABEL},
		{"Label token with valid format and leading whitespace", "loop:", lexer.LABEL},
		{"Label token with valid format and trailing whitespace", "function:", lexer.LABEL},
		{"Label token with valid format and surrounding whitespace", "start:", lexer.LABEL},

		// Identifier tokens
		{"Identifier token with valid format", "my_variable", lexer.IDENT},

		// Int tokens
		{"Int token with decimal format", "12345", lexer.INT},
		{"Int token with hexadecimal format", "0x1A2B3C", lexer.INT},
		{"Int token with octal format", "0o755", lexer.INT},
		{"Int token with binary format", "0b11010101", lexer.INT},

		// Float tokens
		{"Float token with decimal format", "3.14", lexer.FLOAT},
		{"Float token with scientific notation", "1.23e-4", lexer.FLOAT},
		{"Float token with leading decimal point", ".5", lexer.FLOAT},
		{"Float token with trailing decimal point", "2.", lexer.FLOAT},

		// Char tokens
		{"Char token with single character", "'A'", lexer.CHAR},
		{"Char token with escaped character", "'\\n'", lexer.CHAR},

		// String tokens
		{"String token with double quotes", "\"Hello, World!\"", lexer.STRING},
		{"String token with single quotes", "'Hello, World!'", lexer.STRING},
		{"String token with escaped characters", "\"Line 1\\nLine 2\"", lexer.STRING},

		// Instruction tokens
		{"Instruction token with valid mnemonic", "MOV", lexer.INSTRUCTION},
		{"Instruction token with valid mnemonic and lowercase", "ADD", lexer.INSTRUCTION},
		{"Instruction token with valid mnemonic and mixed case", "LEA", lexer.INSTRUCTION},

		// Register tokens
		{"Register token with valid format and uppercase", "RAX", lexer.REGISTER},

		// Illegal tokens
		{"Illegal token without dot", "data", lexer.ILLEGAL},
		{"Illegal token with dot but no directive name", ".:", lexer.ILLEGAL},
		{"Illegal token starting with 'kasm'", ".kasmDirective:", lexer.ILLEGAL_RESERVED},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tokenType := lexer.TokenTypeDetermine(scenario.literal)
			if tokenType != scenario.expected {
				t.Errorf("Expected token type to be '%s', got '%s'", scenario.expected, tokenType)
			}
		})
	}

	// Custom without colon should not be recognized as a directive.
	//
	t.Run("Custom without colon", func(t *testing.T) {
		tokenType := lexer.TokenTypeDetermine(".custom")
		if tokenType == lexer.DIRECTIVE {
			t.Errorf("Expected token type to not be 'DIRECTIVE' for '.custom', got '%s'", tokenType)
		}
	})
}
