package kasm

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regexes for macro processing (AR-6.3).
var (
	// macroDetectRegex checks if any %macro directive exists in the source.
	macroDetectRegex = regexp.MustCompile(`%macro\s+\w+\s*\d*`)
	// macroDefRegex matches a %macro definition line and captures name + param count.
	macroDefRegex = regexp.MustCompile(`(?m)^\s*%macro\s+(\w+)\s*(\d*)\s*$`)
)

// PreProcessingHasMacros returns true if the source contains at least one
// %macro directive, false otherwise. Used as an early-exit check.
func PreProcessingHasMacros(source string) bool {
	return macroDetectRegex.MatchString(source)
}

// PreProcessingMacroTable extracts macro definitions from the source code and
// returns a map of Macro structs indexed by their names. Returns an empty map
// if no macros are found.
func PreProcessingMacroTable(source string) map[string]Macro {
	macroTable := make(map[string]Macro)
	if !PreProcessingHasMacros(source) {
		return macroTable
	}

	matches := macroDefRegex.FindAllStringSubmatch(source, -1)

	for _, match := range matches {
		// Select until the next %endmacro as the body of the macro
		macroName := match[1]
		paramCount := 0
		if len(match) > 2 && match[2] != "" {
			paramCount = int(match[2][0] - '0') // Convert the parameter count from string to integer
		}

		// Per-value regex: depends on the macro name, compiled once per macro (AR-6.4).
		bodyRegex := regexp.MustCompile(`(?s)%macro\s+` + regexp.QuoteMeta(macroName) + `\s*\d*\s*(.*?)%endmacro`)
		bodyMatch := bodyRegex.FindStringSubmatch(source)
		macroBody := ""
		if len(bodyMatch) > 1 {
			macroBody = bodyMatch[1]
		}

		// Extract parameters from the macro body
		//
		parameters := make(map[string]MacroParameter)
		for i := 1; i <= paramCount; i++ {
			paramName := fmt.Sprintf("param%c", 'A'+i-1) // Generate parameter names like paramA, paramB, etc.
			parameters[paramName] = MacroParameter{
				Name: paramName,
			}
		}

		macroTable[macroName] = Macro{
			Name:       macroName,
			Parameters: parameters,
			Body:       macroBody,
			Calls:      []MacroCall{}, // Macro calls can be extracted in a separate pass if needed
		}
	}

	return macroTable
}

// PreProcessingCollectMacroCalls scans the source for invocations of each macro
// in the provided table and appends found calls to Macro.Calls.
// This function mutates macroTable in place — the caller's map is updated directly.
func PreProcessingCollectMacroCalls(source string, macroTable map[string]Macro) {
	for macroName, macro := range macroTable {
		// Per-value regex: depends on the macro name, compiled once per macro (AR-6.4).
		pattern := `(?m)^[^\S\n]*` + regexp.QuoteMeta(macroName) + `\s+(.+)$`
		re := regexp.MustCompile(pattern)

		matches := re.FindAllStringSubmatchIndex(source, -1)
		for _, matchIdx := range matches {
			if len(matchIdx) < 4 {
				continue
			}

			// Determine line number by counting newlines before the match start
			matchStart := matchIdx[0]
			lineNumber := strings.Count(source[:matchStart], "\n") + 1

			// Extract the arguments substring using the capture group indices
			argStr := source[matchIdx[2]:matchIdx[3]]

			// Split arguments by comma and trim whitespace
			rawArgs := strings.Split(argStr, ",")
			args := make([]string, 0, len(rawArgs))
			for _, arg := range rawArgs {
				trimmed := strings.TrimSpace(arg)
				if trimmed != "" {
					args = append(args, trimmed)
				}
			}

			// When arguments does not match the expected parameter count,
			// we crash the assembler with an error message indicating the
			// preprocessing error and the line number where the error occurred.
			if len(args) != len(macro.Parameters) {

				message := fmt.Sprintf("pre-processing error: Macro '%s' expects %d arguments, but got %d at line %d",
					macroName, len(macro.Parameters), len(args), lineNumber)

				panic(message)
			}

			macro.Calls = append(macro.Calls, MacroCall{
				Name:       macroName,
				Arguments:  args,
				LineNumber: lineNumber,
			})
		}

		macroTable[macroName] = macro
	}
}

// PreProcessingReplaceMacroCalls replaces macro invocations in the source code
// with their expanded bodies based on the provided macro table. Placeholders
// (%1, %2, …) are substituted with the call's arguments. Returns the
// transformed source string.
func PreProcessingReplaceMacroCalls(source string, macroTable map[string]Macro) string {
	for _, macro := range macroTable {
		for _, call := range macro.Calls {
			expandedBody := macro.Body

			// Replace %1, %2, ... with the actual arguments from the call
			for i, arg := range call.Arguments {
				placeholder := fmt.Sprintf("%%%d", i+1)
				expandedBody = strings.ReplaceAll(expandedBody, placeholder, arg)
			}

			// Strip leading horizontal whitespace from each line
			lines := strings.Split(expandedBody, "\n")
			trimmedLines := make([]string, 0, len(lines))
			for _, line := range lines {
				trimmed := strings.TrimLeft(line, " \t")
				if trimmed != "" {
					trimmedLines = append(trimmedLines, trimmed)
				}
			}
			expandedBody = strings.Join(trimmedLines, "\n")

			// Prepend a comment indicating the macro name, surrounded by blank lines
			expandedBody = fmt.Sprintf("\n; MACRO: %s\n%s\n", call.Name, expandedBody)

			// Per-value regex: depends on call name + arguments, compiled once per call (AR-6.4).
			callPattern := `(?m)^[^\S\n]*` + regexp.QuoteMeta(call.Name) + `[^\S\n]+` + regexp.QuoteMeta(strings.Join(call.Arguments, ", ")) + `[^\S\n]*$`
			source = regexp.MustCompile(callPattern).ReplaceAllString(source, expandedBody)
		}
	}

	return source
}
