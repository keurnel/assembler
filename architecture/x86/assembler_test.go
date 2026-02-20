package x86_test

import (
	"testing"

	"github.com/keurnel/assembler/architecture/x86"
)

func TestAssembler_ArchitectureName(t *testing.T) {
	if x86.AssemblerNew().ArchitectureName() != "x86" {
		t.Errorf("Expected ArchitectureName to return 'x86', got %q", x86.AssemblerNew().ArchitectureName())
	}
}

func TestAssembler_Directives(t *testing.T) {
	assembler := x86.AssemblerNew()
	directives := assembler.Directives()

	// Directives map should not be empty
	if len(directives) == 0 {
		t.Error("Expected Directives to return a non-empty map, got an empty map")
	}

	for directive := range directives {
		t.Run("Test directive line "+directive, func(t *testing.T) {
			if !assembler.IsDirective(directive) {
				t.Errorf("Expected %q to be recognized as a valid directive, but it was not", directive)
			}
		})
	}
}

func TestAssembler_IsDirective(t *testing.T) {
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
			assembler := x86.AssemblerNew()
			result := assembler.IsDirective(scenario.line)
			if result != scenario.expected {
				t.Errorf("Expected IsDirective(%q) to be %v, got %v", scenario.line, scenario.expected, result)
			}
		})
	}
}
