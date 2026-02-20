package kasm

import (
	"fmt"
	"strings"
)

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
	directiveRegex := conditionalDirectiveRegex
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
