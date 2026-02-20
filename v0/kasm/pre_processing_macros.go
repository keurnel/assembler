package kasm

import (
	"fmt"
	"regexp"
	"strings"
)

// PreProcessingHasMacros - checks if the given source code contains any macros. Returns
// `true` if macros are found, otherwise `false`.
func PreProcessingHasMacros(source string) bool {
	matched, err := regexp.MatchString(`%macro\s+\w+\s*\d*`, source)
	if err != nil {
		return false
	}
	return matched
}

// PreProcessingMacroTable - extracts macro definitions from the source code and returns a map of
// macro structs indexed by their names.
func PreProcessingMacroTable(source string) map[string]Macro {
	macroTable := make(map[string]Macro)
	if !PreProcessingHasMacros(source) {
		return macroTable
	}

	// Find all lines that define macros using a regular expression
	//
	macroDefRegex := regexp.MustCompile(`(?m)^\s*%macro\s+(\w+)\s*(\d*)\s*$`)
	matches := macroDefRegex.FindAllStringSubmatch(source, -1)

	for _, match := range matches {
		// Select until the next %endmacro as the body of the macro
		macroName := match[1]
		paramCount := 0
		if len(match) > 2 && match[2] != "" {
			paramCount = int(match[2][0] - '0') // Convert the parameter count from string to integer
		}

		// Extract the body of the macro
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
			paramName := "param" + string('A'+i-1) // Generate parameter names like paramA, paramB, etc.
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

// PreProcessingColectMacroCalls - collects macro calls from the source code and updates the provided macro table with the calls.
func PreProcessingColectMacroCalls(source string, macroTable map[string]Macro) {
	for macroName, macro := range macroTable {
		// Build a regex pattern that matches: macroName arg1, arg2, ...
		// e.g. "een_macro 1, 2" or "een_macro rax, rbx"
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

// PreProcessingReplaceMacroCalls - replaces macro calls in the source code with their expanded bodies based on the provided macro table.
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

			// Replace the macro call in the source code with the expanded body
			callPattern := `(?m)^[^\S\n]*` + regexp.QuoteMeta(call.Name) + `[^\S\n]+` + regexp.QuoteMeta(strings.Join(call.Arguments, ", ")) + `[^\S\n]*$`
			source = regexp.MustCompile(callPattern).ReplaceAllString(source, expandedBody)
		}
	}

	return source
}
