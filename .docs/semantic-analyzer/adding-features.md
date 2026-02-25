# Adding Features to the Semantic Analyser

This guide explains how to introduce new validation rules, statement kinds,
and operand kinds to the semantic analyser. It assumes familiarity with the
`.kasm` assembler pipeline and the Go source files in `v0/kasm/`.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture Refresher](#architecture-refresher)
3. [Adding a New Validation Rule](#adding-a-new-validation-rule)
4. [Adding a New Statement Kind](#adding-a-new-statement-kind)
5. [Adding a New Operand Kind](#adding-a-new-operand-kind)
6. [Adding a Recognised Directive](#adding-a-recognised-directive)
7. [Extending the Instruction Table](#extending-the-instruction-table)
8. [Testing](#testing)
9. [Checklist](#checklist)

---

## Overview

The semantic analyser (`v0/kasm/semantic.go`) sits between the parser and
code generation. It receives a `*Program` AST and an instruction table
(`map[string]architecture.Instruction`), performs two passes over the
statements (collection then validation), and returns a `[]SemanticError`
slice.

Key design constraints:

- The analyser is **read-only** — it never modifies the AST.
- It is **architecture-agnostic** — instruction metadata is injected, not
  imported from a specific architecture package.
- It reports **all** errors in a single run (no early abort).
- It may be called **only once** per `Analyser` instance.

---

## Architecture Refresher

```
v0/kasm/
├── ast.go                  # Statement and Operand interfaces + concrete types
├── semantic.go             # Analyser struct, constructor, Analyse(), all validation methods
├── semantic_error.go       # SemanticError data type
├── semantic_test.go        # Tests (kasm_test package)

v0/architecture/
├── instruction.go          # Instruction struct, HasVariants(), FindVariant()
├── instruction_variant.go  # InstructionVariant struct
├── instruction_group.go    # InstructionGroup (used by orchestrator to build the table)

cmd/cli/cmd/x86_64/
├── assemble_file.go        # Orchestrator — builds instruction table, wires analyser
```

The analyser's two-pass structure:

| Pass       | Purpose                                               | Method      |
|------------|-------------------------------------------------------|-------------|
| Collection | Gather labels, namespaces, and `use` declarations     | `collect()` |
| Validation | Validate every statement against collected data + arch | `validate()`|

---

## Adding a New Validation Rule

A validation rule is a method on `*Analyser` that inspects one aspect of a
statement or operand and calls `addError()` when the check fails.

### Step 1 — Write the validation method

Add a private method to `semantic.go`. Follow the naming convention
`validate<Thing>`:

```go
// validateRegisterSize checks that register operands use a register whose
// bit-width matches the instruction's operand-size variant.
func (a *Analyser) validateRegisterSize(s *InstructionStmt, v *architecture.InstructionVariant) {
    // ... validation logic ...
    if mismatch {
        a.addError(
            fmt.Sprintf("register '%s' is %d-bit, expected %d-bit", reg.Name, actual, expected),
            reg.Line, reg.Column,
        )
    }
}
```

Rules for the method:

- Accept the narrowest AST node it needs (e.g. `*InstructionStmt`,
  `*ImmediateOperand`).
- Use `a.addError(message, line, column)` to record errors — this ensures
  the error is appended to the internal slice **and** forwarded to the
  debug context when one is attached.
- Never modify the AST node.
- Never `panic` or `log.Fatal`.

### Step 2 — Call it from the appropriate handler

Validation methods are invoked from the existing per-statement-type handlers
in the `validate()` pass. Find the switch case that matches the statement
kind and insert your call:

```go
func (a *Analyser) validate() {
    for _, stmt := range a.program.Statements {
        switch s := stmt.(type) {
        case *InstructionStmt:
            a.validateInstruction(s)
            // existing calls are inside validateInstruction; add yours there
            // or at the top level if it is independent of instruction metadata.
        // ...
        }
    }
}
```

If the new rule applies inside `validateInstruction`, add the call after the
existing variant-matching logic (or wherever it logically fits):

```go
func (a *Analyser) validateInstruction(s *InstructionStmt) {
    // ... existing mnemonic lookup, operand validation, variant matching ...

    // New: check register sizes against the matched variant.
    if matchedVariant != nil {
        a.validateRegisterSize(s, matchedVariant)
    }
}
```

### Step 3 — Write tests

See [Testing](#testing) below. Every rule needs at least:

- A happy-path test (valid input → zero errors).
- An error-case test (invalid input → expected error message substring).
- An edge-case test (boundary conditions).

### Step 4 — Update the validation summary

Add a row to the **Validation Summary** table in
`.requirements/semantics/requirements.md` (after getting approval) and
document the rule as a new FR item.

---

## Adding a New Statement Kind

When a new language construct is added to the parser (e.g. `SectionStmt`),
the analyser must handle it.

### Step 1 — Define the AST node

In `ast.go`, add the new struct implementing the `Statement` interface:

```go
type SectionStmt struct {
    Name   string
    Line   int
    Column int
}

func (s *SectionStmt) statementNode()       {}
func (s *SectionStmt) StatementLine() int   { return s.Line }
func (s *SectionStmt) StatementColumn() int { return s.Column }
```

### Step 2 — Add a case to the collection pass (if needed)

If the new statement introduces a declaration that must be visible for
forward-reference resolution, add a case to `collect()`:

```go
func (a *Analyser) collect() {
    for _, stmt := range a.program.Statements {
        switch s := stmt.(type) {
        // ...existing cases...
        case *SectionStmt:
            a.collectSection(s)
        }
    }
}
```

You will also need to:

1. Define an internal helper type (e.g. `sectionDecl`) with `Name`,
   `Line`, `Column` fields.
2. Add a map field to the `Analyser` struct (e.g.
   `sections map[string]sectionDecl`).
3. Initialise the map in `AnalyserNew`.

### Step 3 — Add a case to the validation pass

In `validate()`, add a case for the new statement kind:

```go
func (a *Analyser) validate() {
    for _, stmt := range a.program.Statements {
        switch s := stmt.(type) {
        // ...existing cases...
        case *SectionStmt:
            a.validateSection(s)
        }
    }
}
```

Write the `validateSection` method following the pattern described in
[Adding a New Validation Rule](#adding-a-new-validation-rule).

### Step 4 — Tests and documentation

Write tests and add requirements entries (see [Checklist](#checklist)).

---

## Adding a New Operand Kind

When a new operand type is added to the parser (e.g. `ExpressionOperand`),
two things must be updated in the analyser.

### Step 1 — Define the AST node

In `ast.go`, add the new struct implementing the `Operand` interface:

```go
type ExpressionOperand struct {
    Expression string
    Line       int
    Column     int
}

func (o *ExpressionOperand) operandNode()       {}
func (o *ExpressionOperand) OperandLine() int   { return o.Line }
func (o *ExpressionOperand) OperandColumn() int { return o.Column }
```

### Step 2 — Update the operand-type mapping

In `semantic.go`, update `operandSemanticType` to return the correct string
for the new operand kind:

```go
func operandSemanticType(op Operand) string {
    switch op.(type) {
    // ...existing cases...
    case *ExpressionOperand:
        return "expression"
    default:
        return "unknown"
    }
}
```

This string must match the operand type strings used in
`InstructionVariant.Operands` so that `FindVariant` can match them.

### Step 3 — Add per-operand validation (if needed)

If the new operand kind requires its own validation (like `ImmediateOperand`
requires numeric parsing), add a case to `validateOperands`:

```go
func (a *Analyser) validateOperands(s *InstructionStmt) {
    for _, op := range s.Operands {
        switch o := op.(type) {
        // ...existing cases...
        case *ExpressionOperand:
            a.validateExpression(o)
        }
    }
}
```

### Step 4 — Consider identifier substitution

If the new operand kind should be treated as compatible with other variant
operand types (the way `identifier` is compatible with `relative` and
`far`), update `tryIdentifierSubstitution` or create an analogous method.

---

## Adding a Recognised Directive

Currently all directives that survive pre-processing produce an
"unrecognised directive" error. To recognise a new post-pre-processing
directive (e.g. `%section`, `%align`):

### Step 1 — Define a recognised-directives set

In `semantic.go`, add a package-level set (or add to an existing one):

```go
var recognisedDirectives = map[string]bool{
    "%section": true,
    "%align":   true,
}
```

### Step 2 — Update `validateDirective`

```go
func (a *Analyser) validateDirective(s *DirectiveStmt) {
    if recognisedDirectives[s.Literal] {
        a.validateDirectiveArgs(s)  // new method for argument validation
        return
    }
    a.addError(
        fmt.Sprintf("unrecognised directive '%s'", s.Literal),
        s.Line, s.Column,
    )
}
```

### Step 3 — Write argument validation

Each recognised directive may have its own argument rules. Write a dedicated
method (e.g. `validateSectionDirective`, `validateAlignDirective`) and
dispatch from `validateDirectiveArgs`.

---

## Extending the Instruction Table

The analyser receives its instruction table from the orchestrator. To add
new instructions:

1. **Define the instruction** in the appropriate architecture file under
   `v0/architecture/x86/_64/` (e.g. `instructions_data_transfer.go`).
   Each instruction needs a `Mnemonic`, optional `Variants` with
   `Operands` type strings, an `Opcode`, etc.

2. **Ensure the orchestrator picks it up.** The `buildInstructionTable()`
   function in `cmd/cli/cmd/x86_64/assemble_file.go` iterates over all
   groups returned by `_64.Instructions()`. If your new instruction is in
   an existing group, no orchestrator change is needed.

3. **Add it to the lexer profile** if the mnemonic is new. The
   `ArchitectureProfile` in `v0/kasm/profile/` tells the lexer which
   identifiers are instruction mnemonics. Without this, the token will be
   classified as `TokenIdentifier` instead of `TokenInstruction`, and the
   parser will not produce an `InstructionStmt`.

The analyser itself requires **no code changes** when new instructions are
added — the instruction table is data-driven.

---

## Testing

Tests live in `v0/kasm/semantic_test.go` in the `kasm_test` package (black-
box testing of the public API).

### Test helpers

| Helper                       | Purpose                                           |
|------------------------------|---------------------------------------------------|
| `minimalInstructions()`      | Returns a small `map[string]Instruction` for unit tests. |
| `requireSemanticErrorCount()`| Asserts the number of errors returned.             |
| `requireNoSemanticErrors()`  | Asserts zero errors.                               |
| `requireErrorContains()`     | Asserts a specific error message substring.        |

### Writing a test

1. **Construct a `*Program` directly** — do not use the lexer/parser for
   unit tests. This isolates the analyser from upstream behaviour:

   ```go
   func TestValidateMyNewRule_ErrorCase(t *testing.T) {
       program := &kasm.Program{
           Statements: []kasm.Statement{
               &kasm.InstructionStmt{
                   Mnemonic: "MOV",
                   Operands: []kasm.Operand{
                       &kasm.RegisterOperand{Name: "eax", Line: 1, Column: 5},
                       &kasm.RegisterOperand{Name: "ebx", Line: 1, Column: 10},
                   },
                   Line: 1, Column: 1,
               },
           },
       }
       errors := kasm.AnalyserNew(program, minimalInstructions()).Analyse()
       // ... assertions ...
   }
   ```

2. **Use `minimalInstructions()`** for most tests. Only use the full
   architecture table in integration tests.

3. **Assert on error count and message content**, not exact message strings
   (use `requireErrorContains` with substrings).

4. **Cover three cases per rule:**
   - Happy path (valid input → zero errors).
   - Error path (invalid input → correct error message).
   - Edge case (empty program, nil operands, no variants, forward
     references, etc.).

### Running tests

```bash
cd v0/kasm && go test -v -run TestMyNewRule
```

Or run the full semantic test suite:

```bash
cd v0/kasm && go test -v -run TestAnalyser
```

---

## Checklist

Use this checklist when adding any feature to the semantic analyser:

- [ ] **AST change** (if any): new `Statement` or `Operand` type added to
      `ast.go` with marker method, `Line`/`Column` fields, and accessor
      methods.
- [ ] **Collection pass**: new declaration type added to `collect()` with
      a corresponding internal helper type and map field on `Analyser`
      (initialised in `AnalyserNew`).
- [ ] **Validation pass**: new case added to `validate()` dispatching to a
      `validate<Thing>` method.
- [ ] **Operand type mapping**: `operandSemanticType()` updated if a new
      operand kind was added.
- [ ] **Operand validation**: `validateOperands()` updated if the new
      operand kind needs per-operand checks.
- [ ] **Identifier substitution**: `tryIdentifierSubstitution()` updated if
      the new operand kind should be compatible with variant operand types.
- [ ] **Error messages**: all `addError()` calls use the exact format
      specified in the requirements (or a new format documented in the
      requirements).
- [ ] **Tests**: unit tests added to `semantic_test.go` covering happy path,
      error path, and edge cases. Tests construct `*Program` directly.
- [ ] **No AST mutation**: verified that the new code does not modify any
      AST node.
- [ ] **No panics**: verified that the new code handles nil/empty/unexpected
      input gracefully.
- [ ] **No new imports**: the analyser does not import architecture-specific
      packages. If new types are needed, they come from `v0/architecture`
      (generic) or `v0/kasm` (same package).
- [ ] **Requirements updated**: new FR/NFR items added to
      `.requirements/semantics/requirements.md` and the validation summary
      table extended.

