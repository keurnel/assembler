package kasm

import "github.com/keurnel/assembler/v0/kasm/preProcessing"

// PreProcessingHasMacros is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingHasMacros(source string) bool {
	return preProcessing.PreProcessingHasMacros(source)
}

// PreProcessingMacroTable is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingMacroTable(source string) map[string]Macro {
	return preProcessing.PreProcessingMacroTable(source)
}

// PreProcessingCollectMacroCalls is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingCollectMacroCalls(source string, macroTable map[string]Macro) {
	preProcessing.PreProcessingCollectMacroCalls(source, macroTable)
}

// PreProcessingReplaceMacroCalls is re-exported from the preProcessing sub-package
// for backward compatibility.
func PreProcessingReplaceMacroCalls(source string, macroTable map[string]Macro) string {
	return preProcessing.PreProcessingReplaceMacroCalls(source, macroTable)
}
