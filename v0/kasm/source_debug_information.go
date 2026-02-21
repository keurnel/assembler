package kasm

import (
	"context"
	"fmt"
	"sync"
)

type ExpansionEvent struct {
	// LineNumber - the original line number in the source code
	// that is being expanded into multiple lines
	LineNumber int
	// ExpandedLineNumbers - the line numbers in the pre-processed source code that originate
	//from the same line in the original source code
	ExpandedLinesCount int
}

type SourceDebugInformation struct {
	source string

	// Expansion information
	//
	ExpansionChannel chan ExpansionEvent
	lineMapping      map[int][]int
	reverseMapping   map[int]int
}

// CanListen - returns nil if the struct is ready to listen for events, or an error if it is not properly initialized.
func (s *SourceDebugInformation) CanListen() error {
	if s.ExpansionChannel == nil {
		return fmt.Errorf("SourceDebugInformation is not properly initialized: ExpansionChannel is nil")
	}
	return nil
}

// Listen - listens for events on channels of the struct and update
// the internal state accordingly. This must be run in a separate
// goroutine to avoid blocking the main execution flow.
func (s *SourceDebugInformation) Listen(ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	state := "listening"

	// While state is "listening" keep listening for events.
	//
	for state == "listening" {
		select {
		case event := <-s.ExpansionChannel:

			// Print details of event
			//
			fmt.Printf("Received expansion event: LineNumber=%d, ExpandedLinesCount=%d\n", event.LineNumber, event.ExpandedLinesCount)

			s.ExpandLine(event.LineNumber, make([]int, event.ExpandedLinesCount)) // Placeholder for expanded line numbers
		case <-ctx.Done():
			state = "stopped"
		}
	}
}

// Lines - returns the original line numbers corresponding to the given line number in the pre-processed source code.
func (s *SourceDebugInformation) Lines() map[int][]int {
	return s.lineMapping
}

// SourceDebugInformationMake - creates SourceDebugInformation from the given source code
func SourceDebugInformationMake(source string) *SourceDebugInformation {
	instance := &SourceDebugInformation{
		source:           source,
		ExpansionChannel: make(chan ExpansionEvent, 100), // buffered channel to avoid blocking
		lineMapping:      make(map[int][]int),
		reverseMapping:   make(map[int]int),
	}

	if instance.source == "" {
		return instance
	}

	lines := splitIntoLines(instance.source)
	for i := range lines {
		lineNumber := i + 1
		instance.lineMapping[lineNumber] = []int{lineNumber}
		instance.reverseMapping[lineNumber] = lineNumber
	}

	return instance
}

// ExpandLine - expands given line number into multiple line numbers that originate from the same
// line in the original source code. This is used to handle cases in which `%include` or macro
// expansion results in multiple lines of pre-processed source code originating from the same line in the original source code.
func (s *SourceDebugInformation) ExpandLine(lineNumber int, expandedLineNumbers []int) {
	if _, exists := s.lineMapping[lineNumber]; !exists {
		s.lineMapping[lineNumber] = []int{}
	}
	s.lineMapping[lineNumber] = append(s.lineMapping[lineNumber], expandedLineNumbers...)
	for _, expanded := range expandedLineNumbers {
		s.reverseMapping[expanded] = lineNumber
	}
}

// LineNumberToOrigin - returns the original line number where the given (expanded) line number in the pre-processed
// source code originated from. If the given line number does not exist in the mapping, it returns -1.
func (s *SourceDebugInformation) LineNumberToOrigin(lineNumber int) int {
	if origin, exists := s.reverseMapping[lineNumber]; exists {
		return origin
	}
	return -1
}

func splitIntoLines(source string) (lines []string) {
	lines = make([]string, 0, 100) // pre-allocate with an estimated capacity to avoid repeated slice growth
	currentLine := ""
	for _, char := range source {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}
	// Append the last line if it exists (handles case where source does not end with a newline)
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}
