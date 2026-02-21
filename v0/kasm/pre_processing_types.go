package kasm

// MacroParameter - represents a single parameter of a macro definition.
type MacroParameter struct {
	Name string // Name of the parameter
}

// MacroCall - represents a single invocation of a macro in the source code.
type MacroCall struct {
	Name       string   // Name of the macro being called
	Arguments  []string // Arguments passed to the macro call, in the order they are provided
	LineNumber int      // Line number in the source code where the macro call occurs (for error reporting and debugging)
}

// Macro - represents a macro definition extracted from the source code.
type Macro struct {
	Name       string                    // Name of the macro
	Parameters map[string]MacroParameter // Parameters of the macro, indexed by their names
	Body       string                    // Body of the macro, which may contain the code to be expanded when the macro is invoked
	Calls      []MacroCall               // Calls to this macro found in the source code
}

// PreProcessingInclusion - represents a single %include directive found in the source code.
type PreProcessingInclusion struct {
	IncludedFilePath string // Path of the included file
	LineNumber       int    // Line number in the source code where the inclusion occurs (for error reporting and debugging)
}

// conditionalBlock - represents a single conditional assembly block found in the source code, including its directive
// type, associated symbol, and the positions of its components for error reporting and debugging.
type conditionalBlock struct {
	ifDirective string
	symbol      string
	ifStart     int
	ifEnd       int
	elseStart   int
	elseEnd     int
	endifStart  int
	endifEnd    int
	lineNumber  int
}

// stackEntry - represents an entry in the stack used to track nested conditional directives during pre-processing.
type stackEntry struct {
	directive  string
	symbol     string
	start      int
	end        int
	lineNumber int
	elseStart  int
	elseEnd    int
}
