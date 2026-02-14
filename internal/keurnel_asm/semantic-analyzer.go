package keurnel_asm

import "errors"

type SemanticAnalyzer struct {
	parser *Parser
}

// SemanticAnalyzerNew - returns a new instance of the SemanticAnalyzer
func SemanticAnalyzerNew(parser *Parser) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		parser: parser,
	}
}

// Analyze - performs semantic analysis on the parsed instruction groups to ensure correctness and resolve namespaces
func (sa *SemanticAnalyzer) Analyze() error {

	// First, we must ensure that there is a `_start:` directive in the global scope, as
	// this is required as the entry-point for the program.
	//
	if _, exists := sa.parser.GetGroup("_start:"); !exists {
		return errors.New("SEMANTIC ERROR: missing `_start:` directive in global scope")
	}

	// Ensure that all `use` statements are valid.
	//
	validUseStatements := sa.validateUseStatements()
	if validUseStatements != nil {
		return validUseStatements
	}

	return nil
}

// validateUseStatements - validates that all `use` statements reference valid namespaces and do not create circular
// references.
func (sa *SemanticAnalyzer) validateUseStatements() error {

	for identifier, group := range sa.parser.Groups() {
		for _, ns := range group.Uses {
			if _, exists := sa.parser.GetGroup(ns); !exists {
				return errors.New("Semantic error: Group '" + identifier + "' references non-existent namespace '" + ns + "' in 'use' statement")
			}
			if ns == identifier {
				return errors.New("Semantic error: Group '" + identifier + "' cannot use itself as a namespace reference in 'use' statement")
			}
		}
	}

	return nil
}
