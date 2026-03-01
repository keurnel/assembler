package kasm

import "github.com/keurnel/assembler/v0/kasm/ast"

// AST types — re-exported from the ast sub-package for backward compatibility.
// All new code should import github.com/keurnel/assembler/v0/kasm/ast directly.

type (
	Statement = ast.Statement
	Operand   = ast.Operand
	Program   = ast.Program

	// Statement types.
	InstructionStmt = ast.InstructionStmt
	LabelStmt       = ast.LabelStmt
	NamespaceStmt   = ast.NamespaceStmt
	UseStmt         = ast.UseStmt
	DirectiveStmt   = ast.DirectiveStmt
	SectionStmt     = ast.SectionStmt

	// Operand types.
	RegisterOperand   = ast.RegisterOperand
	ImmediateOperand  = ast.ImmediateOperand
	IdentifierOperand = ast.IdentifierOperand
	StringOperand     = ast.StringOperand
	MemoryComponent   = ast.MemoryComponent
	MemoryOperand     = ast.MemoryOperand
)
