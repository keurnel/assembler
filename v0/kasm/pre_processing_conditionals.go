package kasm

import "github.com/keurnel/assembler/v0/kasm/preProcessing"

// PreProcessingHandleConditionals is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingHandleConditionals(source string, definedSymbols map[string]bool) string {
	return preProcessing.PreProcessingHandleConditionals(source, definedSymbols)
}
