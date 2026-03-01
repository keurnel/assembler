# Code Generator

The code generator transforms a validated `*Program` AST (produced by the
semantic analyser) into a flat byte slice of machine code for the target
architecture. It sits at the end of the assembly pipeline — after
pre-processing, lexing, parsing, and semantic analysis — and is the component
that produces the actual binary output.

The code generator is **architecture-aware**: it receives an instruction
lookup table (variants, opcodes, encodings) at construction time and uses it
to select the correct binary encoding for each instruction statement. Because
the architecture description is injected, the same generator interface can
serve x86_64, ARM, or any future architecture for which instruction metadata
exists.

The code generator lives in `v0/kasm` and is consumed by the assembly
pipeline in `cmd/cli/cmd/x86_64/assemble_file.go`.

## Pipeline Position

```
semantic analyser output (validated *Program AST)
        │
        ▼
┌──────────────────────────────────────────────────────────────────────┐
│                        Code Generator                                │
│  GeneratorNew(program, instructions) → Generate() → ([]byte, error) │
│                                                                      │
│  ┌─────────────────────────────┐                                     │
│  │  Instruction metadata       │ ← injected at construction          │
│  │  (variants, opcodes,        │                                     │
│  │   encodings, sizes)         │                                     │
│  └─────────────────────────────┘                                     │
└──────────────────────┬───────────────────────────────────────────────┘
                       │ machine code byte slice
                       ▼
                binary output (.bin / ELF / flat binary)
```

---

## Architecture

### AR-1: File Layout

The code generator is part of the `v0/kasm` package, consistent with the
lexer, parser, and semantic analyser:

| File                    | Responsibility                                                          |
|-------------------------|-------------------------------------------------------------------------|
| `codegen.go`            | Generator construction, the `Generate()` driver, and section management.|
| `codegen_encode.go`     | Instruction encoding — opcode selection, operand encoding, REX prefix.  |
| `codegen_labels.go`     | Two-pass label resolution — collection pass and patch pass.             |
| `codegen_sections.go`   | Section handling — `.text`, `.data`, `.bss` layout and ordering.        |

- **AR-1.1** Each concern is isolated in its own file. Encoding logic must not
  leak into the label resolver, and vice versa.
- **AR-1.2** Test files mirror source files: `codegen_test.go` tests
  `codegen.go`, etc. Tests live in the `kasm_test` package (external test
  package) to test only the exported API.

### AR-2: Package Boundary

- **AR-2.1** The code generator lives in `v0/kasm`. All public code-generation
  types and functions are exported from this package.
- **AR-2.2** The code generator must not import the orchestrator
  (`cmd/cli/cmd/x86_64`), the pre-processor functions, or the lexer/parser
  constructors. It consumes only the AST types and the architecture metadata.
- **AR-2.3** The code generator may import `internal/debugcontext` for
  diagnostic recording, following the same pattern as the semantic analyser.
- **AR-2.4** The code generator must not perform file I/O. It produces a
  `[]byte` slice; the orchestrator is responsible for writing it to disk.

### AR-3: Construction Pattern

The code generator follows the same builder pattern as the lexer, parser, and
semantic analyser:

```go
generator := kasm.GeneratorNew(program, instructions).
    WithDebugContext(debugCtx)

output, errors := generator.Generate()
```

- **AR-3.1** `GeneratorNew` is the sole constructor. It is infallible — a nil
  program is treated as empty, a nil instruction table as no instructions.
- **AR-3.2** `WithDebugContext` attaches an optional `*debugcontext.DebugContext`
  for recording trace and error entries. Returns the generator for chaining.
- **AR-3.3** `Generate()` is the sole public method that drives code generation.
  It returns the encoded byte slice and a slice of `CodegenError` values.

### AR-4: Error Model

- **AR-4.1** Code generation errors are represented by a `CodegenError` type
  carrying a human-readable message, line number, and column number — matching
  the `SemanticError` / `ParseError` pattern used elsewhere in the pipeline.
- **AR-4.2** The generator must not panic. Every error condition must be recorded
  into the error slice and, if a debug context is attached, into the context.
- **AR-4.3** When an instruction cannot be encoded (no matching variant, invalid
  operand combination), the generator must record an error and skip the
  instruction rather than aborting the entire pass. This allows maximum error
  reporting in a single invocation.

---

## Functional Requirements

### FR-1: Construction

A `Generator` represents a ready-to-encode consumer of a `*Program` AST.
If a `Generator` value exists, it is guaranteed to hold a valid program
reference and initialised internal state.

- **FR-1.1** `GeneratorNew(program, instructions)` accepts the validated
  `*Program` AST and an instruction lookup table
  (`map[string]architecture.Instruction`, upper-case mnemonic keys). It
  returns a `*Generator` ready for `Generate()`.
- **FR-1.2** `GeneratorNew` is infallible. A `nil` program is treated as an
  empty program (zero statements, zero output). A `nil` instruction table is
  treated as an empty table.
- **FR-1.3** The generator must initialise an empty label table, an empty
  section map, an empty output buffer, and an empty error slice during
  construction.

### FR-2: Generation (Generate)

`Generate()` performs code generation and returns `([]byte, []CodegenError)`.

- **FR-2.1** `Generate()` executes a **two-pass** strategy:
    - **Pass 1 (collection):** Walk all statements to collect label addresses,
      section boundaries, and compute instruction sizes. No bytes are emitted.
    - **Pass 2 (emission):** Walk all statements again, encoding each
      instruction into bytes using the addresses resolved in Pass 1.
- **FR-2.2** `Generate()` must visit every statement in the program exactly
  once per pass, in source order.
- **FR-2.3** `Generate()` must return the machine code as a `[]byte` slice.
  The slice is empty when there are no encodable instructions.
- **FR-2.4** `Generate()` must return all errors accumulated during both
  passes. An empty error slice indicates successful generation.
- **FR-2.5** If a debug context is attached, `Generate()` must set the
  phase to `"codegen"` before processing and record a trace entry with the
  total number of bytes emitted upon completion.

### FR-3: Section Handling

Sections partition the program into logical regions (`.text` for code, `.data`
for initialised data, `.bss` for uninitialised data). The code generator must
respect section boundaries when laying out the binary.

- **FR-3.1** When a `SectionStmt` is encountered, the generator must switch
  to the named section. All subsequent instructions and labels belong to that
  section until the next `SectionStmt`.
- **FR-3.2** If no section is declared before the first instruction, the
  generator must assume a default `.text` section.
- **FR-3.3** The final binary must lay out sections in a deterministic order:
  `.text` first, then `.data`, then `.bss`. Within each section, content
  appears in source order.
- **FR-3.4** The `.bss` section must not emit bytes into the output — it only
  reserves space. The generator must track the `.bss` size but not append
  zero-filled bytes to the output buffer.
- **FR-3.5** Section names are case-sensitive and must start with a `.`
  (matching the lexer/parser convention).

### FR-4: Label Resolution

Labels provide symbolic names for addresses. The code generator must resolve
all label references to concrete byte offsets in the output.

- **FR-4.1** **Pass 1** collects every `LabelStmt` and records its name and
  current byte offset within its section. Duplicate labels within the same
  section must produce a `CodegenError`.
- **FR-4.2** **Pass 2** resolves `IdentifierOperand` references that match a
  known label name. The operand is replaced with the label's byte offset
  (absolute or relative, depending on the instruction encoding).
- **FR-4.3** An `IdentifierOperand` that does not match any known label must
  produce a `CodegenError` with the message "unresolved label" and the
  operand's source position.
- **FR-4.4** Forward references are supported: a label may be used before it
  is declared. Pass 1 computes all addresses before Pass 2 encodes.
- **FR-4.5** Labels must be scoped per section. A label declared in `.text`
  is not visible in `.data`, and vice versa. Cross-section label references
  must produce a `CodegenError`.

### FR-5: Instruction Encoding

The code generator translates each `InstructionStmt` into its binary
representation using the architecture's instruction metadata.

- **FR-5.1** For each `InstructionStmt`, the generator must look up the
  instruction by upper-case mnemonic in the instruction table.
- **FR-5.2** The generator must classify each operand by type (`"register"`,
  `"immediate"`, `"memory"`, `"relative"`, `"far"`) and build an operand-type
  signature.
- **FR-5.3** The generator must call `Instruction.FindVariant(operandTypes...)`
  to locate the matching `InstructionVariant`. If no variant matches, a
  `CodegenError` must be recorded.
- **FR-5.4** Once a variant is found, the generator must emit the variant's
  `Opcode` byte followed by the encoded operands, producing exactly
  `variant.Size` bytes of output per instruction.
- **FR-5.5** Register operands must be encoded using a register-number lookup
  table specific to the target architecture. For x86_64, the standard 64-bit
  register encoding applies:
  ```
  RAX=0, RCX=1, RDX=2, RBX=3, RSP=4, RBP=5, RSI=6, RDI=7,
  R8=8, R9=9, R10=10, R11=11, R12=12, R13=13, R14=14, R15=15
  ```
- **FR-5.6** Immediate operands must be parsed from their string
  representation into integer values. Supported formats:
    - Decimal: `42`, `-1`
    - Hexadecimal: `0xFF`, `0x1A`
    - Binary: `0b1010`
  An unparseable immediate must produce a `CodegenError`.
- **FR-5.7** Memory operands (bracket expressions) must be encoded according
  to the x86_64 ModR/M and SIB byte conventions. The generator must support
  at minimum:
    - `[register]` — register-indirect addressing
    - `[register + immediate]` — base + displacement
    - `[register + register]` — base + index (SIB)

### FR-6: REX Prefix (x86_64)

The REX prefix is required when 64-bit registers or extended registers
(R8–R15) are used.

- **FR-6.1** The generator must emit a REX prefix (`0x40`–`0x4F`) when any
  operand references a 64-bit register.
- **FR-6.2** The REX.W bit (bit 3) must be set for 64-bit operand size.
- **FR-6.3** The REX.R bit (bit 2) must be set when the ModR/M `reg` field
  encodes an extended register (R8–R15).
- **FR-6.4** The REX.B bit (bit 0) must be set when the ModR/M `r/m` field
  or the opcode register field encodes an extended register (R8–R15).
- **FR-6.5** The REX prefix must be emitted immediately before the opcode
  byte. The variant's `Size` field accounts for the REX prefix when
  applicable.

### FR-7: Output Format

- **FR-7.1** The default output is a **flat binary** — a raw byte slice with
  no headers, relocations, or metadata. This is suitable for bootloaders and
  bare-metal programs loaded at a known address.
- **FR-7.2** The orchestrator is responsible for writing the `[]byte` to disk
  with the appropriate file extension (`.bin`).
- **FR-7.3** Future output formats (ELF, PE, Mach-O) are out of scope for
  this version. The generator interface (`[]byte` return) is compatible with
  wrapping in a format-specific emitter later.

### FR-8: Diagnostics & Debug Context Integration

- **FR-8.1** When a `*debugcontext.DebugContext` is attached, the generator
  must set the phase to `"codegen"` before processing begins.
- **FR-8.2** Each encoding error must be recorded via `debugCtx.Error()` with
  the source location of the offending statement.
- **FR-8.3** Upon successful completion, the generator must emit a trace entry
  via `debugCtx.Trace()` summarising the result:
  `"code generation complete: N byte(s) emitted across M section(s)"`.
- **FR-8.4** When verbose mode is enabled, the generator should emit trace
  entries for each encoded instruction, including the mnemonic, the selected
  variant encoding, and the emitted bytes (hex-formatted).

### FR-9: Orchestrator Integration

The code generator is wired into the assembly pipeline by the orchestrator
in `assemble_file.go`, following the same pattern as the lexer, parser, and
semantic analyser.

- **FR-9.1** The orchestrator must call `GeneratorNew(program, instrTable)`
  after semantic analysis succeeds (zero semantic errors).
- **FR-9.2** The orchestrator must attach the same `DebugContext` used by
  earlier pipeline stages via `WithDebugContext(debugCtx)`.
- **FR-9.3** The orchestrator must call `Generate()` and inspect the returned
  error slice. If errors are present, the orchestrator must print them and
  abort with a non-zero exit code.
- **FR-9.4** On success, the orchestrator must write the `[]byte` output to
  the target file. The output file name defaults to the input file name with
  the extension replaced by `.bin` (e.g. `main.kasm` → `main.bin`).
- **FR-9.5** The orchestrator must print debug context entries when verbose
  mode (`-v`) is enabled, consistent with all other pipeline stages.

---

## Types

### CodegenError

Represents a single error encountered during code generation.

```
CodegenError {
    Message string   // Human-readable description of the error.
    Line    int      // 1-based line number of the offending statement.
    Column  int      // 1-based column number.
}
```

### Generator

The code generator struct. Internal fields are unexported.

```
Generator {
    program      *Program
    instructions map[string]architecture.Instruction
    labels       map[string]labelEntry
    sections     map[string]*sectionBuffer
    current      string                          // current section name
    errors       []CodegenError
    debugCtx     *debugcontext.DebugContext
}
```

### labelEntry (unexported)

Tracks a label's resolved address within a section.

```
labelEntry {
    name    string
    section string
    offset  int
    line    int
    column  int
}
```

### sectionBuffer (unexported)

Accumulates bytes for a single section.

```
sectionBuffer {
    name   string
    data   []byte
    size   int      // for .bss, size without data
}
```

