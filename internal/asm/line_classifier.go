package asm

import "regexp"

type LineCharacteristics struct {
	// IsDirective - indicates whether the line is a directive line (e.g., starts with a dot and is not a label).
	IsDirective bool
	// IsComment - indicates whether the line is a comment line (e.g., starts with a semicolon).
	IsComment bool
	// ContainsComment - indicates whether the line contains a comment (e.g., has a semicolon anywhere in the line).
	ContainsComment bool
	// IsEmpty - indicates whether the line is empty or contains only whitespace.
	IsEmpty bool
}

// LineAnalyze - analyzes a line of assembly code and returns its characteristics,
// such as whether it's a directive, a comment, contains a comment, or is empty.
func LineAnalyze(line string) LineCharacteristics {
	return LineCharacteristics{
		IsDirective:     IsDirectiveLine(line),
		IsComment:       LineIsComment(line),
		ContainsComment: regexp.MustCompile(`;`).MatchString(line),
		IsEmpty:         regexp.MustCompile(`^\s*$`).MatchString(line),
	}
}

// IsDirectiveLine - verifies if a given line of assembly code qualifies as a directive line.
// An example of a directive line is `.section` making the criteria for a directive line as follows:
// 1. The line must start with a dot ('.') character.
// 2. The line must contain at least one non-whitespace character following the dot.
// 3. Must not end with a colon (':') character, which would indicate a label rather than a directive.
func IsDirectiveLine(line string) bool {
	// Pattern explanation:
	// ^\s*         - Start of the line, followed by optional whitespace.
	// \.           - A literal dot character, indicating the start of a directive.
	// [^\s:]+      - One or more characters that are not whitespace or a colon, representing the directive name.
	// (?::\s|\s.*)? - An optional non-capturing group that allows for either:
	//                 a) A colon followed by optional whitespace (to exclude labels).
	//                 b) Any additional characters after the directive name (e.g., parameters).
	// $             - End of the line.
	matched, _ := regexp.MatchString(`^\s*\.[^\s:]+(?::\s|\s.*)?$`, line)
	return matched
}

// IsNotDirectiveLine - inverse of IsDirectiveLine
func IsNotDirectiveLine(line string) bool {
	return !IsDirectiveLine(line)
}

// LineIsComment - checks if a line is a comment line. A
// comment line is identified by starting with a semicolon (';') character
// possibly preceded by whitespace.
func LineIsComment(line string) bool {
	matched, _ := regexp.MatchString(`^\s*;`, line)
	return matched
}
