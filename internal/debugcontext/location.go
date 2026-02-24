package debugcontext

import "fmt"

// Location identifies a position in source code. It is a value type — safe to
// copy and compare.
//
// See FR-7 in .requirements/assembler/debug-information.md.
type Location struct {
	filePath string // Absolute or relative path to the source file.
	line     int    // 1-based line number.
	column   int    // 1-based column number, or 0 for "entire line".
}

// Loc creates a Location using the provided file path, line, and column.
// Use this when the file path is known directly (e.g. from the context's
// primary file or from an included file).
func Loc(filePath string, line, column int) Location {
	return Location{
		filePath: filePath,
		line:     line,
		column:   column,
	}
}

// FilePath returns the file path of the location.
func (l Location) FilePath() string { return l.filePath }

// Line returns the 1-based line number.
func (l Location) Line() int { return l.line }

// Column returns the 1-based column number, or 0 for "entire line".
func (l Location) Column() int { return l.column }

// String returns a human-readable representation of the location.
// Format: "filePath:line:column" or "filePath:line" if column is 0.
func (l Location) String() string {
	if l.column == 0 {
		return fmt.Sprintf("%s:%d", l.filePath, l.line)
	}
	return fmt.Sprintf("%s:%d:%d", l.filePath, l.line, l.column)
}
