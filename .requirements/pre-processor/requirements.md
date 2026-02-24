# Pre-Processor

The pre-processor transforms raw `.kasm` source code before it reaches the
lexer. It runs three phases in a fixed order — includes, macros, conditionals —
each consuming the output of the previous phase. Every phase is a pure function
that takes a source string (and optionally a table) and returns a transformed
source string.

The pre-processor lives in `v0/kasm` and is orchestrated by the assembly
pipeline in `cmd/cli/cmd/x86_64/assemble_file.go`.

## Pipeline Order

```
raw source
    │
    ▼
┌──────────────────────────────┐
│ Phase 1: Includes            │  PreProcessingHandleIncludes
│   %include "file.kasm"       │
└──────────────┬───────────────┘
               │ source with files inlined
               ▼
┌──────────────────────────────┐
│ Phase 2: Macros              │  PreProcessingMacroTable
│   %macro / %endmacro         │  PreProcessingColectMacroCalls
│   macro invocations          │  PreProcessingReplaceMacroCalls
└──────────────┬───────────────┘
               │ source with macros expanded
               ▼
┌──────────────────────────────┐
│ Phase 3: Conditionals        │  PreProcessingCreateSymbolTable
│   %define / %ifdef / %ifndef │  PreProcessingHandleConditionals
│   %else / %endif             │
└──────────────┬───────────────┘
               │ final pre-processed source
               ▼
           lexer input
```

- **FR-0.1** The three phases must execute in this order. A later phase may
  depend on output produced by an earlier phase (e.g. macros can appear inside
  included files, conditionals can test for macro existence).
- **FR-0.2** Each phase receives the source string produced by the previous phase.
  Phases do not share mutable state.
- **FR-0.3** After each phase the caller snapshots the source into the
  `lineMap.Tracker` so that line-origin tracing remains accurate.

---

## Types

### PreProcessingInclusion

Represents a single `%include` directive found in the source.

```
PreProcessingInclusion {
    IncludedFilePath  string   // Path of the included file.
    LineNumber        int      // 1-based line number of the directive.
}
```

### Macro

Represents a macro definition extracted from the source.

```
Macro {
    Name        string                       // Macro name (e.g. "my_macro").
    Parameters  map[string]MacroParameter    // Parameters keyed by generated name (paramA, paramB, …).
    Body        string                       // Body text between %macro and %endmacro.
    Calls       []MacroCall                  // Invocations found in the source (populated by ColectMacroCalls).
}
```

### MacroParameter

```
MacroParameter {
    Name  string   // Generated parameter name (paramA, paramB, …).
}
```

### MacroCall

Represents a single invocation of a macro in the source.

```
MacroCall {
    Name        string     // Name of the macro being called.
    Arguments   []string   // Arguments in call order.
    LineNumber  int        // 1-based line number of the invocation.
}
```

---

## FR-1: Includes (`PreProcessingHandleIncludes`)

`PreProcessingHandleIncludes(source) → (source, []PreProcessingInclusion)`

Processes `%include` directives, replacing each with the content of the
referenced file. Returns the transformed source and a list of inclusions for
traceability.

### FR-1.1: Directive Syntax

- **FR-1.1.1** The directive syntax is `%include "path/to/file.kasm"`. Whitespace
  before `%include` and after the closing `"` is allowed.
- **FR-1.1.2** The path is extracted from between the double quotes. It is used
  as-is for `os.ReadFile` (relative to the working directory).
- **FR-1.1.3** Each directive must occupy its own line.

### FR-1.2: Validation

- **FR-1.2.1** Only `.kasm` files may be included. If the path does not end with
  `.kasm`, the function must panic with a message containing the file path and
  line number.
- **FR-1.2.2** A file may only be included once. If the same path appears in
  multiple `%include` directives, the function must panic with a message
  containing both line numbers.
- **FR-1.2.3** If `os.ReadFile` fails for an included file, the function must
  panic with a message containing the file path, line number, and the underlying
  error.

### FR-1.3: Replacement

- **FR-1.3.1** Each `%include` directive line is replaced by the content of the
  included file, wrapped in boundary comments for traceability:
  ```
  ; FILE: path/to/file.kasm
  <trimmed file content>
  ; END FILE: path/to/file.kasm
  ```
- **FR-1.3.2** The included file content must be trimmed of leading and trailing
  whitespace before insertion.
- **FR-1.3.3** Line numbers for all inclusions are computed **before** any
  replacement, so that reported line numbers refer to the original source.

### FR-1.4: Return Value

- **FR-1.4.1** The returned `[]PreProcessingInclusion` contains one entry per
  `%include` directive, each carrying the file path and the line number in the
  **original** source.
- **FR-1.4.2** If there are no `%include` directives, the source is returned
  unchanged and the slice is empty.

---

## FR-2: Macros

Macro processing has three functions that must be called in order:

1. `PreProcessingMacroTable(source)` — extract definitions.
2. `PreProcessingColectMacroCalls(source, macroTable)` — find invocations.
3. `PreProcessingReplaceMacroCalls(source, macroTable)` — expand invocations.

### FR-2.1: Macro Detection (`PreProcessingHasMacros`)

`PreProcessingHasMacros(source) → bool`

- **FR-2.1.1** Returns `true` if the source contains at least one `%macro`
  directive, `false` otherwise.
- **FR-2.1.2** Used as an early-exit check by `PreProcessingMacroTable`.

### FR-2.2: Macro Table (`PreProcessingMacroTable`)

`PreProcessingMacroTable(source) → map[string]Macro`

- **FR-2.2.1** Scans the source for `%macro <name> <paramCount>` directives
  and extracts each macro definition.
- **FR-2.2.2** The macro body is everything between the `%macro` line and the
  matching `%endmacro` line.
- **FR-2.2.3** Parameter count is parsed from the directive line. Parameters are
  generated as `paramA`, `paramB`, `paramC`, etc.
- **FR-2.2.4** If `PreProcessingHasMacros` returns `false`, an empty table is
  returned immediately.
- **FR-2.2.5** The returned `Macro.Calls` slice is initially empty — calls are
  populated by `PreProcessingColectMacroCalls`.

### FR-2.3: Macro Call Collection (`PreProcessingColectMacroCalls`)

`PreProcessingColectMacroCalls(source, macroTable)`

- **FR-2.3.1** For each macro in the table, scans the source for invocations of
  the form `<macroName> arg1, arg2, ...`.
- **FR-2.3.2** Arguments are split by comma and trimmed of whitespace.
- **FR-2.3.3** The line number of each invocation is recorded on the `MacroCall`.
- **FR-2.3.4** If the number of arguments does not match the number of parameters
  defined for the macro, the function must panic with a message containing the
  macro name, expected count, actual count, and line number.
- **FR-2.3.5** Found calls are appended to `Macro.Calls` in the macro table
  (mutates the map in place).

### FR-2.4: Macro Expansion (`PreProcessingReplaceMacroCalls`)

`PreProcessingReplaceMacroCalls(source, macroTable) → source`

- **FR-2.4.1** For each macro call, the function replaces the invocation line
  with the expanded macro body.
- **FR-2.4.2** Placeholders `%1`, `%2`, … in the macro body are replaced with
  the corresponding call arguments (1-indexed).
- **FR-2.4.3** Leading horizontal whitespace (spaces and tabs) is stripped from
  each line of the expanded body.
- **FR-2.4.4** Empty lines in the expanded body are removed.
- **FR-2.4.5** A comment `; MACRO: <name>` is prepended to the expanded body,
  surrounded by blank lines, for traceability:
  ```
  
  ; MACRO: my_macro
  mov rax, 1
  mov rdi, 2
  
  ```
- **FR-2.4.6** The macro invocation line is matched precisely, including the
  arguments, to avoid false replacements.

---

## FR-3: Symbols (`PreProcessingCreateSymbolTable`)

`PreProcessingCreateSymbolTable(source, macroTable) → map[string]bool`

Builds a symbol table for use in conditional assembly. Symbols come from two
sources: `%define` directives and macro names.

### FR-3.1: %define Directives

- **FR-3.1.1** The directive syntax is `%define SYMBOL_NAME`. Whitespace before
  `%define` and after the symbol name is allowed.
- **FR-3.1.2** The symbol name must be a non-empty valid identifier (`\w+`). An
  empty name must cause a panic with the line number.
- **FR-3.1.3** A symbol may only be defined once. Duplicate `%define` directives
  for the same symbol must cause a panic with both line numbers.

### FR-3.2: Macro Symbols

- **FR-3.2.1** Every macro name in the provided `macroTable` is added to the
  symbol table as a defined symbol (`true`).
- **FR-3.2.2** This allows `%ifdef` / `%ifndef` to test for macro existence.

### FR-3.3: Return Value

- **FR-3.3.1** The returned map keys are symbol names; all values are `true`.
- **FR-3.3.2** If there are no `%define` directives and no macros, the map is
  empty.

---

## FR-4: Conditionals (`PreProcessingHandleConditionals`)

`PreProcessingHandleConditionals(source, definedSymbols) → source`

Evaluates conditional assembly blocks (`%ifdef`, `%ifndef`, `%else`, `%endif`)
and produces a source string with only the active branches retained.

### FR-4.1: Directive Syntax

- **FR-4.1.1** `%ifdef SYMBOL` — begins a block that is included if the symbol
  is defined.
- **FR-4.1.2** `%ifndef SYMBOL` — begins a block that is included if the symbol
  is **not** defined.
- **FR-4.1.3** `%else` — optional; begins the alternative branch.
- **FR-4.1.4** `%endif` — closes the conditional block.
- **FR-4.1.5** Each directive must occupy its own line.

### FR-4.2: Nesting

- **FR-4.2.1** Conditional blocks may be nested. A `%endif` always closes the
  most recently opened `%ifdef` / `%ifndef`.

### FR-4.3: Validation

- **FR-4.3.1** A `%else` without a preceding `%ifdef` / `%ifndef` must cause a
  panic with the line number.
- **FR-4.3.2** A duplicate `%else` within the same block must cause a panic with
  the line number.
- **FR-4.3.3** A `%endif` without a matching `%ifdef` / `%ifndef` must cause a
  panic with the line number.
- **FR-4.3.4** An `%ifdef` / `%ifndef` without a matching `%endif` must cause a
  panic with the line number.

### FR-4.4: Evaluation

- **FR-4.4.1** For `%ifdef SYMBOL`: if the symbol exists in `definedSymbols`, the
  body between `%ifdef` and `%else` (or `%endif` if no `%else`) is retained. If
  the symbol does not exist, the `%else` branch (if present) is retained.
- **FR-4.4.2** For `%ifndef SYMBOL`: the logic is inverted — the body is retained
  if the symbol is **not** defined.
- **FR-4.4.3** All directive lines (`%ifdef`, `%ifndef`, `%else`, `%endif`) are
  removed from the output. Only the content of the active branch remains.
- **FR-4.4.4** The retained branch content is trimmed of leading and trailing
  whitespace.

### FR-4.5: Performance

- **FR-4.5.1** If the source is empty, it is returned immediately without
  processing.
- **FR-4.5.2** If the source does not contain `%ifdef`, `%ifndef`, or `%endif`,
  it is returned immediately without regex processing.
- **FR-4.5.3** Line numbers are precomputed in a single pass over the source
  (not per-match) to avoid O(n×m) substring scanning.
- **FR-4.5.4** The output is built using `strings.Builder` to avoid repeated
  string concatenation.

---

## FR-5: Error Reporting

All pre-processing errors are currently reported by panicking with a descriptive
message. Each panic message must include:

- **FR-5.1** The type of error (e.g. "circular inclusion", "duplicate %define",
  "wrong argument count").
- **FR-5.2** The relevant file path or symbol name.
- **FR-5.3** The line number in the source where the error was detected.
- **FR-5.4** For duplicate errors: the line number of the first occurrence.

> **Note:** The assembly pipeline (`assemble_file.go`) catches some of these
> conditions before they reach the panic (e.g. circular inclusion is detected
> by the caller). Future work should migrate all panics to `debugcontext.Error`
> recordings so the pipeline can report multiple errors instead of aborting on
> the first one.

---

## FR-6: Traceability

The pre-processor adds source-level comments so that later pipeline stages and
debugging tools can identify where content originated.

- **FR-6.1** Included file content is wrapped in `; FILE: <path>` /
  `; END FILE: <path>` comments (FR-1.3.1).
- **FR-6.2** Expanded macro bodies are prefixed with `; MACRO: <name>` comments
  (FR-2.4.5).
- **FR-6.3** The `lineMap.Tracker` receives a snapshot after each phase so that
  line-origin tracing works across all transformations (FR-0.3).
- **FR-6.4** `SnapshotWithInclusions` is used after the include phase so that
  expanding lines are annotated with their source file path.

