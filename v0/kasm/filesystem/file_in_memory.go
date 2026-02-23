package filesystem

import "strings"

type ExpansionInformation struct {
	// _type - the type of expansion that generated this line map (e.g., "macro", "include", etc.).
	_type string
	// macroName - the name of the macro that was expanded to generate this line map.
	macroName string
	// arguments - the arguments passed to the macro during expansion, if any.
	arguments []string
}

type LineMap struct {
	// filename - the original persisted file from which this line map was generated.
	filename string
	// line_start - the line number in the original file where this line map starts.
	lineStart int32
	// lineCount - the number of lines in this expansion.
	lineCount int32
	// expansionInfo - optional information about how this line map was generated, such as the macro name and arguments if it was generated from a macro expansion.
	expansionInfo *ExpansionInformation
}

// Filename - returns the filename associated with this line map.
func (l *LineMap) Filename() string {
	return l.filename
}

// LineStart - returns the line number in the original file where this line map starts.
func (l *LineMap) LineStart() int32 {
	return l.lineStart
}

// LineCount - returns the number of lines in this expansion.
func (l *LineMap) LineCount() int32 {
	return l.lineCount
}

// ExpansionInfo - returns the optional information about how this line map was generated, such as the macro name and arguments if it was generated from a macro expansion.
func (l *LineMap) ExpansionInfo() *ExpansionInformation {
	return l.expansionInfo
}

type FileInMemory struct {
	// Source - file content as it is on persistent storage.
	Source PersistedFile
	// Value - the in-memory representation of the source content, which
	// may be modified during processing (e.g., after pre-processing).
	Value string
	// LineMap - maps line numbers in the source content to line numbers in the value content.
	// used for error reporting and debugging to maintain accurate line numbers after transformations.
	LineMap []LineMap
}

// FileInMemoryNew - creates a new FileInMemory with the given source file and initializes the value and line map.
func FileInMemoryNew(source PersistedFile) (*FileInMemory, error) {
	if !source.Loaded() {
		err := source.Load()
		if err != nil {
			return nil, err
		}
	}

	return &FileInMemory{
		Source:  source,
		Value:   strings.Clone(*source.Content),
		LineMap: make([]LineMap, 0),
	}, nil
}

// RecordInclusion - records an inclusion of another file into the in-memory file, updating the line map
// to reflect the new lines added by the included file.
func (f *FileInMemory) RecordInclusion(includedFile *PersistedFile, lineStart int32) {
	lineMapEntry := LineMap{
		filename:  includedFile.Path,
		lineStart: lineStart,
		lineCount: int32(len(strings.Split(*includedFile.Content, "\n"))), // number of lines in the included file
		expansionInfo: &ExpansionInformation{
			_type:     "include",
			macroName: "include",
			arguments: []string{includedFile.Path},
		},
	}
	f.LineMap = append(f.LineMap, lineMapEntry)

	// Push the content on `lineStart` in p.Value 1 line down to make space for the included file content.
	lines := strings.Split(f.Value, "\n")
	includedLines := strings.Split(*includedFile.Content, "\n")

	insertAt := int(lineStart)
	if insertAt > len(lines) {
		insertAt = len(lines)
	}

	updatedLines := make([]string, 0, len(lines)+len(includedLines))
	updatedLines = append(updatedLines, lines[:insertAt]...)
	updatedLines = append(updatedLines, includedLines...)
	updatedLines = append(updatedLines, lines[insertAt:]...)

	f.Value = strings.Join(updatedLines, "\n")
}

// Origin - returns the (full) origin of a particular line in the FileInMemory value. This includes
// each LineMap entry that was recorded to generate the latest version of the line.
func (f *FileInMemory) Origin(valueLineNumber int32) []LineMap {
	origin := make([]LineMap, 0)

	return origin
}

// GetOriginalLineNumber - returns the original line number in the source file for a given line number in the value content, using the line map to translate between them.
func (f *FileInMemory) GetOriginalLineNumber(valueLineNumber int32) (string, int32) {
	for _, lineMapEntry := range f.LineMap {
		if valueLineNumber >= lineMapEntry.lineStart && valueLineNumber < lineMapEntry.lineStart+lineMapEntry.lineCount {
			originalLineNumber := valueLineNumber - lineMapEntry.lineStart
			return lineMapEntry.filename, originalLineNumber
		}
	}

	// If the line number is not found in any line map entry, it corresponds to the original source file.
	return f.Source.Path, valueLineNumber
}
