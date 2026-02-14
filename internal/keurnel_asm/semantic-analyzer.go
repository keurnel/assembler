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

	// Verify that all `use` statements reference valid namespaces
	// use cannot reference a namespace that does not exist in the parsed groups
	// or make a reference to the namespace itself.
	//
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
