package preProcessing

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Pre-compiled regex for %include directives (AR-6.3).
var includeDirectiveRegex = regexp.MustCompile(`(?m)^\s*%include\s+"([^"]+)"\s*$`)

// PreProcessingHandleIncludes processes %include directives in the source code,
// replacing each with the content of the referenced file. It returns the
// updated source code and a list of inclusions for error reporting and debugging.
//
// Only .kasm files may be included; any other file extension is a pre-processing error.
//
// The alreadyIncluded set contains file paths that have been inlined by a
// previous invocation (FR-1.7: Shared Dependency Deduplication). When a
// %include directive references a path in this set, the directive line is
// silently removed without reading or inlining the file content again. Pass
// nil if no deduplication is needed.
//
// The function works in three passes:
//  1. Collect all %include directives and their line numbers into the inclusions slice.
//     Directives targeting paths in alreadyIncluded are collected separately for removal.
//     Panics if a non-.kasm file is referenced.
//  2. Deduplicate %include directives within this invocation (FR-1.7). The first
//     occurrence of each path is kept; subsequent duplicates are silently stripped.
//  3. Replace each new %include directive with the content of the referenced file,
//     wrapped in ; FILE: and ; END FILE: comments for traceability. Only the first
//     match is replaced. After inlining, any remaining %include directives for
//     shared or duplicate paths are stripped from the source.
func HandleIncludes(source string, alreadyIncluded map[string]bool) (string, []Inclusion) {
	// Early-exit: if the source is empty, skip all processing (AR-8.1).
	if len(source) == 0 {
		return source, nil
	}

	// Early-exit: if the source does not contain %include, skip all processing (AR-8.2).
	if !strings.Contains(source, "%include") {
		return source, nil
	}

	matches := includeDirectiveRegex.FindAllStringSubmatchIndex(source, -1)

	// Pre-allocate with known capacity to avoid repeated slice growth
	inclusions := make([]Inclusion, 0, len(matches))

	// FR-1.7: Collect paths that should be silently stripped (shared dependencies).
	var sharedPaths []string

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

		// FR-1.7.1 / FR-1.7.4: If the file has already been inlined by a
		// previous invocation, record it for silent removal — do not add it
		// to the inclusions list.
		if alreadyIncluded != nil && alreadyIncluded[includedFilePath] {
			sharedPaths = append(sharedPaths, includedFilePath)
			continue
		}

		inclusions = append(inclusions, Inclusion{
			IncludedFilePath: includedFilePath,
			LineNumber:       lineNumber,
		})
	}

	// Pass 2: deduplicate %include directives within this invocation.
	// FR-1.7: A shared dependency may appear in multiple included files. The
	// first occurrence is kept; subsequent duplicates are silently stripped.
	seen := make(map[string]bool, len(inclusions))
	deduplicated := make([]Inclusion, 0, len(inclusions))
	for _, inclusion := range inclusions {
		if seen[inclusion.IncludedFilePath] {
			// FR-1.7.3: Duplicate within this invocation — will be stripped after inlining.
			continue
		}
		seen[inclusion.IncludedFilePath] = true
		deduplicated = append(deduplicated, inclusion)
	}
	inclusions = deduplicated

	// Pass 3: replace each %include directive with the file content,
	// surrounded by ; FILE: / ; END FILE: comments.
	// Process in reverse source order so that replacements at later positions
	// do not shift earlier positions, and each directive is replaced at its
	// original location rather than at a match injected by a previous inline.
	for idx := len(inclusions) - 1; idx >= 0; idx-- {
		inclusion := inclusions[idx]
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

		// Per-value regex: depends on the inclusion path, compiled once per path (AR-6.4).
		includeDirectivePattern := regexp.MustCompile(`(?m)^\s*%include\s+"` + regexp.QuoteMeta(inclusion.IncludedFilePath) + `"\s*$`)
		// Replace only the first match so that duplicates are not inlined twice.
		loc := includeDirectivePattern.FindStringIndex(source)
		if loc != nil {
			source = source[:loc[0]] + includedContent + source[loc[1]:]
		}
	}

	// FR-1.7.1: Strip remaining %include directives for shared dependencies.
	// This removes: (a) directives for paths already inlined in a previous
	// invocation (from the alreadyIncluded set), and (b) duplicate directives
	// within this invocation whose first occurrence was already inlined above.
	for _, sharedPath := range sharedPaths {
		pattern := regexp.MustCompile(`(?m)^\s*%include\s+"` + regexp.QuoteMeta(sharedPath) + `"\s*\n?`)
		source = pattern.ReplaceAllString(source, "")
	}
	// Also strip any remaining duplicates that were deduplicated in Pass 2.
	for path := range seen {
		pattern := regexp.MustCompile(`(?m)^\s*%include\s+"` + regexp.QuoteMeta(path) + `"\s*\n?`)
		source = pattern.ReplaceAllString(source, "")
	}

	return source, inclusions
}
