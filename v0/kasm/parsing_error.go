package kasm

import "fmt"

// ParseError represents a single error encountered during parsing. It is a
// plain data struct — not an error interface implementation — so that multiple
// errors can be accumulated and returned as a slice.
type ParseError struct {
	Message string
	Line    int
	Column  int
}

// String returns a human-readable representation of the parse error.
func (e ParseError) String() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}
