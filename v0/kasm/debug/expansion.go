package debug

type LineExpansion struct {
	// ConcreteLineNumber - anchor line number in the original source code that
	// is being expanded while pre-processing the source. The concrete line number
	//
	//
	ConcreteLineNumber int
}
