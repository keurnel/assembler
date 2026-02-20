package kasm

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

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
