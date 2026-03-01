package kasm

import "github.com/keurnel/assembler/v0/kasm/preProcessing"

// PreProcessingCreateSymbolTable is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingCreateSymbolTable(source string, macroTable map[string]Macro) map[string]bool {
	return preProcessing.PreProcessingCreateSymbolTable(source, macroTable)
}
