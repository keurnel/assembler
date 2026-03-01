package kasm

import (
	"fmt"

	"github.com/keurnel/assembler/v0/kasm/ast"
)

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

// collectPass walks all statements to collect label addresses, section
// boundaries, and compute instruction sizes. No bytes are emitted.
func (g *Generator) collectPass() {
	for _, stmt := range g.program.Statements {
		switch s := stmt.(type) {
		case *ast.SectionStmt:
			g.switchSection(s.Type)

		case *ast.LabelStmt:
			g.ensureSection(s.Line, s.Column)
			g.collectLabel(s)

		case *ast.InstructionStmt:
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

// emitPass walks all statements again and encodes each instruction into bytes
// using the addresses resolved in Pass 1.
func (g *Generator) emitPass() {
	// Reset current section for the second pass.
	g.current = ""

	for _, stmt := range g.program.Statements {
		switch s := stmt.(type) {
		case *ast.SectionStmt:
			g.switchSection(s.Type)

		case *ast.LabelStmt:
			g.ensureSection(s.Line, s.Column)
			// Labels are already collected; nothing to emit.

		case *ast.InstructionStmt:
			g.ensureSection(s.Line, s.Column)
			g.encodeInstruction(s)
		}
	}
}
