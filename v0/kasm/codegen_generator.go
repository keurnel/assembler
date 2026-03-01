package kasm

import (
	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/kasm/ast"
)

// Generator transforms a validated *ast.Program AST into a flat byte slice of
// machine code for the target architecture. If a Generator value exists, it
// is guaranteed to hold a valid program reference and initialised internal
// state.
type Generator struct {
	program      *ast.Program
	instructions map[string]architecture.Instruction
	labels       map[string]labelEntry
	sections     map[string]*sectionBuffer
	current      string // current section name
	errors       []CodegenError
	debugCtx     *debugcontext.DebugContext
}

// GeneratorNew is the sole constructor. It accepts the validated *ast.Program AST
// and an instruction lookup table (upper-case mnemonic keys), and returns a
// *Generator ready for Generate() to be called. GeneratorNew is infallible —
// it cannot fail. A nil program is treated as empty (FR-1.2).
func GeneratorNew(program *ast.Program, instructions map[string]architecture.Instruction) *Generator {
	if program == nil {
		program = &ast.Program{Statements: make([]ast.Statement, 0)}
	}
	if instructions == nil {
		instructions = make(map[string]architecture.Instruction)
	}
	return &Generator{
		program:      program,
		instructions: instructions,
		labels:       make(map[string]labelEntry),
		sections:     make(map[string]*sectionBuffer),
		current:      "",
		errors:       make([]CodegenError, 0),
	}
}

// WithDebugContext attaches a debug context to the generator for diagnostic
// recording. When set, the generator records errors and trace entries into
// the context. When nil, the generator operates silently using only the
// internal error slice. Returns the generator for chaining (AR-3.2).
func (g *Generator) WithDebugContext(ctx *debugcontext.DebugContext) *Generator {
	g.debugCtx = ctx
	return g
}

// addError records a code generation error at the given position. If a debug
// context is attached, the error is also recorded there. The generator never
// panics (AR-4.2).
func (g *Generator) addError(message string, line, column int) {
	g.errors = append(g.errors, CodegenError{
		Message: message,
		Line:    line,
		Column:  column,
	})
	if g.debugCtx != nil {
		g.debugCtx.Error(
			g.debugCtx.Loc(line, column),
			message,
		)
	}
}
