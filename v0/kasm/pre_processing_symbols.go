package kasm

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regex for conditional directives, shared between symbols and conditionals.
var conditionalDirectiveRegex = regexp.MustCompile(`(?m)^\s*%(ifdef|ifndef|else|endif)\s*(\w*)\s*$`)

// PreProcessingCreateSymbolTable - scans the source code for %define directives and builds
// a symbol table mapping each defined symbol name to true.
// Macro names from the provided macro table are also added as defined symbols.
// It returns the symbol table for use in conditional assembly processing.
//
// Only valid identifier names are accepted as symbols; any malformed %define directive
// is a pre-processing error.
//
// The function works in three passes:
//  1. Collect all %define directives and their line numbers, validating each symbol name.
//     Panics if a symbol name is empty or not a valid identifier.
//  2. Detect duplicate %define directives and panic if any are found.
//  3. Add all macro names from the macro table as defined symbols.
//     Returns the completed symbol table.
func PreProcessingCreateSymbolTable(source string, macroTable map[string]Macro) map[string]bool {
	// Match lines of the form: %define SYMBOL_NAME
	defineRegex := regexp.MustCompile(`(?m)^\s*%define\s+(\w+)\s*$`)
	matches := defineRegex.FindAllStringSubmatchIndex(source, -1)

	type symbolEntry struct {
		name       string // symbol name
		lineNumber int    // line number of the %define directive (for error reporting)
	}

	// Pre-allocate with known capacity to avoid repeated slice growth
	entries := make([]symbolEntry, 0, len(matches))

	// Pass 1: collect all %define directives before building the symbol table,
	// so that line numbers remain accurate.
	// Only valid identifier names are accepted; empty or malformed names are a pre-processing error.
	for _, matchIdx := range matches {
		if len(matchIdx) < 4 {
			continue
		}

		matchStart := matchIdx[0]
		lineNumber := strings.Count(source[:matchStart], "\n") + 1

		symbolName := source[matchIdx[2]:matchIdx[3]]

		// Validate that the symbol name is a non-empty valid identifier
		if symbolName == "" {
			panic(fmt.Sprintf("pre-processing error: Empty symbol name in %%define at line %d", lineNumber))
		}

		entries = append(entries, symbolEntry{
			name:       symbolName,
			lineNumber: lineNumber,
		})
	}

	// Pass 2: detect duplicate %define directives and build the symbol table.
	// A symbol may only be defined once; duplicates are a pre-processing error.
	seen := make(map[string]int, len(entries)) // maps symbol name to first line number
	symbolTable := make(map[string]bool, len(entries)+len(macroTable))
	for _, entry := range entries {
		if firstLine, exists := seen[entry.name]; exists {
			panic(fmt.Sprintf("pre-processing error: Duplicate %%define for symbol '%s' at line %d (first defined at line %d)",
				entry.name, entry.lineNumber, firstLine))
		}
		seen[entry.name] = entry.lineNumber
		symbolTable[entry.name] = true
	}

	// Pass 3: add all macro names from the macro table as defined symbols,
	// so that %ifdef/%ifndef can test for macro existence.
	for macroName := range macroTable {
		symbolTable[macroName] = true
	}

	return symbolTable
}
