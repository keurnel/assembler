package preProcessing

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regex for conditional directives: %ifdef, %ifndef, %else, %endif (AR-6.1).
var conditionalDirectiveRegex = regexp.MustCompile(`(?m)^\s*%(ifdef|ifndef|else|endif)\s*(\w*)\s*$`)

// Pre-compiled regex for %define directives used for stripping (FR-3.4, AR-6.3).
var defineStripRegex = regexp.MustCompile(`(?m)^\s*%define\s+\w+\s*\n?`)

// HandleConditionals evaluates conditional assembly blocks
// (%ifdef, %ifndef, %else, %endif) and produces a source string with only
// the active branches retained. Directive lines are removed from the output.
func HandleConditionals(source string, definedSymbols map[string]bool) string {

	// When the source is empty, there is nothing
	// to process, so we can return it immediately.
	//
	if len(source) == 0 {
		return source
	}

	// Quick check to skip regex processing if there
	// are no conditional directives in the source code.
	//
	hasConditionals := strings.Contains(source, "%ifdef") ||
		strings.Contains(source, "%ifndef") ||
		strings.Contains(source, "%endif")

	if !hasConditionals {
		// FR-3.4: Even without conditionals, strip %define lines so they
		// do not leak into the lexer.
		return stripDefineDirectives(source)
	}

	directiveRegex := conditionalDirectiveRegex
	matches := directiveRegex.FindAllStringSubmatchIndex(source, -1)

	// Uncommon case: if there are no matches, we can skip all the processing
	// and return the original source immediately.
	//
	if len(matches) == 0 {
		return source
	}

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

	sortBlocksByStart(blocks)

	var sb strings.Builder
	sb.Grow(len(source))
	cursor := 0

	for _, block := range blocks {
		if block.ifStart > cursor {
			sb.WriteString(source[cursor:block.ifStart])
		}

		conditionMet := definedSymbols[block.symbol]
		if block.ifDirective == "ifndef" {
			conditionMet = !conditionMet
		}

		var branchStart, branchEnd int
		hasBranch := false
		if block.elseStart == -1 {
			if conditionMet {
				branchStart = block.ifEnd
				branchEnd = block.endifStart
				hasBranch = true
			}
		} else {
			if conditionMet {
				branchStart = block.ifEnd
				branchEnd = block.elseStart
			} else {
				branchStart = block.elseEnd
				branchEnd = block.endifStart
			}
			hasBranch = true
		}

		if hasBranch {
			// Inline trim without allocation: find first/last non-whitespace byte offsets
			s, e := trimSpaceBounds(source, branchStart, branchEnd)
			if s < e {
				sb.WriteByte('\n')
				sb.WriteString(source[s:e])
				sb.WriteByte('\n')
			}
		}

		cursor = block.endifEnd
	}

	if cursor < len(source) {
		sb.WriteString(source[cursor:])
	}

	// FR-3.4: Strip %define directives from the output so they do not leak
	// into the lexer.
	return stripDefineDirectives(sb.String())
}

// stripDefineDirectives removes all %define directive lines from the source
// so that they do not leak into the lexer (FR-3.4). Uses the early-exit
// pattern (AR-8.2, AR-8.3) to skip processing when no %define exists.
func stripDefineDirectives(source string) string {
	if !strings.Contains(source, "%define") {
		return source
	}
	return defineStripRegex.ReplaceAllString(source, "")
}

// trimSpaceBounds returns the start and end indices within source[start:end]
// (as absolute offsets) with leading/trailing whitespace removed, without allocating.
func trimSpaceBounds(source string, start, end int) (int, int) {
	for start < end && (source[start] == ' ' || source[start] == '\t' || source[start] == '\n' || source[start] == '\r') {
		start++
	}
	for end > start && (source[end-1] == ' ' || source[end-1] == '\t' || source[end-1] == '\n' || source[end-1] == '\r') {
		end--
	}
	return start, end
}

// precomputeLineNumbers computes the line number for each match in a single pass.
// Uses direct byte scanning instead of strings.Count to avoid sub-slice overhead.
func precomputeLineNumbers(source string, matches [][]int) []int {
	result := make([]int, len(matches))
	line := 1
	prev := 0
	for i, m := range matches {
		if len(m) < 2 {
			continue
		}
		offset := m[0]
		// Count newlines from prev to offset by scanning bytes directly
		for j := prev; j < offset; j++ {
			if source[j] == '\n' {
				line++
			}
		}
		result[i] = line
		prev = offset
	}
	return result
}

// sortBlocksByStart sorts conditional blocks by their ifStart offset using
// insertion sort. The number of blocks is typically small (< 20), making
// insertion sort efficient and allocation-free.
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
