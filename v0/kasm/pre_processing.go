package kasm

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type MacroParameter struct {
	Name string // Name of the parameter
}

type MacroCall struct {
	Name       string   // Name of the macro being called
	Arguments  []string // Arguments passed to the macro call, in the order they are provided
	LineNumber int      // Line number in the source code where the macro call occurs (for error reporting and debugging)
}

type Macro struct {
	Name       string                    // Name of the macro
	Parameters map[string]MacroParameter // Parameters of the macro, indexed by their names
	Body       string                    // Body of the macro, which may contain the code to be expanded when the macro is invoked
	Calls      []MacroCall
}

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

type PreProcessingInclusion struct {
	IncludedFilePath string // Path of the included file
	LineNumber       int    // Line number in the source code where the inclusion occurs (for error reporting and debugging)
}

// PreProcessingHandleIncludes - processes %include directives in the source code, replacing them with the content of the included files.
// It returns the updated source code and a list of inclusions for error reporting and debugging.
//
// Only .kasm files may be included; any other file extension is a pre-processing error.
//
// The function works in three passes:
//  1. Collect all %include directives and their line numbers into the inclusions slice.
//     Panics if a non-.kasm file is referenced.
//  2. Detect duplicate %include directives and panic if any are found.
//  3. Replace each %include directive with the content of the referenced file,
//     wrapped in ; FILE: and ; END FILE: comments for traceability.
func PreProcessingHandleIncludes(source string) (string, []PreProcessingInclusion) {
	// Match lines of the form: %include "path/to/file"
	includeRegex := regexp.MustCompile(`(?m)^\s*%include\s+"([^"]+)"\s*$`)
	matches := includeRegex.FindAllStringSubmatchIndex(source, -1)

	// Pre-allocate with known capacity to avoid repeated slice growth
	inclusions := make([]PreProcessingInclusion, 0, len(matches))

	// Pass 1: collect all %include directives before modifying the source,
	// so that line numbers remain accurate.
	// Only .kasm files are accepted; including any other file type is a pre-processing error.
	for _, matchIdx := range matches {
		if len(matchIdx) < 4 {
			continue
		}

		matchStart := matchIdx[0]
		lineNumber := strings.Count(source[:matchStart], "\n") + 1

		includedFilePath := source[matchIdx[2]:matchIdx[3]]

		// Validate that the included file has the .kasm extension
		if !strings.HasSuffix(includedFilePath, ".kasm") {
			message := fmt.Sprintf("pre-processing error: Included file '%s' at line %d must have a .kasm extension",
				includedFilePath, lineNumber)
			panic(message)
		}

		inclusions = append(inclusions, PreProcessingInclusion{
			IncludedFilePath: includedFilePath,
			LineNumber:       lineNumber,
		})
	}

	// Pass 2: detect duplicate %include directives.
	// A file may only be included once; duplicates are a pre-processing error.
	seen := make(map[string]int, len(inclusions)) // maps file path to first line number
	for _, inclusion := range inclusions {
		if firstLine, exists := seen[inclusion.IncludedFilePath]; exists {
			message := fmt.Sprintf("pre-processing error: Duplicate %%include for file '%s' at line %d (first included at line %d)",
				inclusion.IncludedFilePath, inclusion.LineNumber, firstLine)
			panic(message)
		}
		seen[inclusion.IncludedFilePath] = inclusion.LineNumber
	}

	// Pass 3: replace each %include directive with the file content,
	// surrounded by ; FILE: / ; END FILE: comments.
	// Compile the replacement pattern once per inclusion path.
	for _, inclusion := range inclusions {
		includedContentBytes, err := os.ReadFile(inclusion.IncludedFilePath)
		if err != nil {
			message := fmt.Sprintf("pre-processing error: Failed to read included file '%s' at line %d: %v",
				inclusion.IncludedFilePath, inclusion.LineNumber, err)
			panic(message)
		}

		// Wrap the file content with file boundary comments for traceability
		includedContent := fmt.Sprintf("; FILE: %s\n%s\n; END FILE: %s\n",
			inclusion.IncludedFilePath,
			strings.TrimSpace(string(includedContentBytes)),
			inclusion.IncludedFilePath,
		)

		includeDirectivePattern := regexp.MustCompile(`(?m)^\s*%include\s+"` + regexp.QuoteMeta(inclusion.IncludedFilePath) + `"\s*$`)
		source = includeDirectivePattern.ReplaceAllString(source, includedContent)
	}

	return source, inclusions
}
