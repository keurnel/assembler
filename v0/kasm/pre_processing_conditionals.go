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
	if len(source) == 0 {
		return source
	}

	directiveRegex := conditionalDirectiveRegex
	matches := directiveRegex.FindAllStringSubmatchIndex(source, -1)

	// Fast path: no directives found, return source unchanged.
	if len(matches) == 0 {
		return source
	}

	// Pre-compute a line number lookup table: lineOffsets[i] = byte offset where line i+1 starts.
	// This avoids repeated strings.Count calls (O(n) each) per directive.
	lineNumbers := precomputeLineNumbers(source, matches)

	stack := make([]stackEntry, 0, 4)
	blocks := make([]conditionalBlock, 0, len(matches)/2)

	for mi, matchIdx := range matches {
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
		lineNumber := lineNumbers[mi]

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

	if len(stack) > 0 {
		top := stack[len(stack)-1]
		panic(fmt.Sprintf("pre-processing error: %%ifdef/%%ifndef at line %d has no matching %%endif", top.lineNumber))
	}

	// Pass 2: build result in a single pass using a strings.Builder.
	// Sort blocks by ifStart ascending, then walk source copying gaps and evaluated branches.
	sortBlocksByStart(blocks)

	var sb strings.Builder
	sb.Grow(len(source))
	cursor := 0

	for _, block := range blocks {
		// Copy source between previous block end and this block start
		if block.ifStart > cursor {
			sb.WriteString(source[cursor:block.ifStart])
		}

		conditionMet := definedSymbols[block.symbol]
		if block.ifDirective == "ifndef" {
			conditionMet = !conditionMet
		}

		var branch string
		if block.elseStart == -1 {
			if conditionMet {
				branch = strings.TrimSpace(source[block.ifEnd:block.endifStart])
			}
		} else {
			if conditionMet {
				branch = strings.TrimSpace(source[block.ifEnd:block.elseStart])
			} else {
				branch = strings.TrimSpace(source[block.elseEnd:block.endifStart])
			}
		}

		if branch != "" {
			sb.WriteByte('\n')
			sb.WriteString(branch)
			sb.WriteByte('\n')
		}

		cursor = block.endifEnd
	}

	// Copy remaining source after the last block
	if cursor < len(source) {
		sb.WriteString(source[cursor:])
	}

	return sb.String()
}

// precomputeLineNumbers computes the line number for each match in a single pass over source.
func precomputeLineNumbers(source string, matches [][]int) []int {
	// Collect all match start offsets and their indices
	type offsetIndex struct {
		offset int
		index  int
	}
	items := make([]offsetIndex, 0, len(matches))
	for i, m := range matches {
		if len(m) >= 2 {
			items = append(items, offsetIndex{offset: m[0], index: i})
		}
	}

	// Sort by offset (they should already be in order from regex, but be safe)
	// Already sorted since FindAllStringSubmatchIndex returns in order.

	result := make([]int, len(matches))
	line := 1
	prev := 0
	for _, item := range items {
		// Count newlines between prev and item.offset
		line += strings.Count(source[prev:item.offset], "\n")
		result[item.index] = line
		prev = item.offset
	}
	return result
}

// sortBlocksByStart sorts conditional blocks by ifStart in ascending order (insertion sort,
// typically very few blocks).
func sortBlocksByStart(blocks []conditionalBlock) {
	for i := 1; i < len(blocks); i++ {
		key := blocks[i]
		j := i - 1
		for j >= 0 && blocks[j].ifStart > key.ifStart {
			blocks[j+1] = blocks[j]
			j--
		}
		blocks[j+1] = key
	}
}
