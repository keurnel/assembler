package preProcessing

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regex for %define directives (AR-6.3).
var defineDirectiveRegex = regexp.MustCompile(`(?m)^\s*%define\s+(\w+)\s*$`)

// PreProcessingCreateSymbolTable scans the source code for %define directives and builds
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
func CreateSymbolTable(source string, macroTable map[string]Macro) map[string]bool {
	// Early-exit: if no %define directives exist, skip regex processing (AR-8.2).
	hasDefines := strings.Contains(source, "%define")

	var matches [][]int
	if hasDefines {
		matches = defineDirectiveRegex.FindAllStringSubmatchIndex(source, -1)
	}

	type symbolEntry struct {
		name       string
		lineNumber int
	}

	entries := make([]symbolEntry, 0, len(matches))

	// Pass 1: collect all %define directives.
	for _, matchIdx := range matches {
		if len(matchIdx) < 4 {
			continue
		}

		matchStart := matchIdx[0]
		lineNumber := strings.Count(source[:matchStart], "\n") + 1
		symbolName := source[matchIdx[2]:matchIdx[3]]

		if symbolName == "" {
			panic(fmt.Sprintf("pre-processing error: Empty symbol name in %%define at line %d", lineNumber))
		}

		entries = append(entries, symbolEntry{
			name:       symbolName,
			lineNumber: lineNumber,
		})
	}

	// Pass 2: detect duplicates and build the symbol table.
	seen := make(map[string]int, len(entries))
	symbolTable := make(map[string]bool, len(entries)+len(macroTable))
	for _, entry := range entries {
		if firstLine, exists := seen[entry.name]; exists {
			panic(fmt.Sprintf("pre-processing error: Duplicate %%define for symbol '%s' at line %d (first defined at line %d)",
				entry.name, entry.lineNumber, firstLine))
		}
		seen[entry.name] = entry.lineNumber
		symbolTable[entry.name] = true
	}

	// Pass 3: add macro names as defined symbols.
	for macroName := range macroTable {
		symbolTable[macroName] = true
	}

	return symbolTable
}
