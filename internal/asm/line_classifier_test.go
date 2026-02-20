package asm_test

import (
	"testing"

	"github.com/keurnel/assembler/internal/asm"
)

func TestLineAnalyze(t *testing.T) {
	scenarios := []struct {
		name     string
		line     string
		expected asm.LineCharacteristics
	}{
		{"Empty line", "", asm.LineCharacteristics{IsDirective: false, IsComment: false, ContainsComment: false, IsEmpty: true}},
		{"Whitespace line", "   ", asm.LineCharacteristics{IsDirective: false, IsComment: false, ContainsComment: false, IsEmpty: true}},
		{"Directive line", ".section .data", asm.LineCharacteristics{IsDirective: true, IsComment: false, ContainsComment: false, IsEmpty: false}},
		{"Comment line", "; This is a comment", asm.LineCharacteristics{IsDirective: false, IsComment: true, ContainsComment: true, IsEmpty: false}},
		{"Line with directive and comment", ".section .data ; This is a comment", asm.LineCharacteristics{IsDirective: true, IsComment: false, ContainsComment: true, IsEmpty: false}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := asm.LineAnalyze(scenario.line)
			if result != scenario.expected {
				t.Errorf("Expected LineAnalyze(%q) to be %+v, got %+v", scenario.line, scenario.expected, result)
			}
		})
	}
}

func TestIsDirectiveLine(t *testing.T) {
	scenarios := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Valid directive with no parameters", ".section", true},
		{"Valid directive with parameters", ".data 0x1000", true},
		{"Valid directive with leading whitespace", "   .text", true},
		{"Invalid directive (missing dot)", "section .data", false},
		{"Invalid directive (is label)", ".section:", false},
		{"Invalid directive (only dot)", ".", false},
		{"Valid directive followed by a comment", ".section ; This is a comment", true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := asm.IsDirectiveLine(scenario.line)
			if result != scenario.expected {
				t.Errorf("Expected IsDirectiveLine(%q) to be %v, got %v", scenario.line, scenario.expected, result)
			}
		})
	}
}

func TestIsNotDirectiveLine(t *testing.T) {
	scenarios := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Valid directive with no parameters", ".section", false},
		{"Valid directive with parameters", ".data 0x1000", false},
		{"Valid directive with leading whitespace", "   .text", false},
		{"Invalid directive (missing dot)", "section .data", true},
		{"Invalid directive (is label)", ".section:", true},
		{"Invalid directive (only dot)", ".", true},
		{"Valid directive followed by a comment", ".section ; This is a comment", false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := asm.IsNotDirectiveLine(scenario.line)
			if result != scenario.expected {
				t.Errorf("Expected IsNotDirectiveLine(%q) to be %v, got %v", scenario.line, scenario.expected, result)
			}
		})
	}
}

func TestLineIsComment(t *testing.T) {
	scenarios := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Valid comment with no leading whitespace", "; This is a comment", true},
		{"Valid comment with leading whitespace", "   ; This is a comment", true},
		{"Invalid comment (missing semicolon)", "This is not a comment", false},
		{"Invalid comment (semicolon in the middle)", "This is; not a comment", false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := asm.LineIsComment(scenario.line)
			if result != scenario.expected {
				t.Errorf("Expected LineIsComment(%q) to be %v, got %v", scenario.line, scenario.expected, result)
			}
		})
	}
}
