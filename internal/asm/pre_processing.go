package asm

// PreProcessingRemoveEmptyLines - Remove all empty lines from the assembly code.
func PreProcessingRemoveEmptyLines(assemblyCode string) string {
	var result string
	lines := splitLines(assemblyCode)
	for _, line := range lines {
		if trim(line) != "" {
			result += line + "\n"
		}
	}
	return result
}

// PreProcessingTrimWhitespace - Trim leading and trailing whitespace from each line of the assembly code.
func PreProcessingTrimWhitespace(assemblyCode string) string {
	var result string
	lines := splitLines(assemblyCode)
	for _, line := range lines {
		result += trim(line) + "\n"
	}
	return result
}

// PreProcessingRemoveComments - Remove all comments from the assembly code. Comments start with `;` and continue to the end of the line.
func PreProcessingRemoveComments(assemblyCode string) string {
	var result string
	lines := splitLines(assemblyCode)
	for _, line := range lines {
		if idx := indexOf(line, ';'); idx != -1 {
			line = line[:idx] // Remove comment part
		}
		result += line + "\n"
	}
	return result
}

func trim(s string) string {
	start := 0
	end := len(s) - 1

	// Trim leading whitespace
	for start <= end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	// Trim trailing whitespace
	for end >= start && (s[end] == ' ' || s[end] == '\t') {
		end--
	}

	if start > end {
		return "" // All whitespace
	}
	return s[start : end+1]
}

func splitLines(s string) []string {
	var lines []string
	currentLine := ""
	for _, char := range s {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func indexOf(s string, char rune) int {
	for i, c := range s {
		if c == char {
			return i
		}
	}
	return -1
}
