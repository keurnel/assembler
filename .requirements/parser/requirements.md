# Parser

The parser transforms an ordered sequence of `Token` values (produced by the
lexer) into a structured Abstract Syntax Tree (AST). Each AST node represents
a syntactic construct in the `.kasm` language — an instruction with its
operands, a label declaration, a namespace block, a `use` import, or a
directive. The parser sits between the lexer and the semantic analyser /
code-generation stages in the assembly pipeline.

The parser is **architecture-agnostic**: it does not validate instruction
mnemonics, register names, or operand counts. It recognises the _shape_ of
constructs (e.g. "instruction followed by operands separated by commas") but
defers validation to a later semantic-analysis pass. Because the parser
operates on token types — not literal values — the same parser handles any
architecture for which a lexer profile exists.

The parser lives in `v0/kasm` (`parsing.go`) and is consumed by the assembly
pipeline in `cmd/cli/cmd/x86_64/assemble_file.go`.

## Pipeline Position

```
lexer output ([]Token)
        │
        ▼
┌──────────────────────────────────────────────────────────┐
│                       Parser                              │
│  ParserNew(tokens) → Parse() → (*Program, []ParseError)  │
└──────────────────────┬────────────────────────────────────┘
                       │ AST + diagnostics
                       ▼
              semantic analysis / code generation
```

---

## Functional Requirements

### FR-1: Construction

A `Parser` represents a ready-to-parse consumer of a token slice. If a
`Parser` value exists, it is guaranteed to hold a valid token slice and
initialised position state. There is no uninitialised or partially-constructed
state.

- **FR-1.1** `ParserNew(tokens)` is the sole constructor. It accepts the
  `[]Token` slice produced by `Lexer.Start()` and returns a `*Parser` that
  is ready for `Parse()` to be called. There is no separate `Init()` step.
- **FR-1.2** `ParserNew` is infallible — it cannot fail. An empty token slice
  is valid and will produce an empty `Program`. A `nil` slice must be treated
  as empty (zero tokens). Because the lexer always returns a non-nil slice
  (lexer FR-2.5), a `nil` input indicates a programming error — but the
  parser must not panic.
- **FR-1.3** The `Position` field must start at `0`, pointing to the first
  token. Because the parser does not pre-read a token during construction
  (unlike the lexer), `Position` is at the beginning of the slice.
- **FR-1.4** The `Tokens` slice must be stored by reference. The parser must
  not copy or modify the tokens — it reads them in order. Because tokens are
  value types (lexer FR-8), storing the slice header is sufficient.

### FR-2: Parsing (Parse)

`Parse()` performs a single left-to-right pass over the token slice and
returns a `*Program` AST and a slice of `ParseError` values. It is the sole
public method that drives parsing.

- **FR-2.1** `Parse()` must consume the entire token slice, stopping when
  `Position` reaches the end of the slice. Because each branch of the main
  loop consumes at least one token, the parser always makes progress —
  infinite loops are impossible.
- **FR-2.2** `Parse()` must return a `*Program` containing all successfully
  parsed AST nodes in source order. Even when errors are encountered, the
  parser must produce as many nodes as it can — it does not abort on the
  first error.
- **FR-2.3** `Parse()` must return a `[]ParseError` slice containing all
  errors encountered during parsing. If no errors occurred, the slice must
  be empty (not `nil`). Each error must carry the originating token's `Line`
  and `Column` so that later stages can report accurate source positions.
- **FR-2.4** `Parse()` may be called only once per `Parser` instance. Calling
  it again would re-parse from the current (exhausted) position and return
  an empty program. The parser does not reset internal state.

### FR-3: AST Node Types

Every construct in the `.kasm` language maps to exactly one AST node type.
The parser produces a flat list of top-level statements inside a `Program`.
Because `.kasm` is a line-oriented assembly language, there is no nested
expression tree — operands are leaves, not recursive sub-expressions.

#### FR-3.1: Program

- **FR-3.1.1** `Program` is the root AST node. It holds an ordered slice of
  `Statement` nodes representing every top-level construct in the source.
- **FR-3.1.2** The `Program` must preserve source order. Statements appear
  in the same order as the corresponding tokens in the input slice.

#### FR-3.2: Statement

`Statement` is a sum type (interface or tagged union) representing one
top-level construct. Every statement carries the `Line` and `Column` of
its first token for diagnostic purposes.

The following statement kinds must be supported:

| Kind                 | Description                                                    |
|----------------------|----------------------------------------------------------------|
| `InstructionStmt`    | An instruction mnemonic followed by zero or more operands.     |
| `LabelStmt`          | A label declaration (identifier ending in `:`).                |
| `NamespaceStmt`      | A `namespace` keyword followed by a name identifier.           |
| `UseStmt`            | A `use` instruction followed by a module name identifier.      |
| `DirectiveStmt`      | A pre-processor directive that survived into the token stream. |

#### FR-3.3: InstructionStmt

- **FR-3.3.1** An `InstructionStmt` must store the instruction mnemonic
  (literal string) and an ordered slice of `Operand` nodes. Because the
  mnemonic is stored as a string (not a token type constant), the parser
  does not restrict which instructions are valid — that is a semantic
  concern.
- **FR-3.3.2** Operands are separated by `,` tokens (`TokenIdentifier` with
  literal `","`). The comma is consumed but not stored — it is syntactic
  punctuation, not a semantic operand.
- **FR-3.3.3** Zero operands is valid (e.g. `ret`, `syscall`, `nop`, `hlt`).
  One operand is valid (e.g. `push rax`, `jmp label`). Two operands is the
  common case (e.g. `mov rax, 1`). The parser must accept any number of
  operands without an upper bound — operand-count validation is a semantic
  concern.
- **FR-3.3.4** The `InstructionStmt` must carry `Line` and `Column` from the
  instruction token so that errors can reference the instruction position.

#### FR-3.4: Operand

An `Operand` represents a single argument to an instruction. Operands are
not recursive — there are no sub-expressions. Each operand is one of the
following kinds:

| Kind              | Token Type         | Example          |
|-------------------|--------------------|------------------|
| `RegisterOperand` | `TokenRegister`    | `rax`, `r8`      |
| `ImmediateOperand`| `TokenImmediate`   | `42`, `0xFF`     |
| `IdentifierOperand`| `TokenIdentifier` | `label`, `msg`   |
| `StringOperand`   | `TokenString`      | `"Hello"`        |
| `MemoryOperand`   | composite          | `[rbp]`, `[rax + 8]` |

- **FR-3.4.1** `RegisterOperand` wraps a `TokenRegister`. The literal is
  stored verbatim (preserving original casing).
- **FR-3.4.2** `ImmediateOperand` wraps a `TokenImmediate`. The literal is
  stored verbatim (`"42"`, `"0xFF"`). Numeric conversion is deferred to
  semantic analysis or code generation.
- **FR-3.4.3** `IdentifierOperand` wraps a `TokenIdentifier` that is not a
  label (no trailing `:`), not a comma, and not a bracket. This covers
  symbolic references such as label names and data symbols.
- **FR-3.4.4** `StringOperand` wraps a `TokenString`. The literal contains
  the content between the quotes (delimiters already stripped by the lexer).
- **FR-3.4.5** `MemoryOperand` represents a memory reference enclosed in
  `[` and `]`. The parser must consume the opening `[`, collect the inner
  tokens (base register, optional displacement, optional index), and consume
  the closing `]`. The inner tokens are stored as an ordered slice of
  `Operand` or component nodes, preserving operators (`+`, `-`).
- **FR-3.4.6** Each `Operand` must carry `Line` and `Column` from its
  originating token for diagnostic purposes.

#### FR-3.5: LabelStmt

- **FR-3.5.1** A `LabelStmt` is produced when the parser encounters a
  `TokenIdentifier` whose literal ends with `:`. Because the lexer appends
  the `:` to the literal (lexer FR-4.6.2), the parser can identify labels
  by checking the trailing character.
- **FR-3.5.2** The `LabelStmt` must store the label name _without_ the
  trailing `:`. Stripping the colon is the parser's responsibility — the
  lexer provides the raw form, the parser provides the semantic form.
- **FR-3.5.3** The `LabelStmt` must carry `Line` and `Column` from the
  label token.

#### FR-3.6: NamespaceStmt

- **FR-3.6.1** A `NamespaceStmt` is produced when the parser encounters a
  `TokenKeyword` with literal `namespace` (case-insensitive). The next
  token must be a `TokenIdentifier` providing the namespace name.
- **FR-3.6.2** If the `TokenKeyword` `namespace` is not followed by a
  `TokenIdentifier`, the parser must record a `ParseError` (e.g. "expected
  namespace name") and recover. The `NamespaceStmt` is not emitted.
- **FR-3.6.3** The `NamespaceStmt` must store the namespace name (the
  identifier's literal) and carry `Line`/`Column` from the keyword token.

#### FR-3.7: UseStmt

- **FR-3.7.1** A `UseStmt` is produced when the parser encounters a
  `TokenInstruction` with literal `use` (case-insensitive). The next token
  must be a `TokenIdentifier` providing the module name.
- **FR-3.7.2** If the `TokenInstruction` `use` is not followed by a
  `TokenIdentifier`, the parser must record a `ParseError` and recover.
- **FR-3.7.3** The `UseStmt` must store the module name and carry
  `Line`/`Column` from the `use` token.

#### FR-3.8: DirectiveStmt

- **FR-3.8.1** A `DirectiveStmt` is produced when the parser encounters a
  `TokenDirective`. Directives that survived pre-processing into the token
  stream (e.g. stray `%define` not consumed by the pre-processor) are
  captured as-is for later stages to handle or ignore.
- **FR-3.8.2** The `DirectiveStmt` must store the full directive literal
  (including the `%` prefix) and any argument tokens that follow it on the
  same logical line. The parser collects argument tokens until it encounters
  the start of a new statement (instruction, label, keyword, directive, or
  end of input).
- **FR-3.8.3** The `DirectiveStmt` must carry `Line`/`Column` from the
  directive token.

### FR-4: Token Consumption

The parser advances through the token slice one token at a time, using a set
of helper methods. Because all advancement goes through these helpers,
bounds-checking is centralised — there is no alternative path that could
read past the end of the slice.

- **FR-4.1** `current()` must return the token at `Position`, or a sentinel
  zero-value `Token` if `Position` is at or past the end. Because every
  parsing branch calls `current()` before inspecting a token, out-of-bounds
  access is impossible.
- **FR-4.2** `peek()` must return the token at `Position + 1` without
  advancing, or a sentinel zero-value `Token` if no next token exists.
  Because `peek()` does not modify `Position`, it is safe to call at any
  point without side effects.
- **FR-4.3** `advance()` must increment `Position` by one and return the
  token that was at the previous position. If already at the end, it must
  return the sentinel zero-value `Token` without advancing further.
- **FR-4.4** `expect(tokenType)` must check that the current token matches
  the expected type. If it matches, the token is consumed (advanced past)
  and returned. If it does not match, a `ParseError` is recorded and the
  parser does not advance — allowing recovery logic to decide how to
  proceed.
- **FR-4.5** `isAtEnd()` must return `true` when `Position` is at or past
  the length of the token slice. Because this is checked at the top of
  the main parsing loop (FR-2.1), the loop always terminates.

### FR-5: Error Handling and Recovery

The parser must be resilient. A syntax error in one statement must not
prevent parsing of subsequent statements. Because `.kasm` is line-oriented,
recovery is straightforward — skip to the next statement boundary.

- **FR-5.1** `ParseError` must carry: `Message` (human-readable description),
  `Line` (1-based), and `Column` (1-based) from the token that triggered the
  error. Because every token carries position information (lexer FR-8.3,
  FR-8.4), the parser always has a source location for errors.
- **FR-5.2** When the parser encounters an unexpected token, it must record a
  `ParseError` in its error accumulator and attempt recovery. Recovery
  consists of advancing past tokens until the start of a recognisable
  statement is found (instruction, label, keyword, directive, or end of
  input).
- **FR-5.3** The parser must not panic on any input. Malformed token
  sequences, empty slices, and unexpected token types must all be handled
  gracefully. Because every branch of the main loop either parses a known
  construct or triggers recovery (FR-5.2), the parser always makes
  progress.
- **FR-5.4** Multiple errors may be accumulated during a single `Parse()`
  call. The parser does not abort on the first error — it continues parsing
  to report as many issues as possible. Because errors are appended to a
  slice, they are returned in source order.
- **FR-5.5** An empty token slice must produce an empty `Program` and zero
  errors. This is the normal case for an empty source file.

### FR-6: Statement Dispatch

The main parsing loop inspects the current token's type to determine which
parsing method to invoke. Because each token type maps to at most one
statement kind, dispatch is a simple switch — there is no ambiguity.

- **FR-6.1** `TokenInstruction` → parse as `InstructionStmt` (or `UseStmt`
  if the literal is `use`).
- **FR-6.2** `TokenIdentifier` with trailing `:` → parse as `LabelStmt`.
- **FR-6.3** `TokenIdentifier` without trailing `:` → treat as an operand
  that appears outside an instruction context. This is a parse error — record
  it and recover.
- **FR-6.4** `TokenKeyword` → dispatch by keyword literal. `namespace` →
  parse as `NamespaceStmt`. Unknown keywords → record a parse error and
  recover.
- **FR-6.5** `TokenDirective` → parse as `DirectiveStmt`.
- **FR-6.6** `TokenRegister`, `TokenImmediate`, `TokenString` outside an
  instruction context → parse error (operand without instruction). Record
  the error and recover.
- **FR-6.7** Any other token at the top level (e.g. stray punctuation) →
  record a parse error and advance past the token.

### FR-7: Instruction Parsing

When the parser encounters a `TokenInstruction`, it must collect the
instruction's operands.

- **FR-7.1** The parser must consume the instruction token and record its
  literal as the mnemonic.
- **FR-7.2** The parser must then consume zero or more operands separated by
  `,` (`TokenIdentifier` with literal `","`). Operand parsing continues
  until a token is encountered that cannot be an operand or a comma (i.e.
  the start of the next statement or end of input).
- **FR-7.3** Each operand is parsed by an operand sub-parser that dispatches
  on the token type:
  - `TokenRegister` → `RegisterOperand`
  - `TokenImmediate` → `ImmediateOperand`
  - `TokenIdentifier` (not `,`, not `[`, not `]`) → `IdentifierOperand`
  - `TokenString` → `StringOperand`
  - `TokenIdentifier` with literal `[` → begin `MemoryOperand` parsing
- **FR-7.4** Memory operand parsing (`[...]`): consume the `[`, collect
  inner tokens until `]` or end of input, consume the `]`. Inner tokens
  are parsed as a sequence of components (registers, immediates,
  identifiers, and operators `+` / `-`). An unterminated `[` (no matching
  `]`) must produce a `ParseError`.
- **FR-7.5** A missing comma between operands where one is expected must
  produce a `ParseError` but must not halt parsing of the current
  instruction. The parser should attempt to continue collecting operands.
- **FR-7.6** If the instruction literal (case-insensitive) is `use`, the
  parser must delegate to `UseStmt` parsing (FR-3.7) instead of generic
  instruction parsing. Because `use` is classified as `TokenInstruction`
  by the lexer profile, the parser must distinguish it by literal value.

### FR-8: Label Parsing

- **FR-8.1** The parser must consume the label token (a `TokenIdentifier`
  whose literal ends with `:`).
- **FR-8.2** The parser must strip the trailing `:` from the literal to
  produce the label name.
- **FR-8.3** A `LabelStmt` is emitted with the stripped name and the
  token's `Line`/`Column`.

### FR-9: Namespace Parsing

- **FR-9.1** The parser must consume the `TokenKeyword` with literal
  `namespace`.
- **FR-9.2** The parser must then expect and consume a `TokenIdentifier`
  as the namespace name. Because the lexer's context-sensitive
  classification rule (lexer FR-11.2) ensures the token after `namespace`
  is always `TokenIdentifier` (even if the name matches a register or
  instruction), the parser can rely on the type being `TokenIdentifier`.
- **FR-9.3** If the next token is not `TokenIdentifier` (e.g. end of
  input, or a non-identifier token), a `ParseError` must be recorded.

### FR-10: Use Parsing

- **FR-10.1** The parser must consume the `TokenInstruction` with literal
  `use`.
- **FR-10.2** The parser must then expect and consume a `TokenIdentifier`
  as the module name.
- **FR-10.3** If the next token is not `TokenIdentifier`, a `ParseError`
  must be recorded.

### FR-11: Directive Parsing

- **FR-11.1** The parser must consume the `TokenDirective`.
- **FR-11.2** The parser must then collect any argument tokens that follow
  on the same logical statement. Arguments are tokens that do not start a
  new statement (not `TokenInstruction`, not a label `TokenIdentifier`,
  not `TokenKeyword`, not `TokenDirective`). Argument tokens are stored
  as raw `Token` values on the `DirectiveStmt`.
- **FR-11.3** If the directive has no arguments, the argument slice must
  be empty (not `nil`).

---

## Architecture

### AR-1: File Layout

The parser lives in `v0/kasm` alongside the lexer and token definitions.
Because the parser consumes `Token` and `TokenType` from the same package,
no cross-package import is required for the core data types.

| File            | Responsibility                                            |
|-----------------|-----------------------------------------------------------|
| `parsing.go`    | `Parser` struct, `ParserNew`, `Parse`, parsing methods.   |
| `ast.go`        | AST node types (`Program`, `Statement`, `Operand`, etc.). |
| `parse_error.go`| `ParseError` type definition.                             |

- **AR-1.1** The parser (`parsing.go`) must not import any architecture-
  specific package. It operates exclusively on `Token` types produced by
  the lexer. Because the token types are architecture-independent (lexer
  AR-2.3), the parser inherits this independence.
- **AR-1.2** AST node types live in `ast.go`, separate from the parsing
  logic. This separation allows later stages (semantic analysis, code
  generation) to import the AST definitions without pulling in the parser
  implementation.
- **AR-1.3** The `ParseError` type lives in `parse_error.go`. It is a
  plain data struct — not an `error` interface implementation — so that
  multiple errors can be accumulated and returned as a slice.

### AR-2: Separation of Concerns

```
┌─────────────────────────────────────────────────────────┐
│                       v0/kasm                            │
│                                                         │
│  ┌──────────────────┐     ┌──────────────────────────┐  │
│  │   Lexer           │────▶│        []Token            │  │
│  │   lexer.go        │     └────────────┬─────────────┘  │
│  └──────────────────┘                   │                │
│                                         ▼                │
│  ┌──────────────────┐     ┌──────────────────────────┐  │
│  │   Parser          │────▶│   *Program (AST)         │  │
│  │   parsing.go      │     │   []ParseError           │  │
│  └──────────────────┘     └──────────────────────────┘  │
│                                                         │
│  ┌──────────────────┐     ┌──────────────────────────┐  │
│  │   ast.go          │     │  parse_error.go          │  │
│  │   AST node types  │     │  ParseError struct       │  │
│  └──────────────────┘     └──────────────────────────┘  │
│                                                         │
│  ┌──────────────────┐     ┌──────────────────────────┐  │
│  │   token.go        │     │  token_types.go          │  │
│  └──────────────────┘     └──────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

- **AR-2.1** The parser depends on `Token`, `TokenType`, and the token type
  constants — all defined in the same package. There are no external
  dependencies beyond the standard library.
- **AR-2.2** The parser does not depend on the `ArchitectureProfile`
  interface or the `v0/kasm/profile` sub-package. Profile selection is a
  lexer concern — by the time tokens reach the parser, architecture-
  specific classification has already occurred. Because the parser reads
  only token types and literals, it is fully architecture-agnostic.
- **AR-2.3** The parser does not depend on `debugcontext`. Diagnostic
  integration is the orchestrator's responsibility — the orchestrator
  translates `ParseError` values into `debugcontext.Entry` values. Because
  the parser returns plain data (AST + errors), it remains testable without
  the debug infrastructure.

### AR-3: AST Design Principles

- **AR-3.1** AST nodes must be value types or simple structs with public
  fields. Consumers (semantic analyser, code generator) access fields
  directly — there are no getter methods. Because the AST is a data
  structure (not a behaviour-bearing object), direct field access is
  appropriate and idiomatic in Go.
- **AR-3.2** AST nodes must not store raw `Token` values (except within
  `DirectiveStmt` argument lists, FR-3.8.2). They store extracted data
  (mnemonic string, operand kind, label name). Because the AST is a
  semantic representation — not a token mirror — it discards syntactic
  noise (commas, brackets, colons).
- **AR-3.3** The `Statement` sum type should be implemented as a Go
  interface with a marker method (e.g. `statementNode()`) or as a
  concrete struct with a `Kind` discriminator field. Both patterns are
  acceptable. The chosen pattern must be consistent across all statement
  types.
- **AR-3.4** All AST nodes must carry `Line` and `Column` fields (int) for
  source position tracking. These values are copied from the originating
  token at parse time.

---

## Non-Functional Requirements

### NFR-1: Performance

- **NFR-1.1** The parser must perform a single pass over the token slice.
  No backtracking is permitted. Because each branch of the main loop
  consumes at least one token (FR-2.1), the total work is bounded by the
  number of tokens.
- **NFR-1.2** Token lookup must be O(1) via index access into the slice.
  Because the parser uses `Position` as an index into `Tokens`, each
  `current()`, `peek()`, and `advance()` call is a constant-time array
  access.
- **NFR-1.3** The parser must not allocate intermediate data structures
  beyond the AST nodes and the error slice. Because each statement is
  appended directly to the `Program` and each error is appended directly
  to the error slice, there are no temporary buffers.

### NFR-2: Correctness

- **NFR-2.1** The parser must handle all valid token sequences without
  panicking. Because the lexer guarantees every token has a valid type
  (lexer FR-4), the parser's switch cases cover all possibilities.
- **NFR-2.2** Source positions on AST nodes must be accurate. `Line` and
  `Column` must correspond to the first token of the construct. Because
  these values are copied directly from the originating token, they are
  always consistent with the lexer's position tracking.
- **NFR-2.3** Label names in `LabelStmt` must never contain the trailing
  `:`. Because the parser strips it (FR-3.5.2), consumers never need to
  handle the colon themselves.
- **NFR-2.4** The parser must produce identical output for a given token
  slice regardless of when or how many times it is invoked (deterministic,
  side-effect-free). Because the parser has no mutable package-level state,
  the only state that changes is local to the `Parser` instance.

### NFR-3: Testability

- **NFR-3.1** The parser must be testable with only a `[]Token` slice — no
  file I/O, no profile, no debug context. Because the parser takes a plain
  slice and returns plain data, all dependencies are injectable via the
  input.
- **NFR-3.2** Tests must live in the `kasm_test` package
  (`v0/kasm/parsing_test.go`) to verify the public API surface only.
- **NFR-3.3** Test cases should construct `[]Token` slices directly (not
  via the lexer) to isolate the parser from lexer behaviour. When
  integration tests are needed, using `LexerNew(...).Start()` to produce
  the token slice is acceptable.
- **NFR-3.4** Each statement type must have dedicated test cases covering:
  - The happy path (well-formed input).
  - Missing required tokens (e.g. `namespace` without a name).
  - Empty input (zero tokens).
  - Error recovery (malformed statement followed by a valid one).

### NFR-4: Integration

- **NFR-4.1** The parser receives its input from `Lexer.Start()` (lexer
  NFR-4.2). The `[]Token` slice is passed directly to `ParserNew()`.
- **NFR-4.2** The `*Program` AST produced by `Parse()` is consumed by
  downstream stages (semantic analyser, code generator). These stages
  traverse the `Program.Statements` slice and dispatch on statement kind.
- **NFR-4.3** The `[]ParseError` slice is consumed by the orchestrator
  (`assemble_file.go`), which translates each error into a
  `debugcontext.Entry` with severity `"error"`. Because the parser returns
  `Line` and `Column` on each error, the orchestrator can construct
  accurate `debugcontext.Location` values.
- **NFR-4.4** The orchestrator must check `len(errors) > 0` after `Parse()`
  and abort the pipeline if errors are present — analogous to the
  pre-processor error check. Because the parser reports all errors (not
  just the first), the user sees the full set of issues in one pass.

### NFR-5: Extensibility

- **NFR-5.1** Adding a new statement kind (e.g. `SectionStmt`) requires:
  adding a new AST node type in `ast.go`, adding a dispatch case in the
  main parsing loop, and adding a parsing method in `parsing.go`. No
  existing statement types or the `Program` struct are modified.
- **NFR-5.2** Adding a new operand kind (e.g. `FloatOperand`) requires:
  adding a new `Operand` variant in `ast.go` and adding a case in the
  operand sub-parser. Existing operand kinds are not modified.
- **NFR-5.3** Adding a new keyword (e.g. `section`) requires: updating
  `defaultKeywords()` in the profile package (lexer NFR-5.2), adding a
  dispatch case in the keyword handler (FR-6.4), and adding the
  corresponding statement type. The parser core and existing keywords
  are not modified.

---

## Data Model

### AST Node Types

| Type               | Fields                                                       |
|--------------------|--------------------------------------------------------------|
| `Program`          | `Statements []Statement`                                     |
| `InstructionStmt`  | `Mnemonic string`, `Operands []Operand`, `Line`, `Column`    |
| `LabelStmt`        | `Name string`, `Line`, `Column`                              |
| `NamespaceStmt`    | `Name string`, `Line`, `Column`                              |
| `UseStmt`          | `ModuleName string`, `Line`, `Column`                        |
| `DirectiveStmt`    | `Literal string`, `Args []Token`, `Line`, `Column`           |

### Operand Types

| Type                | Fields                                                     |
|---------------------|------------------------------------------------------------|
| `RegisterOperand`   | `Name string`, `Line`, `Column`                            |
| `ImmediateOperand`  | `Value string`, `Line`, `Column`                           |
| `IdentifierOperand` | `Name string`, `Line`, `Column`                            |
| `StringOperand`     | `Value string`, `Line`, `Column`                           |
| `MemoryOperand`     | `Components []MemoryComponent`, `Line`, `Column`           |

### Supporting Types

| Type               | Fields                                                      |
|--------------------|-------------------------------------------------------------|
| `ParseError`       | `Message string`, `Line int`, `Column int`                  |
| `MemoryComponent`  | `Token Token` (register, immediate, identifier, or `+`/`-`)|

### Parser Struct

| Field      | Type      | Description                                        |
|------------|-----------|----------------------------------------------------|
| `Position` | `int`     | Current index into the `Tokens` slice.             |
| `Tokens`   | `[]Token` | The input token slice from the lexer.              |
| `errors`   | `[]ParseError` | Accumulated parse errors.                     |

