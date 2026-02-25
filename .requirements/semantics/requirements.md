# Semantic Analysis

The semantic analyser validates a `*Program` AST (produced by the parser)
against the rules of the `.kasm` language and the target architecture. It
detects errors that are syntactically legal but semantically invalid — unknown
instructions, wrong operand counts, mismatched operand types, duplicate labels,
unresolved symbol references, and namespace violations. The semantic analyser
sits between the parser and the code-generation stage in the assembly pipeline.

The semantic analyser is **architecture-aware**: it receives an architecture
description (instruction groups with their variants) at construction time and
uses it to validate instruction operands. Because the architecture description
is injected, the same analyser logic handles any architecture for which
instruction metadata exists.

The semantic analyser lives in `v0/kasm` and is consumed by the assembly
pipeline in `cmd/cli/cmd/x86_64/assemble_file.go`.

## Pipeline Position

```
parser output (*Program AST)
        │
        ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Semantic Analyser                             │
│  AnalyserNew(program, instructions) → Analyse() → []SemanticError│
│                                                                  │
│  ┌─────────────────────────────┐                                 │
│  │  Instruction metadata       │ ← injected at construction      │
│  │  (groups, variants, operand │                                 │
│  │   types)                    │                                 │
│  └─────────────────────────────┘                                 │
└──────────────────────┬───────────────────────────────────────────┘
                       │ validated AST + diagnostics
                       ▼
                 code generation
```

---

## Functional Requirements

### FR-1: Construction

An `Analyser` represents a ready-to-validate consumer of a `*Program` AST.
If an `Analyser` value exists, it is guaranteed to hold a valid program
reference and initialised internal state. There is no uninitialised or
partially-constructed state.

- **FR-1.1** `AnalyserNew(program, instructions)` is the sole constructor. It
  accepts the `*Program` AST produced by `Parser.Parse()` and an instruction
  lookup table, and returns an `*Analyser` that is ready for `Analyse()` to
  be called. There is no separate `Init()` step.
- **FR-1.2** `AnalyserNew` is infallible — it cannot fail. An empty `Program`
  (zero statements) is valid and will produce zero errors. A `nil` program
  must be treated as an empty program. Because the parser always returns a
  non-nil `*Program` (parser FR-2.2), a `nil` input indicates a programming
  error — but the analyser must not panic.
- **FR-1.3** The instruction lookup table must provide O(1) mnemonic-to-
  instruction resolution. The table is a `map[string]Instruction` keyed by
  upper-case mnemonic (matching the `v0/architecture` convention). Because
  the table is built by the orchestrator from `InstructionGroup` data, the
  analyser does not import any architecture-specific package directly.
- **FR-1.4** The analyser must initialise an empty label table, an empty
  namespace tracker, and an empty error slice during construction. Because
  all internal state is created at construction time, `Analyse()` has no
  precondition checks.

### FR-2: Analysis (Analyse)

`Analyse()` performs a single left-to-right pass over the `Program.Statements`
slice and returns a `[]SemanticError` slice. It is the sole public method that
drives analysis.

- **FR-2.1** `Analyse()` must visit every statement in the program exactly
  once, in source order. Because the AST preserves source order (parser
  FR-3.1.2), the analyser processes statements in the same order they appear
  in the source file.
- **FR-2.2** `Analyse()` must return a `[]SemanticError` slice containing all
  errors encountered during analysis. If no errors occurred, the slice must
  be empty (not `nil`). Each error must carry `Line`, `Column`, and a
  human-readable `Message`.
- **FR-2.3** `Analyse()` must not modify the AST. The analyser is a read-only
  consumer — it inspects nodes and records diagnostics but does not transform
  the tree. Because the AST is shared with downstream stages, mutation would
  corrupt their input.
- **FR-2.4** `Analyse()` must not abort on the first error. It must continue
  analysing subsequent statements to report as many issues as possible in a
  single pass. Because errors are appended to a slice, they are returned in
  source order.
- **FR-2.5** `Analyse()` may be called only once per `Analyser` instance.
  Calling it again would re-analyse from an already-populated label table
  and could produce duplicate-label false positives. The analyser does not
  reset internal state.
- **FR-2.6** The analysis must be performed in two logical phases within the
  single pass:
  1. **Collection phase** — gather all label declarations and namespace
     declarations into lookup tables so that forward references can be
     resolved.
  2. **Validation phase** — validate every statement against the collected
     tables and the instruction metadata.

  Because `.kasm` allows forward references (e.g. `jmp label` before
  `label:` is declared), the collection phase must complete before the
  validation phase begins. The two phases may be implemented as two
  sequential loops over `Program.Statements`, or as a single loop with
  deferred validation — the requirement is that forward references are
  resolvable, not a specific implementation strategy.

### FR-3: Instruction Validation

When the analyser encounters an `InstructionStmt`, it must validate the
mnemonic and its operands against the architecture's instruction metadata.

#### FR-3.1: Mnemonic Validation

- **FR-3.1.1** The analyser must look up the instruction mnemonic
  (case-insensitive) in the instruction table. If the mnemonic is not found,
  a `SemanticError` must be recorded: `"unknown instruction '<mnemonic>'"`.
  Because the lexer classifies mnemonics via the `ArchitectureProfile`
  (lexer FR-1), an unknown mnemonic here indicates a profile/table mismatch
  — this is a configuration error, but the analyser must still report it
  gracefully.
- **FR-3.1.2** The `use` instruction is a language-level construct (parser
  FR-7.6). It is parsed as a `UseStmt`, not an `InstructionStmt`, so the
  analyser never encounters it via instruction validation. No special-casing
  is needed.

#### FR-3.2: Operand Count Validation

- **FR-3.2.1** If the instruction has variants (i.e. `Instruction.HasVariants()`
  returns `true`), the analyser must check whether the number of operands
  supplied matches the operand count of at least one variant. If no variant
  matches the supplied operand count, a `SemanticError` must be recorded:
  `"instruction '<mnemonic>' expects <n> operand(s), got <m>"`.
  Because variants define the valid operand counts, the analyser delegates
  count validation to the variant metadata.
- **FR-3.2.2** If the instruction has no variants (the `Variants` slice is
  empty), operand-count validation is skipped. This accommodates instructions
  whose variant metadata has not yet been defined — the analyser does not
  block assembly of instructions that lack variant data.

#### FR-3.3: Operand Type Validation

- **FR-3.3.1** For each operand in the `InstructionStmt`, the analyser must
  determine the operand's semantic type based on its AST node kind:

  | AST Node Kind      | Semantic Operand Type |
  |--------------------|-----------------------|
  | `RegisterOperand`  | `"register"`          |
  | `ImmediateOperand` | `"immediate"`         |
  | `MemoryOperand`    | `"memory"`            |
  | `IdentifierOperand`| `"identifier"`        |
  | `StringOperand`    | `"string"`            |

- **FR-3.3.2** The analyser must attempt to find a variant whose operand
  type signature matches the supplied operands using
  `Instruction.FindVariant(operandTypes...)`. If no matching variant is
  found and at least one variant exists, a `SemanticError` must be recorded:
  `"no variant of '<mnemonic>' accepts operands (<type1>, <type2>, ...)"`.
  Because `FindVariant` performs an exact match on operand types and count,
  this check subsumes count validation when variants are present.
- **FR-3.3.3** `IdentifierOperand` may represent a label reference, a data
  symbol, or other named value whose type cannot be resolved until link time.
  The analyser must treat `"identifier"` as compatible with `"relative"` and
  `"far"` variant operand types (used by jump/call instructions). Because
  label addresses are resolved by the code generator, the analyser cannot
  determine the exact address type — it must accept the identifier
  optimistically.
- **FR-3.3.4** If the instruction has no variants, operand-type validation is
  skipped (same rationale as FR-3.2.2).

### FR-4: Label Validation

Labels are declaration-site identifiers. The analyser must ensure they are
unique within their scope and that all references to labels can be resolved.

#### FR-4.1: Duplicate Label Detection

- **FR-4.1.1** The analyser must maintain a label table (map of label name to
  declaration location). When a `LabelStmt` is encountered, the analyser
  must check whether the label name already exists in the table. If it does,
  a `SemanticError` must be recorded:
  `"duplicate label '<name>', previously declared at <line>:<column>"`.
  Because labels are scoped to the compilation unit (single file after
  pre-processing), the table is flat — there is no nested scope.
- **FR-4.1.2** The first declaration of a label is always accepted. Only the
  second (and subsequent) declarations of the same name produce errors.
  Because the table stores the first declaration's location, the error
  message can reference where the original declaration was.

#### FR-4.2: Undefined Label Detection

- **FR-4.2.1** The analyser must check every `IdentifierOperand` in every
  `InstructionStmt` against the label table. If the identifier does not
  match any declared label and does not match any other known symbol (e.g.
  a namespace-qualified name), a `SemanticError` must be recorded:
  `"undefined reference to '<name>'"`.
- **FR-4.2.2** Because `.kasm` allows forward references (FR-2.6), the
  undefined-label check must run after all labels have been collected. If
  the analyser uses two passes, this check belongs in the second pass. If
  it uses a single pass with deferred validation, the check must be
  deferred until all statements have been visited.
- **FR-4.2.3** Identifiers that appear in non-instruction contexts (e.g.
  `UseStmt.ModuleName`, `NamespaceStmt.Name`) are not label references and
  must not be checked against the label table. Because these identifiers
  have different semantics (module names, namespace names), they are
  validated by their own rules (FR-5, FR-6).

### FR-5: Namespace Validation

Namespaces group related code under a name. The analyser must validate
namespace declarations.

- **FR-5.1** When a `NamespaceStmt` is encountered, the analyser must record
  the namespace name. If the same namespace name is declared more than once
  in the same compilation unit, a `SemanticError` must be recorded:
  `"duplicate namespace '<name>', previously declared at <line>:<column>"`.
  Because namespace declarations are top-level statements, the check is a
  simple duplicate-name detection — analogous to duplicate-label detection.
- **FR-5.2** The namespace name must be a valid identifier (non-empty, does
  not start with a digit). Because the parser guarantees the name is a
  `TokenIdentifier` (parser FR-3.6.1), this check is a defence-in-depth
  measure — it should never fail in practice.
- **FR-5.3** Future extensions may introduce namespace-scoped label
  resolution (e.g. `namespace.label`). The current analyser does not need
  to implement scoped resolution, but the namespace table must be available
  for downstream stages that do.

### FR-6: Use Statement Validation

`use` imports a module by name. The analyser must validate the module
reference.

- **FR-6.1** When a `UseStmt` is encountered, the analyser must record the
  module name. If the same module name is imported more than once, a
  `SemanticError` must be recorded:
  `"duplicate use of module '<name>', previously imported at <line>:<column>"`.
- **FR-6.2** The module name must be a valid identifier (non-empty). Because
  the parser guarantees the name is a `TokenIdentifier` (parser FR-3.7.1),
  this is a defence-in-depth check.
- **FR-6.3** Module resolution (locating the module's source file or compiled
  artefact) is not the analyser's responsibility. The analyser validates
  the `use` statement syntactically and records it — a later linker or
  module resolver consumes the information.

### FR-7: Directive Validation

Directives that survive into the AST (not consumed by the pre-processor) are
captured as `DirectiveStmt` nodes. The analyser must validate them.

- **FR-7.1** A `DirectiveStmt` whose literal is not a recognised
  post-pre-processing directive should produce a `SemanticError`:
  `"unrecognised directive '<literal>'"`. Because the pre-processor consumes
  `%include`, `%macro`, `%endmacro`, `%define`, `%ifdef`, `%ifndef`,
  `%else`, and `%endif`, any directive that reaches the AST is either a
  language-level directive not yet defined, or a user error.
- **FR-7.2** If future language-level directives are added (e.g. `%section`,
  `%align`), the analyser must recognise them and validate their arguments.
  The current implementation may treat all surviving directives as
  unrecognised (FR-7.1) — this is a valid starting point.

### FR-8: Immediate Value Validation

`ImmediateOperand` values are stored as verbatim strings by the parser. The
analyser must validate that they represent legal numeric values.

- **FR-8.1** Decimal immediates must consist of one or more digits (`0`–`9`).
  A `SemanticError` must be recorded if the string cannot be parsed as a
  valid integer: `"invalid immediate value '<value>'"`.
- **FR-8.2** Hexadecimal immediates must start with `0x` or `0X` followed by
  one or more hex digits (`0`–`9`, `a`–`f`, `A`–`F`). The same error
  message applies if parsing fails.
- **FR-8.3** Overflow detection is optional in the initial implementation.
  If implemented, the analyser should warn (not error) when an immediate
  exceeds the maximum value for the instruction's operand size. Because
  operand sizes are variant-specific, overflow detection requires variant
  resolution to have succeeded.

### FR-9: Memory Operand Validation

`MemoryOperand` nodes contain a `Components` slice of raw tokens. The
analyser must validate the structure of the memory reference.

- **FR-9.1** A memory operand must contain at least one component. An empty
  `Components` slice (i.e. `[]`) must produce a `SemanticError`:
  `"empty memory operand"`.
- **FR-9.2** The base of a memory operand must be a register or an
  identifier. If the first non-operator component is an immediate, a
  `SemanticError` must be recorded:
  `"memory operand base must be a register or identifier, got immediate"`.
- **FR-9.3** Displacement components (after a `+` or `-` operator) must be
  registers or immediates. An identifier as a displacement is valid
  (representing a symbolic offset).
- **FR-9.4** Operator tokens within a memory operand must be `+` or `-`.
  Any other operator (e.g. `*`, `/`) must produce a `SemanticError`:
  `"invalid operator '<op>' in memory operand"`.

---

## Architecture

### AR-1: File Layout

The semantic analyser lives in `v0/kasm` alongside the parser, lexer, and AST
definitions. Because the analyser consumes `*Program`, `Statement`, `Operand`,
and their concrete types from the same package, no cross-package import is
required for the core data types.

| File                 | Responsibility                                          |
|----------------------|---------------------------------------------------------|
| `semantic.go`        | `Analyser` struct, `AnalyserNew`, `Analyse`, validation methods. |
| `semantic_error.go`  | `SemanticError` type definition.                        |

- **AR-1.1** The analyser (`semantic.go`) must not import any architecture-
  specific package. It receives instruction metadata as a
  `map[string]Instruction` — the orchestrator is responsible for building
  this map from `v0/architecture` data. Because the analyser depends on the
  generic `Instruction` type (not a concrete architecture package), it is
  architecture-agnostic at the source level.
- **AR-1.2** The `SemanticError` type lives in `semantic_error.go`, separate
  from the analysis logic. It is a plain data struct (like `ParseError`) —
  not an `error` interface implementation — so that multiple errors can be
  accumulated and returned as a slice.
- **AR-1.3** The analyser imports `v0/architecture` for the `Instruction` and
  `InstructionVariant` types. This is the only external dependency beyond the
  standard library. Because `v0/architecture` defines generic data structures
  (not architecture-specific logic), this import does not violate the
  architecture-agnostic principle.

### AR-2: Separation of Concerns

```
┌────────────────────────────────────────────────────────────────┐
│                          v0/kasm                                │
│                                                                │
│  ┌──────────────────┐     ┌──────────────────────────────────┐ │
│  │   Parser          │────▶│   *Program (AST)                 │ │
│  │   parsing.go      │     └───────────────┬─────────────────┘ │
│  └──────────────────┘                      │                   │
│                                            ▼                   │
│  ┌──────────────────┐     ┌──────────────────────────────────┐ │
│  │ Semantic Analyser │────▶│   []SemanticError                │ │
│  │ semantic.go       │     └──────────────────────────────────┘ │
│  └────────┬─────────┘                                          │
│           │ imports                                             │
│           ▼                                                    │
│  ┌──────────────────┐                                          │
│  │ v0/architecture   │                                          │
│  │  Instruction      │                                          │
│  │  InstructionVariant│                                         │
│  └──────────────────┘                                          │
└────────────────────────────────────────────────────────────────┘
```

- **AR-2.1** The analyser depends on the AST types (`Program`, `Statement`,
  `Operand`, and all concrete node types) from the same `v0/kasm` package.
  There is no cross-package import for AST access.
- **AR-2.2** The analyser depends on `v0/architecture.Instruction` and
  `v0/architecture.InstructionVariant` for instruction metadata. This is
  the sole external dependency.
- **AR-2.3** The analyser does not depend on the `ArchitectureProfile`
  interface or the `v0/kasm/profile` sub-package. Profile concerns belong
  to the lexer — the analyser works with the richer `Instruction` model
  that includes variants and operand types.
- **AR-2.4** The analyser does not depend on `debugcontext` directly. Like
  the parser, it returns plain data (`[]SemanticError`). The orchestrator
  translates semantic errors into `debugcontext.Entry` values. An optional
  `WithDebugContext` method may be provided for convenience (following the
  parser's pattern), but it is not required — the orchestrator can perform
  the translation externally.

### AR-3: Instruction Table Construction

The instruction table passed to `AnalyserNew` must be constructed by the
orchestrator from the architecture package's instruction groups.

- **AR-3.1** The orchestrator must flatten all `InstructionGroup` maps into a
  single `map[string]Instruction` keyed by upper-case mnemonic. Because the
  architecture package stores mnemonics in upper case (e.g. `"MOV"`,
  `"JMP"`), the analyser must upper-case the AST mnemonic before lookup.
- **AR-3.2** If two groups contain the same mnemonic, the last one wins. This
  is consistent with `InstructionGroup.Merge` semantics. The orchestrator
  should log a warning via the debug context if a mnemonic collision is
  detected.
- **AR-3.3** The instruction table is read-only after construction. The
  analyser does not modify it. Because the table is shared with no writers,
  no synchronisation is required.

---

## Non-Functional Requirements

### NFR-1: Performance

- **NFR-1.1** The analyser must perform at most two passes over the statement
  slice — one for collection (labels, namespaces, uses) and one for
  validation. No additional passes are permitted. Because each pass is
  linear in the number of statements, the total work is O(n).
- **NFR-1.2** Label, namespace, and module-name lookups must be O(1) via
  `map[string]` access. Because the tables are built during the collection
  pass, per-statement validation cost is constant.
- **NFR-1.3** Instruction lookup must be O(1) via the injected instruction
  table. Variant matching is O(v) where v is the number of variants per
  instruction — typically 1–4, effectively constant.

### NFR-2: Correctness

- **NFR-2.1** The analyser must handle all valid AST structures without
  panicking. Because the parser guarantees every statement is one of the
  defined `Statement` kinds (parser FR-3.2), the analyser's type switch
  covers all possibilities.
- **NFR-2.2** Source positions on `SemanticError` values must be accurate.
  `Line` and `Column` must correspond to the first token of the construct
  that triggered the error. Because AST nodes carry `Line`/`Column` from
  the parser (parser AR-3.4), the analyser copies these values directly.
- **NFR-2.3** The analyser must produce identical output for a given
  `(*Program, instruction-table)` pair regardless of when or how many times
  it is invoked (deterministic, side-effect-free). Because the analyser has
  no mutable package-level state, the only state that changes is local to
  the `Analyser` instance.
- **NFR-2.4** Forward references must resolve correctly. A `jmp label`
  before `label:` must not produce an "undefined reference" error. Because
  the collection pass runs before the validation pass (FR-2.6), all labels
  are known before any reference is checked.

### NFR-3: Testability

- **NFR-3.1** The analyser must be testable with only a `*Program` AST and
  an instruction table — no file I/O, no profile, no debug context. Because
  the analyser takes plain data and returns plain data, all dependencies are
  injectable via the constructor.
- **NFR-3.2** Tests must live in the `kasm_test` package
  (`v0/kasm/semantic_test.go`) to verify the public API surface only.
- **NFR-3.3** Test cases should construct `*Program` values directly (not
  via the parser or lexer) to isolate the analyser from upstream behaviour.
  When integration tests are needed, using the full
  `LexerNew → Start → ParserNew → Parse → AnalyserNew → Analyse` pipeline
  is acceptable.
- **NFR-3.4** Each validation rule must have dedicated test cases covering:
  - The happy path (well-formed input, zero errors).
  - The error case (invalid input, expected error message).
  - Edge cases (empty program, instruction with no variants, forward
    reference).
- **NFR-3.5** A minimal instruction table (e.g. containing only `MOV` with
  two variants) should be used in most tests. Full architecture tables
  should only be used in integration tests.

### NFR-4: Integration

- **NFR-4.1** The analyser receives its input from `Parser.Parse()`. The
  `*Program` AST is passed directly to `AnalyserNew()`.
- **NFR-4.2** The `[]SemanticError` slice is consumed by the orchestrator
  (`assemble_file.go`), which translates each error into a
  `debugcontext.Entry` with severity `"error"`. Because the analyser
  returns `Line` and `Column` on each error, the orchestrator can construct
  accurate `debugcontext.Location` values.
- **NFR-4.3** The orchestrator must check `len(errors) > 0` after
  `Analyse()` and abort the pipeline if errors are present — analogous to
  the parser error check. Because the analyser reports all errors (not just
  the first), the user sees the full set of issues in one pass.
- **NFR-4.4** The orchestrator must set the debug context phase to
  `"semantic-analysis"` before invoking `Analyse()`. If the analyser has a
  `WithDebugContext` method, the phase is set internally; otherwise, the
  orchestrator sets it externally.

### NFR-5: Extensibility

- **NFR-5.1** Adding a new validation rule (e.g. register-size mismatch
  detection) requires adding a validation method in `semantic.go` and
  calling it from the appropriate statement handler. No existing validation
  rules are modified.
- **NFR-5.2** Adding a new statement kind to the AST requires adding a case
  to the analyser's statement dispatch switch. The existing cases are not
  modified.
- **NFR-5.3** Adding a new operand kind to the AST requires updating the
  operand-type-to-string mapping (FR-3.3.1) and potentially adding a new
  validation rule. Existing operand kinds are not modified.

---

## Data Model

### SemanticError

| Field    | Type     | Description                                         |
|----------|----------|-----------------------------------------------------|
| `Message`| `string` | Human-readable description of the semantic error.   |
| `Line`   | `int`    | 1-based line number where the error was detected.   |
| `Column` | `int`    | 1-based column number where the error was detected. |

### Analyser Struct

| Field          | Type                          | Description                                            |
|----------------|-------------------------------|--------------------------------------------------------|
| `program`      | `*Program`                    | The AST to analyse.                                    |
| `instructions` | `map[string]Instruction`      | Instruction lookup table (upper-case mnemonic keys).   |
| `labels`       | `map[string]labelDecl`        | Label name → declaration location.                     |
| `namespaces`   | `map[string]namespaceDecl`    | Namespace name → declaration location.                 |
| `modules`      | `map[string]useDecl`          | Module name → import location.                         |
| `errors`       | `[]SemanticError`             | Accumulated semantic errors.                           |

### Internal Helper Types

| Type            | Fields                    | Description                              |
|-----------------|---------------------------|------------------------------------------|
| `labelDecl`     | `Name string`, `Line int`, `Column int` | Tracks where a label was declared.  |
| `namespaceDecl` | `Name string`, `Line int`, `Column int` | Tracks where a namespace was declared. |
| `useDecl`       | `Name string`, `Line int`, `Column int` | Tracks where a module was imported.  |

### Validation Summary

| Check                        | Statement Type    | Error Condition                                    | Severity |
|------------------------------|-------------------|----------------------------------------------------|----------|
| Unknown instruction          | `InstructionStmt` | Mnemonic not in instruction table.                  | Error    |
| Operand count mismatch       | `InstructionStmt` | No variant matches the supplied operand count.      | Error    |
| Operand type mismatch        | `InstructionStmt` | No variant matches the supplied operand types.      | Error    |
| Duplicate label              | `LabelStmt`       | Label name already declared.                        | Error    |
| Undefined reference          | `InstructionStmt` | `IdentifierOperand` does not match any label.       | Error    |
| Duplicate namespace          | `NamespaceStmt`   | Namespace name already declared.                    | Error    |
| Duplicate use                | `UseStmt`         | Module name already imported.                       | Error    |
| Unrecognised directive       | `DirectiveStmt`   | Directive literal not in recognised set.            | Error    |
| Invalid immediate value      | `InstructionStmt` | `ImmediateOperand.Value` cannot be parsed as number.| Error    |
| Empty memory operand         | `InstructionStmt` | `MemoryOperand.Components` is empty.                | Error    |
| Invalid memory operand base  | `InstructionStmt` | First component is an immediate.                    | Error    |
| Invalid memory operator      | `InstructionStmt` | Operator in memory operand is not `+` or `-`.       | Error    |

