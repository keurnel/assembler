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

// PreProcessingHandleConditionals - processes conditional assembly directives in the source code,
// such as %ifdef, %ifndef, %else, and %endif. It evaluates the conditions based on defined symbols and
// returns the updated source code with the appropriate sections included or excluded.
// PreProcessingHandleConditionals - processes %ifdef, %ifndef, %else, and %endif directives
// in the source code, evaluating conditions based on the provided defined symbols map.
// It returns the updated source code with the appropriate sections included or excluded.
//
// The function works in two passes:
//  1. Validate the structure of all conditional directives, ensuring every %ifdef/%ifndef
//     has a matching %endif, with at most one %else in between. Panics on structural errors.
//  2. Evaluate each conditional block and replace it with the appropriate branch content,
//     or remove it entirely if the condition is not met.
func PreProcessingHandleConditionals(source string, definedSymbols map[string]bool) string {
	// Match %ifdef, %ifndef, %else, and %endif directives
	directiveRegex := regexp.MustCompile(`(?m)^\s*%(ifdef|ifndef|else|endif)\s*(\w*)\s*$`)
	matches := directiveRegex.FindAllStringSubmatchIndex(source, -1)

	type conditionalBlock struct {
		ifDirective string // "ifdef" or "ifndef"
		symbol      string // symbol being tested
		ifStart     int    // byte offset of the start of the %ifdef/%ifndef line
		ifEnd       int    // byte offset of the end of the %ifdef/%ifndef line
		elseStart   int    // byte offset of the start of the %else line (-1 if absent)
		elseEnd     int    // byte offset of the end of the %else line (-1 if absent)
		endifStart  int    // byte offset of the start of the %endif line
		endifEnd    int    // byte offset of the end of the %endif line
		lineNumber  int    // line number of the opening directive (for error reporting)
	}

	// Pass 1: validate and collect all conditional blocks.
	// Uses a stack to match opening directives with their %endif.
	type stackEntry struct {
		directive  string
		symbol     string
		start      int
		end        int
		lineNumber int
		elseStart  int
		elseEnd    int
	}

	stack := make([]stackEntry, 0, len(matches))
	blocks := make([]conditionalBlock, 0, len(matches)/2)

	for _, matchIdx := range matches {
		if len(matchIdx) < 6 {
			continue
		}

		matchStart := matchIdx[0]
		matchEnd := matchIdx[1]
		directive := source[matchIdx[2]:matchIdx[3]]
		symbol := ""
		if matchIdx[4] != matchIdx[5] {
			symbol = source[matchIdx[4]:matchIdx[5]]
		}
		lineNumber := strings.Count(source[:matchStart], "\n") + 1

		switch directive {
		case "ifdef", "ifndef":
			stack = append(stack, stackEntry{
				directive:  directive,
				symbol:     symbol,
				start:      matchStart,
				end:        matchEnd,
				lineNumber: lineNumber,
				elseStart:  -1,
				elseEnd:    -1,
			})

		case "else":
			if len(stack) == 0 {
				panic(fmt.Sprintf("pre-processing error: %%else without matching %%ifdef/%%ifndef at line %d", lineNumber))
			}
			top := &stack[len(stack)-1]
			if top.elseStart != -1 {
				panic(fmt.Sprintf("pre-processing error: Duplicate %%else for %%ifdef/%%ifndef at line %d (first %%else already seen)", lineNumber))
			}
			top.elseStart = matchStart
			top.elseEnd = matchEnd

		case "endif":
			if len(stack) == 0 {
				panic(fmt.Sprintf("pre-processing error: %%endif without matching %%ifdef/%%ifndef at line %d", lineNumber))
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			blocks = append(blocks, conditionalBlock{
				ifDirective: top.directive,
				symbol:      top.symbol,
				ifStart:     top.start,
				ifEnd:       top.end,
				elseStart:   top.elseStart,
				elseEnd:     top.elseEnd,
				endifStart:  matchStart,
				endifEnd:    matchEnd,
				lineNumber:  top.lineNumber,
			})
		}
	}

	// Validate that all opened conditional blocks are closed.
	if len(stack) > 0 {
		top := stack[len(stack)-1]
		panic(fmt.Sprintf("pre-processing error: %%ifdef/%%ifndef at line %d has no matching %%endif", top.lineNumber))
	}

	// Pass 2: evaluate each conditional block and replace with the appropriate branch.
	// Process blocks in reverse order to preserve byte offsets during replacement.
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]

		conditionMet := definedSymbols[block.symbol]
		if block.ifDirective == "ifndef" {
			conditionMet = !conditionMet
		}

		var replacement string
		if block.elseStart == -1 {
			// No %else branch: include body if condition met, otherwise remove the whole block.
			if conditionMet {
				replacement = strings.TrimRight(source[block.ifEnd:block.endifStart], " \t\n\r")
			} else {
				replacement = ""
			}
		} else {
			// Has %else branch: pick the appropriate branch.
			if conditionMet {
				replacement = strings.TrimRight(source[block.ifEnd:block.elseStart], " \t\n\r")
			} else {
				replacement = strings.TrimRight(source[block.elseEnd:block.endifStart], " \t\n\r")
			}
		}

		replacement = strings.TrimSpace(replacement)
		if replacement != "" {
			replacement = "\n" + replacement + "\n"
		}

		source = source[:block.ifStart] + replacement + source[block.endifEnd:]
	}

	return source
}

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
