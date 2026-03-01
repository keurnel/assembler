package kasm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/v0/architecture"
)

// ---------------------------------------------------------------------------
// Internal helper types
// ---------------------------------------------------------------------------

// labelDecl tracks where a label was declared.
type labelDecl struct {
	Name   string
	Line   int
	Column int
}

// namespaceDecl tracks where a namespace was declared.
type namespaceDecl struct {
	Name   string
	Line   int
	Column int
}

// useDecl tracks where a module was imported.
type useDecl struct {
	Name   string
	Line   int
	Column int
}

// ---------------------------------------------------------------------------
// Analyser
// ---------------------------------------------------------------------------

// LineMapper translates a pre-processed line number back to its original line
// number in the source file before any pre-processing transformations were
// applied. Implemented by lineMap.Tracker.
type LineMapper interface {
	// Origin returns the original 1-based line number for the given
	// pre-processed line number, or -1 if the line was inserted during
	// pre-processing (e.g. a ; FILE: boundary comment).
	Origin(lineNumber int) int
}

// Analyser validates a *Program AST against the rules of the .kasm language
// and the target architecture. It detects errors that are syntactically legal
// but semantically invalid. If an Analyser value exists, it is guaranteed to
// hold a valid program reference and initialised internal state.
type Analyser struct {
	program      *Program
	instructions map[string]architecture.Instruction // Upper-case mnemonic keys.
	labels       map[string]labelDecl
	namespaces   map[string]namespaceDecl
	modules      map[string]useDecl
	errors       []SemanticError
	debugCtx     *debugcontext.DebugContext
	lineMapper   LineMapper
}

// AnalyserNew is the sole constructor. It accepts the *Program AST produced by
// Parser.Parse() and an instruction lookup table (upper-case mnemonic keys),
// and returns an *Analyser that is ready for Analyse() to be called.
// AnalyserNew is infallible — it cannot fail. A nil program is treated as empty.
func AnalyserNew(program *Program, instructions map[string]architecture.Instruction) *Analyser {
	if program == nil {
		program = &Program{Statements: make([]Statement, 0)}
	}
	if instructions == nil {
		instructions = make(map[string]architecture.Instruction)
	}
	return &Analyser{
		program:      program,
		instructions: instructions,
		labels:       make(map[string]labelDecl),
		namespaces:   make(map[string]namespaceDecl),
		modules:      make(map[string]useDecl),
		errors:       make([]SemanticError, 0),
	}
}

// WithDebugContext attaches a debug context to the analyser for diagnostic
// recording. When set, the analyser records errors and trace entries into the
// context. When nil, the analyser operates silently using only the internal
// error slice. Returns the analyser for chaining.
func (a *Analyser) WithDebugContext(ctx *debugcontext.DebugContext) *Analyser {
	a.debugCtx = ctx
	return a
}

// WithLineMapper attaches a LineMapper that translates pre-processed line
// numbers back to original source line numbers. When set, every error
// recorded via debugCtx uses the mapped (original) line number instead of
// the pre-processed one. Returns the analyser for chaining.
func (a *Analyser) WithLineMapper(m LineMapper) *Analyser {
	a.lineMapper = m
	return a
}

// ---------------------------------------------------------------------------
// Error recording
// ---------------------------------------------------------------------------

// addError records a semantic error at the given position. If a debug context
// is attached, the error is also recorded there. If a LineMapper is attached,
// the line number is translated to the original source line before recording
// in debugCtx.
func (a *Analyser) addError(message string, line, column int) {
	a.errors = append(a.errors, SemanticError{
		Message: message,
		Line:    line,
		Column:  column,
	})
	if a.debugCtx != nil {
		a.debugCtx.Error(
			a.debugCtx.Loc(a.mapLine(line), column),
			message,
		)
	}
}

// ---------------------------------------------------------------------------
// Analyse (FR-2)
// ---------------------------------------------------------------------------

// Analyse performs semantic analysis on the program AST and returns all
// accumulated semantic errors. It uses two passes: a collection phase
// (labels, namespaces, uses) followed by a validation phase.
func (a *Analyser) Analyse() []SemanticError {
	if a.debugCtx != nil {
		a.debugCtx.SetPhase("semantic-analysis")
	}

	// Pass 1: Collection — gather labels, namespaces, and use declarations.
	a.collect()

	// Pass 2: Validation — validate all statements.
	a.validate()

	if a.debugCtx != nil {
		a.debugCtx.Trace(
			a.debugCtx.Loc(0, 0),
			fmt.Sprintf("semantic analysis complete: %d statement(s), %d error(s)",
				len(a.program.Statements), len(a.errors)),
		)
	}

	return a.errors
}

// ---------------------------------------------------------------------------
// Pass 1: Collection
// ---------------------------------------------------------------------------

// collect gathers all label, namespace, and use declarations into lookup
// tables. Duplicate declarations are recorded as errors immediately.
func (a *Analyser) collect() {
	for _, stmt := range a.program.Statements {
		switch s := stmt.(type) {
		case *LabelStmt:
			a.collectLabel(s)
		case *NamespaceStmt:
			a.collectNamespace(s)
		case *UseStmt:
			a.collectUse(s)
		}
	}
}

// mapLine translates a 1-based pre-processed line number to its 1-based
// original source line number using the attached LineMapper. Returns the
// input unchanged when no mapper is set or the line was inserted during
// pre-processing.
func (a *Analyser) mapLine(line int) int {
	if a.lineMapper == nil {
		return line
	}
	if orig := a.lineMapper.Origin(line - 1); orig >= 0 {
		return orig + 1
	}
	return line
}

// collectLabel adds a label to the label table or records a duplicate error.
func (a *Analyser) collectLabel(s *LabelStmt) {
	if prev, exists := a.labels[s.Name]; exists {
		a.addError(
			fmt.Sprintf("duplicate label '%s', previously declared at %d:%d", s.Name, a.mapLine(prev.Line), prev.Column),
			s.Line, s.Column,
		)
		return
	}
	a.labels[s.Name] = labelDecl{Name: s.Name, Line: s.Line, Column: s.Column}
}

// collectNamespace adds a namespace to the table or records a duplicate error.
func (a *Analyser) collectNamespace(s *NamespaceStmt) {
	if prev, exists := a.namespaces[s.Name]; exists {
		a.addError(
			fmt.Sprintf("duplicate namespace '%s', previously declared at %d:%d", s.Name, a.mapLine(prev.Line), prev.Column),
			s.Line, s.Column,
		)
		return
	}
	a.namespaces[s.Name] = namespaceDecl{Name: s.Name, Line: s.Line, Column: s.Column}
}

// collectUse adds a module import to the table or records a duplicate error.
func (a *Analyser) collectUse(s *UseStmt) {
	if prev, exists := a.modules[s.ModuleName]; exists {
		a.addError(
			fmt.Sprintf("duplicate use of module '%s', previously imported at %d:%d", s.ModuleName, a.mapLine(prev.Line), prev.Column),
			s.Line, s.Column,
		)
		return
	}
	a.modules[s.ModuleName] = useDecl{Name: s.ModuleName, Line: s.Line, Column: s.Column}
}

// ---------------------------------------------------------------------------
// Pass 2: Validation
// ---------------------------------------------------------------------------

// validate walks every statement and performs semantic checks.
func (a *Analyser) validate() {
	for _, stmt := range a.program.Statements {
		switch s := stmt.(type) {
		case *InstructionStmt:
			a.validateInstruction(s)
		case *LabelStmt:
			// Already validated during collection (duplicate check).
		case *NamespaceStmt:
			a.validateNamespace(s)
		case *UseStmt:
			a.validateUse(s)
		case *DirectiveStmt:
			a.validateDirective(s)
		}
	}
}

// ---------------------------------------------------------------------------
// Instruction validation (FR-3)
// ---------------------------------------------------------------------------

// validateInstruction validates mnemonic, operand count, operand types,
// immediate values, and memory operands.
func (a *Analyser) validateInstruction(s *InstructionStmt) {
	upper := strings.ToUpper(s.Mnemonic)

	// FR-3.1: Mnemonic validation
	instr, found := a.instructions[upper]
	if !found {
		a.addError(
			fmt.Sprintf("unknown instruction '%s'", s.Mnemonic),
			s.Line, s.Column,
		)
		// Cannot validate operands without instruction metadata.
		// Still validate individual operands for immediate/memory errors.
		a.validateOperands(s)
		return
	}

	// Validate individual operands (immediates, memory) regardless of variant matching.
	a.validateOperands(s)

	// FR-3.2 / FR-3.3: Operand count and type validation via variants.
	if instr.HasVariants() {
		a.validateVariantMatch(s, &instr)
	}
}

// validateOperands validates each operand in isolation (immediate values,
// memory operands). Also checks identifier references against the label table.
func (a *Analyser) validateOperands(s *InstructionStmt) {
	for _, op := range s.Operands {
		switch o := op.(type) {
		case *ImmediateOperand:
			a.validateImmediate(o)
		case *MemoryOperand:
			a.validateMemoryOperand(o)
		case *IdentifierOperand:
			a.validateIdentifierReference(o)
		}
	}
}

// validateVariantMatch attempts to find a matching instruction variant for the
// supplied operands. If no variant matches, an error is recorded.
func (a *Analyser) validateVariantMatch(s *InstructionStmt, instr *architecture.Instruction) {
	operandTypes := make([]string, len(s.Operands))
	for i, op := range s.Operands {
		operandTypes[i] = operandSemanticType(op)
	}

	// Try exact match first.
	if instr.FindVariant(operandTypes...) != nil {
		return
	}

	// FR-3.3.3: Try with identifier → relative/far substitution.
	if a.tryIdentifierSubstitution(instr, operandTypes) {
		return
	}

	// Determine whether this is a count mismatch or a type mismatch.
	if !a.anyVariantMatchesCount(instr, len(s.Operands)) {
		// FR-3.2.1: No variant has this operand count.
		expected := a.variantOperandCounts(instr)
		a.addError(
			fmt.Sprintf("instruction '%s' expects %s operand(s), got %d",
				s.Mnemonic, expected, len(s.Operands)),
			s.Line, s.Column,
		)
	} else {
		// FR-3.3.2: Count matches some variant, but types don't.
		a.addError(
			fmt.Sprintf("no variant of '%s' accepts operands (%s)",
				s.Mnemonic, strings.Join(operandTypes, ", ")),
			s.Line, s.Column,
		)
	}
}

// operandSemanticType maps an AST Operand node to its semantic type string.
func operandSemanticType(op Operand) string {
	switch op.(type) {
	case *RegisterOperand:
		return "register"
	case *ImmediateOperand:
		return "immediate"
	case *MemoryOperand:
		return "memory"
	case *IdentifierOperand:
		return "identifier"
	case *StringOperand:
		return "string"
	default:
		return "unknown"
	}
}

// tryIdentifierSubstitution tries replacing each "identifier" operand type
// with "relative" and "far" to see if a variant matches.
func (a *Analyser) tryIdentifierSubstitution(instr *architecture.Instruction, types []string) bool {
	// Find indices of identifier operands.
	idxs := make([]int, 0)
	for i, t := range types {
		if t == "identifier" {
			idxs = append(idxs, i)
		}
	}
	if len(idxs) == 0 {
		return false
	}

	// Try each substitution. For simplicity, try "relative" and "far" for each
	// identifier position. This handles the common case of 0–2 identifier operands.
	substitutions := []string{"relative", "far"}
	return a.trySubstitutionRecursive(instr, types, idxs, 0, substitutions)
}

// trySubstitutionRecursive recursively tries all substitution combinations.
func (a *Analyser) trySubstitutionRecursive(instr *architecture.Instruction, types []string, idxs []int, pos int, subs []string) bool {
	if pos >= len(idxs) {
		return instr.FindVariant(types...) != nil
	}
	idx := idxs[pos]
	original := types[idx]
	for _, sub := range subs {
		types[idx] = sub
		if a.trySubstitutionRecursive(instr, types, idxs, pos+1, subs) {
			types[idx] = original // restore
			return true
		}
	}
	types[idx] = original // restore
	return false
}

// anyVariantMatchesCount returns true if any variant has the given operand count.
func (a *Analyser) anyVariantMatchesCount(instr *architecture.Instruction, count int) bool {
	for _, v := range instr.Variants {
		if len(v.Operands) == count {
			return true
		}
	}
	return false
}

// variantOperandCounts returns a human-readable string of unique operand counts
// across all variants (e.g. "1 or 2").
func (a *Analyser) variantOperandCounts(instr *architecture.Instruction) string {
	seen := make(map[int]bool)
	counts := make([]int, 0)
	for _, v := range instr.Variants {
		c := len(v.Operands)
		if !seen[c] {
			seen[c] = true
			counts = append(counts, c)
		}
	}
	parts := make([]string, len(counts))
	for i, c := range counts {
		parts[i] = strconv.Itoa(c)
	}
	return strings.Join(parts, " or ")
}

// ---------------------------------------------------------------------------
// Label reference validation (FR-4.2)
// ---------------------------------------------------------------------------

// validateIdentifierReference checks an identifier operand against the label table.
func (a *Analyser) validateIdentifierReference(o *IdentifierOperand) {
	if _, exists := a.labels[o.Name]; exists {
		return // Resolved to a label.
	}
	// Check if it matches a namespace-qualified pattern (future extension).
	// For now, record an undefined reference error.
	a.addError(
		fmt.Sprintf("undefined reference to '%s'", o.Name),
		o.Line, o.Column,
	)
}

// ---------------------------------------------------------------------------
// Namespace validation (FR-5)
// ---------------------------------------------------------------------------

// validateNamespace performs defence-in-depth checks on a namespace statement.
func (a *Analyser) validateNamespace(s *NamespaceStmt) {
	if s.Name == "" {
		a.addError("namespace name must not be empty", s.Line, s.Column)
		return
	}
	if s.Name[0] >= '0' && s.Name[0] <= '9' {
		a.addError(
			fmt.Sprintf("namespace name '%s' must not start with a digit", s.Name),
			s.Line, s.Column,
		)
	}
}

// ---------------------------------------------------------------------------
// Use validation (FR-6)
// ---------------------------------------------------------------------------

// validateUse performs defence-in-depth checks on a use statement.
func (a *Analyser) validateUse(s *UseStmt) {
	if s.ModuleName == "" {
		a.addError("module name must not be empty", s.Line, s.Column)
	}
}

// ---------------------------------------------------------------------------
// Directive validation (FR-7)
// ---------------------------------------------------------------------------

// validateDirective checks that a surviving directive is recognised.
func (a *Analyser) validateDirective(s *DirectiveStmt) {
	// Currently no post-pre-processing directives are defined.
	// All surviving directives are unrecognised.
	a.addError(
		fmt.Sprintf("unrecognised directive '%s'", s.Literal),
		s.Line, s.Column,
	)
}

// ---------------------------------------------------------------------------
// Immediate value validation (FR-8)
// ---------------------------------------------------------------------------

// validateImmediate checks that an immediate operand is a valid numeric literal.
func (a *Analyser) validateImmediate(o *ImmediateOperand) {
	val := o.Value
	if val == "" {
		a.addError("invalid immediate value ''", o.Line, o.Column)
		return
	}

	// Hex literal: 0x or 0X prefix.
	if len(val) >= 2 && val[0] == '0' && (val[1] == 'x' || val[1] == 'X') {
		hexPart := val[2:]
		if len(hexPart) == 0 {
			a.addError(
				fmt.Sprintf("invalid immediate value '%s'", val),
				o.Line, o.Column,
			)
			return
		}
		for _, ch := range hexPart {
			if !isHexRune(ch) {
				a.addError(
					fmt.Sprintf("invalid immediate value '%s'", val),
					o.Line, o.Column,
				)
				return
			}
		}
		return
	}

	// Decimal literal.
	for _, ch := range val {
		if ch < '0' || ch > '9' {
			a.addError(
				fmt.Sprintf("invalid immediate value '%s'", val),
				o.Line, o.Column,
			)
			return
		}
	}
}

// isHexRune returns true if r is a valid hexadecimal digit.
func isHexRune(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// ---------------------------------------------------------------------------
// Memory operand validation (FR-9)
// ---------------------------------------------------------------------------

// validateMemoryOperand validates the structure of a memory operand.
func (a *Analyser) validateMemoryOperand(o *MemoryOperand) {
	// FR-9.1: Must contain at least one component.
	if len(o.Components) == 0 {
		a.addError("empty memory operand", o.Line, o.Column)
		return
	}

	// FR-9.2: Base must be a register or identifier, not an immediate.
	first := o.Components[0].Token
	if first.Type == TokenImmediate {
		a.addError(
			"memory operand base must be a register or identifier, got immediate",
			o.Line, o.Column,
		)
	}

	// FR-9.4: Validate operators.
	for _, comp := range o.Components {
		tok := comp.Token
		if tok.Type == TokenIdentifier && len(tok.Literal) == 1 {
			ch := tok.Literal[0]
			if ch == '+' || ch == '-' {
				continue // Valid operator.
			}
			if ch == '*' || ch == '/' || ch == '%' || ch == '&' || ch == '|' || ch == '^' {
				a.addError(
					fmt.Sprintf("invalid operator '%s' in memory operand", tok.Literal),
					tok.Line, tok.Column,
				)
			}
		}
	}
}
