package kasm

import "fmt"

// CodegenError represents a single error encountered during code generation.
// It is a plain data struct — not an error interface implementation — so that
// multiple errors can be accumulated and returned as a slice.
type CodegenError struct {
	Message string
	Line    int
	Column  int
}

// String returns a human-readable representation of the code generation error.
func (e CodegenError) String() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}
