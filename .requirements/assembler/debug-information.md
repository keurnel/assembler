# Debug Context

The debug context is a passive data structure that accumulates contextual
information as the assembler pipeline progresses. Any component in the pipeline
(pre-processor, lexer, parser, code generator) can record entries into the
context. When an error, warning, or diagnostic needs to be reported to the
user, the context provides all the information required to produce a
meaningful, human-readable message — including the original source file, the
original line number, the pipeline phase that produced the issue, and a
relevant source snippet.

The debug context does **not** perform I/O, formatting, or reporting itself.
It is a structured log of what happened and where. A separate renderer
consumes the context to produce terminal output, log files, or IDE-compatible
diagnostics.

## Concepts

```
┌─────────────┐      ┌──────────────┐      ┌──────────────┐
│ Pre-process  │─────▶│    Lexer      │─────▶│    Parser     │──▶ ...
└─────┬───────┘      └──────┬───────┘      └──────┬───────┘
      │ record              │ record              │ record
      ▼                     ▼                     ▼
   ┌─────────────────────────────────────────────────────┐
   │                   DebugContext                       │
   │  entries: []Entry                                    │
   │  phases:  current pipeline phase                     │
   └─────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌───────────────┐
                  │   Renderer    │  (terminal, JSON, LSP)
                  └───────────────┘
```

- **Entry**: a single recorded event (error, warning, info, trace).
- **Phase**: the pipeline stage that produced the entry (e.g. `"pre-processing"`,
  `"lexing"`, `"parsing"`, `"codegen"`).
- **Location**: the position in source code the entry refers to — file path, line
  number, column number.

---

## FR-1: Construction

- **FR-1.1** `NewDebugContext(filePath)` is the sole constructor. It returns a
  `*DebugContext` initialised with the primary source file path, an empty entry
  list, and the phase set to `""` (no phase).
- **FR-1.2** The primary file path is stored once and used as the default file for
  entries that do not specify a different file (e.g. lines from the main source as
  opposed to lines from an included file).

## FR-2: Phases

A phase is a label that identifies which pipeline stage is currently active.
The phase is attached to every entry recorded while it is set.

- **FR-2.1** `SetPhase(name)` sets the current phase. Subsequent entries are tagged
  with this phase until it is changed again.
- **FR-2.2** Phase names are free-form strings. Recommended values:
  - `"pre-processing/includes"` — handling `%include` directives.
  - `"pre-processing/macros"` — macro expansion.
  - `"pre-processing/conditionals"` — `%ifdef` / `%ifndef` evaluation.
  - `"lexing"` — tokenisation.
  - `"parsing"` — syntax analysis.
  - `"codegen"` — machine code generation.
- **FR-2.3** `Phase()` returns the current phase name.

## FR-3: Recording Entries

Entries are the core data of the context. Each entry captures **what**
happened, **where** it happened, and **how severe** it is.

### FR-3.1: Severity Levels

Each entry has exactly one severity:

| Severity  | Meaning |
|-----------|---------|
| `error`   | The assembly cannot continue or produce correct output. |
| `warning` | The assembly can continue, but the result may be unexpected. |
| `info`    | Informational note for the user (e.g. "macro X expanded here"). |
| `trace`   | Internal diagnostic, hidden by default, useful for debugging the assembler itself. |

### FR-3.2: Entry Fields

Every entry must carry the following fields:

- **FR-3.2.1** `severity` — one of the four severity levels.
- **FR-3.2.2** `phase` — the pipeline phase that was active when the entry was
  recorded. Set automatically from the current phase.
- **FR-3.2.3** `message` — a human-readable description of the event.
- **FR-3.2.4** `location` — the source position the entry refers to:
  - `filePath` — the file the line belongs to. Uses the primary file path by
    default; overridden for lines from included files.
  - `line` — 1-based line number in the original (pre-processed) source.
  - `column` — 1-based column number, or `0` if not applicable.
- **FR-3.2.5** `snippet` — optional: the actual source line text for inline display
  in diagnostics. Empty string if not provided.
- **FR-3.2.6** `hint` — optional: a suggestion for the user on how to fix the issue
  (e.g. `"did you mean 'mov'?"`). Empty string if not provided.

### FR-3.3: Recording Methods

- **FR-3.3.1** `Error(location, message)` — records an entry with severity `error`.
- **FR-3.3.2** `Warning(location, message)` — records an entry with severity `warning`.
- **FR-3.3.3** `Info(location, message)` — records an entry with severity `info`.
- **FR-3.3.4** `Trace(location, message)` — records an entry with severity `trace`.
- **FR-3.3.5** Each method returns the `*Entry` it just created so the caller can
  attach optional fields:
  ```
  ctx.Error(loc, "unknown instruction").WithSnippet("  mvo rax, 1").WithHint("did you mean 'mov'?")
  ```
- **FR-3.3.6** All recording methods are safe to call from any goroutine. The context
  must be safe for concurrent writes.

## FR-4: Location

`Location` is a value type that identifies a position in source code.

- **FR-4.1** `Loc(line, column)` creates a `Location` using the primary file path
  from the context.
- **FR-4.2** `LocIn(filePath, line, column)` creates a `Location` with an explicit
  file path (used for lines originating from included files).
- **FR-4.3** `column` may be `0` to indicate "the entire line" (e.g. a pre-processing
  directive that applies to the full line).

## FR-5: Querying Entries

The context must support filtering entries so that renderers and callers can
select what to display.

- **FR-5.1** `Entries()` returns all recorded entries in insertion order.
- **FR-5.2** `Errors()` returns only entries with severity `error`.
- **FR-5.3** `Warnings()` returns only entries with severity `warning`.
- **FR-5.4** `HasErrors()` returns `true` if at least one `error` entry exists.
  This is the primary check used to decide whether the pipeline should abort.
- **FR-5.5** `Count()` returns the total number of entries.

## FR-6: Entry Type

```
Entry {
    severity  string    // "error" | "warning" | "info" | "trace"
    phase     string    // pipeline phase at recording time
    message   string    // human-readable description
    location  Location  // file, line, column
    snippet   string    // optional source line text
    hint      string    // optional fix suggestion
}
```

- **FR-6.1** `WithSnippet(text)` sets the snippet and returns the same `*Entry`
  for chaining.
- **FR-6.2** `WithHint(text)` sets the hint and returns the same `*Entry` for
  chaining.
- **FR-6.3** `String()` returns a single-line human-readable representation for
  quick debugging:
  ```
  error [pre-processing/includes] main.kasm:12:0: unknown file 'missing.kasm'
  ```

## FR-7: Location Type

```
Location {
    filePath  string  // absolute or relative path to the source file
    line      int     // 1-based line number
    column    int     // 1-based column number, or 0 for "entire line"
}
```

- **FR-7.1** `String()` returns `filePath:line:column` (e.g. `"main.kasm:12:5"`).
  If column is `0`, it returns `filePath:line` (e.g. `"main.kasm:12"`).

## FR-8: Design Constraints

- **FR-8.1** The debug context is a **passive data structure**. It does not print,
  format, or output anything. Rendering is a separate concern.
- **FR-8.2** The debug context is **append-only**. Entries cannot be modified or
  removed after recording.
- **FR-8.3** The debug context must be **thread-safe**. Multiple pipeline stages may
  record concurrently (e.g. parallel macro expansion).
- **FR-8.4** The debug context does **not** own or reference the `lineMap.Tracker`.
  The caller is responsible for resolving a processed-source line number back to
  its original location (via `tracker.Origin()`) before recording an entry. The
  context only stores resolved, original locations.
- **FR-8.5** The debug context is created once per assembly invocation and passed
  through the pipeline by reference.

## FR-9: Integration with the Pipeline

The debug context is threaded through every stage of the assembly pipeline.
Each stage sets its phase and records entries as issues are encountered.

- **FR-9.1** `runAssembleFile` creates the `DebugContext` and passes it to each
  pipeline function.
- **FR-9.2** Pre-processing functions record entries for issues such as:
  - File not found for `%include`.
  - Circular inclusion detected.
  - Undefined symbol in `%ifdef`.
  - Malformed macro definition.
- **FR-9.3** The lexer records entries for issues such as:
  - Unexpected character.
  - Unterminated string literal.
- **FR-9.4** The parser records entries for issues such as:
  - Unknown instruction mnemonic.
  - Wrong number of operands.
  - Invalid operand type for instruction.
- **FR-9.5** After the pipeline completes (or aborts), the caller checks
  `ctx.HasErrors()` to decide whether to produce output or report failures.
- **FR-9.6** A renderer (out of scope for this document) consumes `ctx.Entries()`
  and formats them for the user's terminal.
