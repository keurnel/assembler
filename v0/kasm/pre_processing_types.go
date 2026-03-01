package kasm

import "github.com/keurnel/assembler/v0/kasm/preProcessing"

// Pre-processing types — re-exported from the preProcessing sub-package for
// backward compatibility. All new code should import
// github.com/keurnel/assembler/v0/kasm/preProcessing directly.

type (
	MacroParameter         = preProcessing.MacroParameter
	MacroCall              = preProcessing.MacroCall
	Macro                  = preProcessing.Macro
	PreProcessingInclusion = preProcessing.PreProcessingInclusion
)
