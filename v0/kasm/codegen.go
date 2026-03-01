package kasm

import (
	"fmt"

	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/v0/architecture"
)

// ---------------------------------------------------------------------------
// CodegenError (AR-4.1)
// ---------------------------------------------------------------------------

// CodegenError represents a single error encountered during code generation.
// It is a plain data struct — not an error interface implementation — so that
// multiple errors can be accumulated and returned as a slice.
type CodegenError struct {
	Message string
	Line    int
	Column  int
}

// String returns a human-readable representation of the code generation error.
func (e CodegenError) String() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}

// ---------------------------------------------------------------------------
// Generator (FR-1)
// ---------------------------------------------------------------------------

// Generator transforms a validated *Program AST into a flat byte slice of
// machine code for the target architecture. If a Generator value exists, it
// is guaranteed to hold a valid program reference and initialised internal
// state.
type Generator struct {
	program      *Program
	instructions map[string]architecture.Instruction
	labels       map[string]labelEntry
	sections     map[string]*sectionBuffer
	current      string // current section name
	errors       []CodegenError
	debugCtx     *debugcontext.DebugContext
}

// GeneratorNew is the sole constructor. It accepts the validated *Program AST
// and an instruction lookup table (upper-case mnemonic keys), and returns a
// *Generator ready for Generate() to be called. GeneratorNew is infallible —
// it cannot fail. A nil program is treated as empty (FR-1.2).
func GeneratorNew(program *Program, instructions map[string]architecture.Instruction) *Generator {
	if program == nil {
		program = &Program{Statements: make([]Statement, 0)}
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

// ---------------------------------------------------------------------------
// Error recording (AR-4.2)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Generate (FR-2)
// ---------------------------------------------------------------------------

// Generate performs code generation using a two-pass strategy and returns the
// machine code as a byte slice together with all errors accumulated during
// both passes (FR-2.1).
func (g *Generator) Generate() ([]byte, []CodegenError) {
	// FR-8.1: Set the debug context phase.
	if g.debugCtx != nil {
		g.debugCtx.SetPhase("codegen")
	}

	// Pass 1 — collection: gather labels, sections, and compute instruction
	// sizes. No bytes are emitted (FR-2.1).
	g.collectPass()

	// Pass 2 — emission: encode each instruction into bytes using the
	// addresses resolved in Pass 1 (FR-2.1).
	g.emitPass()

	// Assemble the final binary from all section buffers (FR-3.3).
	output := g.assemble()

	// FR-8.3: Trace summary.
	if g.debugCtx != nil {
		g.debugCtx.Trace(
			g.debugCtx.Loc(0, 0),
			fmt.Sprintf("code generation complete: %d byte(s) emitted across %d section(s)",
				len(output), g.sectionCount()),
		)
	}

	return output, g.errors
}

// ---------------------------------------------------------------------------
// Pass 1: Collection (FR-2.1, FR-4.1)
// ---------------------------------------------------------------------------

// collectPass walks all statements to collect label addresses, section
// boundaries, and compute instruction sizes. No bytes are emitted.
func (g *Generator) collectPass() {
	for _, stmt := range g.program.Statements {
		switch s := stmt.(type) {
		case *SectionStmt:
			g.switchSection(s.Type)

		case *LabelStmt:
			g.ensureSection(s.Line, s.Column)
			g.collectLabel(s)

		case *InstructionStmt:
			g.ensureSection(s.Line, s.Column)
			size := g.computeInstructionSize(s)
			sec := g.currentSection()
			if sec != nil {
				sec.size += size
			}
		}
	}

	// Reset section offsets for Pass 2.
	for _, sec := range g.sections {
		sec.size = 0
	}
}

// ---------------------------------------------------------------------------
// Pass 2: Emission (FR-2.1, FR-5)
// ---------------------------------------------------------------------------

// emitPass walks all statements again and encodes each instruction into bytes
// using the addresses resolved in Pass 1.
func (g *Generator) emitPass() {
	// Reset current section for the second pass.
	g.current = ""

	for _, stmt := range g.program.Statements {
		switch s := stmt.(type) {
		case *SectionStmt:
			g.switchSection(s.Type)

		case *LabelStmt:
			g.ensureSection(s.Line, s.Column)
			// Labels are already collected; nothing to emit.

		case *InstructionStmt:
			g.ensureSection(s.Line, s.Column)
			g.encodeInstruction(s)
		}
	}
}
