package kasm

import "github.com/keurnel/assembler/v0/kasm/preProcessing"

// PreProcessingHandleIncludes is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingHandleIncludes(source string, alreadyIncluded map[string]bool) (string, []PreProcessingInclusion) {
	return preProcessing.HandleIncludes(source, alreadyIncluded)
}
