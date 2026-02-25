package kasm

import "fmt"

// SemanticError represents a single error encountered during semantic analysis.
// It is a plain data struct — not an error interface implementation — so that
// multiple errors can be accumulated and returned as a slice.
type SemanticError struct {
	Message string
	Line    int
	Column  int
}

// String returns a human-readable representation of the semantic error.
func (e SemanticError) String() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}
