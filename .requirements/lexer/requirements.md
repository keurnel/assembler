# Lexer

The lexer (tokeniser) transforms a pre-processed `.kasm` source string into an
ordered sequence of tokens. Each token carries a type, literal value, and source
location. The lexer sits between the pre-processor and the parser in the
assembly pipeline.

The lexer is **architecture-agnostic**: it does not hardcode any register names,
instruction mnemonics, or keywords. Instead, it receives an
`ArchitectureProfile` at construction time that supplies these sets for the
target architecture. Because the profile is the sole source of
architecture-specific vocabulary, the same lexer tokenises source code for
x86_64, ARM, RISC-V, or any future architecture without modification.

The lexer lives in `v0/internal/kasm` and is consumed by the parser
(`ParserNew`) and the assembly pipeline in
`cmd/cli/cmd/x86_64/assemble_file.go`.

## Pipeline Position

```
pre-processed source
        │
        ▼
┌──────────────────────────────────────────────┐
│              Lexer                            │
│  LexerNew(input, profile) → Start() → []Token│
│                                              │
│  ┌────────────────────┐                      │
│  │ ArchitectureProfile│ ← injected at        │
│  │  · Registers()     │   construction       │
│  │  · Instructions()  │                      │
│  │  · Keywords()      │                      │
│  └────────────────────┘                      │
└──────────────────────┬───────────────────────┘
                       │ ordered token slice
                       ▼
                   parser input
```

---

## Functional Requirements

### FR-1: Architecture Profile

An `ArchitectureProfile` represents a validated, immutable vocabulary for a
specific hardware architecture. If an `ArchitectureProfile` value exists, it
is guaranteed to hold three non-nil maps — registers, instructions, and
keywords — all keyed by lower-case strings. There is no partially-initialised
or mutable state.

#### FR-1.1: Interface

```go
type ArchitectureProfile interface {
    // Registers returns the set of recognised register names (lower-case).
    Registers() map[string]bool
    // Instructions returns the set of recognised instruction mnemonics (lower-case).
    Instructions() map[string]bool
    // Keywords returns the set of reserved language keywords (lower-case).
    Keywords() map[string]bool
}
```

- **FR-1.1.1** `Registers()` must return a `map[string]bool` of recognised
  register names, all in lower-case. The lexer performs a case-insensitive
  lookup by lower-casing the word before checking the map. Because the map is
  pre-built at profile construction time, each lookup is O(1).
- **FR-1.1.2** `Instructions()` must return a `map[string]bool` of recognised
  instruction mnemonics, all in lower-case. The same O(1) lookup applies.
- **FR-1.1.3** `Keywords()` must return a `map[string]bool` of reserved
  language keywords, all in lower-case (e.g. `namespace`).
- **FR-1.1.4** All three methods must return non-nil maps. An empty map is
  valid — the architecture simply has no entries of that kind. Because the
  maps are guaranteed non-nil, the lexer never needs a nil guard before
  performing a lookup.
- **FR-1.1.5** The profile must not be modified after construction. The lexer
  stores the reference — mutations would corrupt classification. Because the
  profile is immutable, it is safe for concurrent use: multiple lexer
  instances may share the same profile without synchronisation (NFR-1.4).

#### FR-1.2: Built-in Profiles

The package must ship at least one concrete profile for x86_64 that can be
used directly or serve as a reference for other architectures.

- **FR-1.2.1** `NewX8664Profile()` returns an `ArchitectureProfile` populated
  with the x86_64 register set (FR-5), instruction set (FR-6), and the
  default keyword set (FR-7). Because all three sets are assembled at
  construction time, the returned profile is immediately ready for use —
  there is no separate initialisation step.
- **FR-1.2.2** Additional profiles (e.g. `NewARM64Profile()`,
  `NewRISCVProfile()`) may be added in future without changing the lexer
  itself. Because the lexer depends only on the `ArchitectureProfile`
  interface (AR-2.1), adding a profile is a purely additive change.
- **FR-1.2.3** Profiles should be constructable from the existing
  `v0/architecture` package. A helper `ProfileFromArchitecture(groups
  map[string]architecture.InstructionGroup, registers map[string]bool,
  keywords []string)` may be provided to bridge the architecture package to
  the lexer. Because this helper lower-cases all mnemonics and merges the
  default keyword set, callers do not need to normalise data themselves.

#### FR-1.3: Integration with `v0/architecture`

- **FR-1.3.1** The x86_64 profile must derive its instruction set from the
  `v0/architecture/x86/_64.Instructions()` providers. All mnemonics returned
  by all providers must appear in the profile's `Instructions()` map
  (lower-cased). Because the profile is the single source of truth for the
  lexer, adding a new `InstructionProvider` to the architecture package has
  no effect on the lexer until the profile is updated.
- **FR-1.3.2** The x86_64 profile must include all registers listed in FR-5.
  These may be defined as a static set within the profile, or loaded from a
  register provider if one exists.
- **FR-1.3.3** The lexer never queries the architecture package directly.
  Because all vocabulary flows through the profile (FR-1.1), there is no
  coupling between the lexer core and any specific architecture package.

### FR-2: Construction

A `Lexer` represents a ready-to-tokenise scanner. If a `Lexer` value exists,
it is guaranteed to hold a valid input string, a valid profile, initialised
position state, and an empty token slice. There is no uninitialised or
partially-constructed state.

- **FR-2.1** `LexerNew(input, profile)` is the sole constructor. It accepts
  the pre-processed source string and an `ArchitectureProfile`, and returns a
  `*Lexer` that is ready for `Start()` to be called. There is no separate
  `Init()` or `SetProfile()` step.
- **FR-2.2** `LexerNew` is infallible — it cannot fail. Any valid string
  (including the empty string) is accepted. The profile must not be `nil`;
  passing `nil` may panic. This is a programming error, not a runtime error
  — because the orchestrator constructs the profile before the lexer (NFR-4.4),
  a `nil` profile indicates a bug in the orchestrator, not bad user input.
- **FR-2.3** `LexerNew` must initialise the lexer to the first character of
  the input by calling `readChar()` during construction. After construction,
  `Ch` holds the first character (or `0` / NUL if the input is empty).
  Because `readChar()` is called at construction, `Position` and
  `ReadPosition` are already advanced — the lexer is positioned at the first
  character, not before it.
- **FR-2.4** `Position` and `ReadPosition` must start at `0`. `Line` must
  start at `1`. `Column` must start at `0` (it is incremented to `1` when
  the first character is read by `readChar()` during construction).
- **FR-2.5** The `Tokens` slice must be initialised as an empty, non-nil
  slice. Because it is pre-allocated, `addToken()` can append without a nil
  check.
- **FR-2.6** The profile reference must be stored on the `Lexer` and used by
  `classifyWord()` for all register, instruction, and keyword lookups.
  Because the profile is immutable (FR-1.1.5), storing a reference is safe —
  the lexer does not need to copy the maps.

### FR-3: Tokenisation (Start)

`Start()` performs a single-pass, left-to-right scan of the input and returns
an ordered slice of tokens. It is the sole public method that drives
tokenisation. Because the `Lexer` is guaranteed fully initialised after
`LexerNew` (FR-2), `Start()` has no precondition checks or error paths.

- **FR-3.1** `Start()` must consume the entire input, stopping when `Ch`
  equals `0` (NUL / end of input). Because `readChar()` sets `Ch` to `0`
  when `ReadPosition` exceeds the input length (FR-10.1), the loop
  termination condition is always reached.
- **FR-3.2** `Start()` must return a `[]Token` slice containing all emitted
  tokens in the order they appear in the source. Because tokens are appended
  sequentially via `addToken()`, ordering is preserved by construction.
- **FR-3.3** `Start()` is infallible — it cannot fail or panic on any input.
  Malformed constructs (e.g. unterminated strings) must be handled gracefully
  by consuming until EOF. Because every character is consumed by exactly one
  branch of the main switch (FR-4), the scanner always makes progress —
  infinite loops are impossible.
- **FR-3.4** `Start()` may be called only once per `Lexer` instance. Calling
  it again would re-scan from the current (exhausted) position and return an
  empty slice. The lexer does not reset internal state.
- **FR-3.5** Each emitted token must carry accurate `Line` and `Column` values
  reflecting its starting position in the source. Because `Line` and `Column`
  are captured before each token's characters are consumed, the values
  correspond to the first character of the token.

### FR-4: Token Types

Every token emitted by the lexer is classified into exactly one of the
following types. The type is determined by the character context at the point
of consumption, combined with the `ArchitectureProfile` lookup tables for
known words. Because every character is handled by exactly one branch and
each branch either emits a token or skips content, no character is silently
dropped or double-counted.

| Type               | Constant          | Description                                                   |
|--------------------|-------------------|---------------------------------------------------------------|
| Whitespace         | `TokenWhitespace` | Sequence of spaces, tabs, `\r`, `\n`. **Never emitted.**     |
| Comment            | `TokenComment`    | `;` to end of line. **Never emitted.**                        |
| Directive          | `TokenDirective`  | `%`-prefixed word (e.g. `%define`, `%include`, `%endif`).     |
| Instruction        | `TokenInstruction`| Known mnemonic from profile (e.g. `mov`, `add`, `syscall`).   |
| Register           | `TokenRegister`   | Known register name from profile (e.g. `rax`, `x0`, `a0`).   |
| Immediate          | `TokenImmediate`  | Decimal (`42`) or hexadecimal (`0xFF`) numeric literal.       |
| String             | `TokenString`     | `"…"` delimited string literal. The quotes are not stored.    |
| Keyword            | `TokenKeyword`    | Reserved keyword from profile (e.g. `namespace`).             |
| Identifier         | `TokenIdentifier` | Any other word, label (`_start:`), or single punctuation.     |

#### FR-4.1: Whitespace

- **FR-4.1.1** Whitespace characters (`' '`, `'\t'`, `'\r'`, `'\n'`) must be
  consumed and skipped. No `TokenWhitespace` token is emitted. Because
  whitespace is skipped, consecutive whitespace never produces tokens —
  the parser receives a clean, whitespace-free stream.
- **FR-4.1.2** Consecutive whitespace characters must be consumed as a single
  unit (not one token per character). Because `skipWhitespace()` loops until
  a non-whitespace character is found, a single invocation handles any run.

#### FR-4.2: Comments

- **FR-4.2.1** A `;` character starts a comment that extends to the end of the
  current line (or end of input).
- **FR-4.2.2** Comments must be consumed and skipped. No `TokenComment` token
  is emitted. Because both whitespace (FR-4.1) and comments are skipped, the
  token slice contains only semantically meaningful tokens.
- **FR-4.2.3** The comment includes the leading `;`. Because `skipComment()`
  starts at the `;` and advances to `'\n'` or NUL, the entire comment is
  consumed in one pass.

#### FR-4.3: Directives

- **FR-4.3.1** A `%` character followed by word characters must be read as a
  single directive token.
- **FR-4.3.2** The literal value must include the `%` prefix (e.g. `%define`).
  Because `readDirective()` starts at the `%` and captures through the
  trailing word characters, the prefix is always included.
- **FR-4.3.3** The token type must be `TokenDirective`. Directive
  classification is determined by the leading `%` character, not by profile
  lookup — directives are architecture-independent.

#### FR-4.4: String Literals

- **FR-4.4.1** A `"` character starts a string literal. The lexer must consume
  characters until the closing `"` or end of input.
- **FR-4.4.2** The literal value must contain only the content between the
  quotes — the delimiting `"` characters are not included. Because
  `readString()` skips the opening `"` before capturing and skips the
  closing `"` after capturing, delimiters are never part of the literal.
- **FR-4.4.3** An unterminated string (no closing `"` before EOF) must not
  cause a panic. The lexer must consume until EOF and emit a `TokenString`
  with whatever content was found. Because `readString()` checks for both
  `'"'` and `0` (NUL) in its loop condition, EOF terminates the read
  gracefully.
- **FR-4.4.4** An empty string `""` must emit a `TokenString` with an empty
  literal. Because the opening and closing `"` are immediately adjacent, the
  captured slice is empty — this is a valid state, not an error.

#### FR-4.5: Numeric Literals (Immediates)

- **FR-4.5.1** A digit (`0`–`9`) starts a numeric literal. Because digits
  cannot begin a word (FR-4.6.1 requires a letter, `_`, or `.`), there is
  no ambiguity between numbers and words.
- **FR-4.5.2** Decimal literals: consecutive digits are consumed.
- **FR-4.5.3** Hexadecimal literals: a `0x` or `0X` prefix followed by hex
  digits (`0`–`9`, `a`–`f`, `A`–`F`) are consumed as a single token.
  Because `readNumber()` checks `peekChar()` for `'x'` or `'X'` after
  reading `'0'`, the prefix detection does not advance past a standalone
  `0`.
- **FR-4.5.4** The literal value preserves the original casing (e.g. `0xFF`,
  `0XAB`). Because `readNumber()` uses slice indexing
  (`Input[start:Position]`), the source text is captured verbatim.
- **FR-4.5.5** The token type must be `TokenImmediate`.

#### FR-4.6: Words (Instructions, Registers, Keywords, Identifiers)

A word is a contiguous sequence of letters, digits, underscores (`_`), and
dots (`.`). Words are classified using the `ArchitectureProfile` and context.
Because the profile supplies the vocabulary (FR-1), the lexer core has no
hardcoded knowledge of any specific register or instruction name.

- **FR-4.6.1** A word starting with a letter, `_`, or `.` must be consumed via
  `readWord()`.
- **FR-4.6.2** If a `:` immediately follows the word, it is consumed and
  appended to the literal (forming a label, e.g. `_start:`). The token is
  always classified as `TokenIdentifier`. Because the `:` is consumed before
  `classifyWord()` is called, labels are structurally prevented from being
  classified as instructions or registers.
- **FR-4.6.3** Classification is performed by `classifyWord()` using
  case-insensitive lookup against the profile's register, instruction, and
  keyword maps. The original casing is preserved in the literal. Because
  the maps store lower-case keys (FR-1.1.1, FR-1.1.2, FR-1.1.3) and the
  lexer lower-cases the word before lookup, classification is always
  case-insensitive.
- **FR-4.6.4** If the lower-cased word matches `profile.Registers()`, the
  type is `TokenRegister`.
- **FR-4.6.5** If the lower-cased word matches `profile.Instructions()`, the
  type is `TokenInstruction`.
- **FR-4.6.6** If the lower-cased word matches `profile.Keywords()`, the type
  is `TokenKeyword`.
- **FR-4.6.7** Otherwise the type is `TokenIdentifier`. Because the fallback
  is always `TokenIdentifier`, every word receives a classification — no
  word is ever dropped or left untyped.
- **FR-4.6.8** When the previous token is a `TokenKeyword`, the current word
  must be classified as `TokenIdentifier` regardless of lookup results
  (FR-11.2). This rule takes precedence over FR-4.6.4 through FR-4.6.6.
  Because it is checked first, keywords can introduce arbitrary names
  (e.g. `namespace mov` → keyword `namespace` + identifier `mov`) without
  the name being misclassified as an instruction.

#### FR-4.7: Punctuation and Other Characters

- **FR-4.7.1** Any character that does not match the above categories (e.g.
  `,`, `[`, `]`, `+`, `-`) must be emitted as a single-character
  `TokenIdentifier`. Because this is the `default` branch of the main
  switch, it is impossible for a character to be silently consumed without
  producing a token — every non-whitespace, non-comment character either
  matches a specific rule or falls through to punctuation.
- **FR-4.7.2** The lexer must advance past the character after emitting the
  token. Because `readChar()` is called after the token is emitted,
  progress is guaranteed and the scanner cannot stall.

### FR-5: x86_64 Register Set

The x86_64 profile maintains the following register names. All entries are
lower-case in the lookup table; classification is case-insensitive (FR-4.6.3).
Because these entries are assembled at profile construction time
(FR-1.2.1), they are immutable and do not change between lexer invocations.

- **FR-5.1** 64-bit general-purpose registers: `rax`, `rbx`, `rcx`, `rdx`,
  `rsi`, `rdi`, `rbp`, `rsp`, `r8`–`r15`.
- **FR-5.2** 32-bit general-purpose registers: `eax`, `ebx`, `ecx`, `edx`,
  `esi`, `edi`, `ebp`, `esp`, `r8d`–`r15d`.
- **FR-5.3** 16-bit general-purpose registers: `ax`, `bx`, `cx`, `dx`, `si`,
  `di`, `bp`, `sp`.
- **FR-5.4** 8-bit registers: `al`, `bl`, `cl`, `dl`, `ah`, `bh`, `ch`,
  `dh`, `sil`, `dil`, `bpl`, `spl`.
- **FR-5.5** Segment registers: `cs`, `ds`, `es`, `fs`, `gs`, `ss`.
- **FR-5.6** Instruction pointer and flags: `rip`, `eip`, `rflags`, `eflags`.

### FR-6: x86_64 Instruction Set

The x86_64 profile maintains the following instruction mnemonics. All entries
are lower-case in the lookup table; classification is case-insensitive
(FR-4.6.3). These mnemonics must match those provided by
`v0/architecture/x86/_64` providers (FR-1.3.1). Because the profile is the
single source of truth, adding or removing a mnemonic from the architecture
package has no effect on the lexer until the profile is updated.

- **FR-6.1** Data transfer: `mov`, `movzx`, `movsx`, `lea`, `push`, `pop`,
  `xchg`.
- **FR-6.2** Arithmetic: `add`, `sub`, `mul`, `imul`, `div`, `idiv`, `inc`,
  `dec`, `neg`.
- **FR-6.3** Bitwise / shift: `and`, `or`, `xor`, `not`, `shl`, `shr`, `sal`,
  `sar`, `rol`, `ror`.
- **FR-6.4** Comparison: `cmp`, `test`.
- **FR-6.5** Control flow: `jmp`, `je`, `jne`, `jz`, `jnz`, `jg`, `jge`,
  `jl`, `jle`, `ja`, `jae`, `jb`, `jbe`, `call`, `ret`, `syscall`, `int`.
- **FR-6.6** System / misc: `nop`, `hlt`, `cli`, `sti`.
- **FR-6.7** Loop: `loop`, `loope`, `loopne`.
- **FR-6.8** Conditional move: `cmove`, `cmovne`, `cmovg`, `cmovl`.
- **FR-6.9** Set byte: `sete`, `setne`, `setg`, `setl`.
- **FR-6.10** String / repeat: `rep`, `movsb`, `stosb`.
- **FR-6.11** Sign extension: `cbw`, `cwd`, `cdq`, `cqo`.
- **FR-6.12** Custom: `use` (module import instruction).

### FR-7: Default Keyword Set

Keywords are language-level reserved words that are architecture-independent.
They are part of the `.kasm` language, not a specific CPU. Because keywords
are language-level, they must be present in every profile — regardless of
hardware architecture.

- **FR-7.1** The default keyword set contains: `namespace`.
- **FR-7.2** Keywords are shared across all architecture profiles. Each
  built-in profile constructor (e.g. `NewX8664Profile()`) must include the
  default keyword set. Because the keyword set is merged at construction
  time, profile constructors do not need to know about each other.
- **FR-7.3** A helper `defaultKeywords() map[string]bool` must exist so that
  profile constructors do not duplicate the keyword list. Because the helper
  returns a fresh map each time (not a shared reference), callers may extend
  it with profile-specific keywords without affecting other profiles.

### FR-8: Token Structure

Each token produced by the lexer is a value type carrying four fields.
Because `Token` is a value type (not a pointer), tokens are safe to copy,
compare, and store without aliasing concerns.

- **FR-8.1** `Type` (`TokenType`) — the classification of the token, as
  described in FR-4. Because every branch of the scanner assigns a type
  (FR-4), no token can have an uninitialised type.
- **FR-8.2** `Literal` (`string`) — the verbatim text from the source code
  that produced the token. For string literals this is the content between
  the quotes (without the `"` delimiters). Because `readWord()`,
  `readNumber()`, and `readDirective()` all use slice indexing
  (`Input[start:Position]`), the literal is always a direct substring of
  the input — no allocations beyond the slice header.
- **FR-8.3** `Line` (`int`) — the 1-based line number where the token starts.
  Because `Line` is captured before the token's characters are consumed
  (FR-3.5), it reflects the first character of the token.
- **FR-8.4** `Column` (`int`) — the 1-based column number where the token
  starts. Same capture semantics as `Line`.

### FR-9: Token Type Methods

`TokenType` exposes convenience methods for classification queries. These
methods eliminate raw integer comparisons in consuming code. Because each
method maps to exactly one (or two) constants, the intent is always explicit.

- **FR-9.1** `ToInt()` — returns the underlying integer value.
- **FR-9.2** `Ignored()` — returns `true` for `TokenWhitespace` and
  `TokenComment`. Used by the parser to skip non-semantic tokens. Because
  these two types are never emitted by `Start()` (FR-4.1, FR-4.2), this
  method is relevant only if tokens are constructed manually (e.g. in
  tests).
- **FR-9.3** `Whitespace()` — returns `true` only for `TokenWhitespace`.
- **FR-9.4** `Comment()` — returns `true` only for `TokenComment`.
- **FR-9.5** `Identifier()` — returns `true` only for `TokenIdentifier`.
- **FR-9.6** `Directive()` — returns `true` only for `TokenDirective`.
- **FR-9.7** `Instruction()` — returns `true` only for `TokenInstruction`.
- **FR-9.8** `Register()` — returns `true` only for `TokenRegister`.
- **FR-9.9** `Immediate()` — returns `true` only for `TokenImmediate`.
- **FR-9.10** `StringLiteral()` — returns `true` only for `TokenString`.

### FR-10: Character Reading

The lexer advances through the input one byte at a time, maintaining accurate
position state. Because every scanning method delegates to `readChar()` for
advancement, position tracking is centralised — there is no alternative path
that could desynchronise `Line`, `Column`, or `Position`.

- **FR-10.1** `readChar()` advances `Position` and `ReadPosition` by one. If
  `ReadPosition` is beyond the input length, `Ch` is set to `0` (NUL).
  Because NUL is the loop termination signal for `Start()` (FR-3.1),
  reaching end-of-input always terminates the scanner.
- **FR-10.2** When `Ch` is `'\n'`, `Line` must increment by one and `Column`
  must reset to `0`. For all other characters `Column` must increment by one.
  Because this logic runs on every character, line and column counters are
  always accurate — no special-case tracking is needed for multi-line tokens.
- **FR-10.3** `peekChar()` returns the next character without advancing any
  state. Returns `0` if at end of input. Because it does not modify
  `Position`, `ReadPosition`, `Line`, or `Column`, it is safe to call
  at any point without side effects.

### FR-11: Context-Sensitive Classification

Some classification decisions depend on the previously emitted token. Because
context-sensitivity is limited to a single lookback (the previous token), the
lexer remains a simple state machine — there is no stack, no nesting, and no
multi-token lookahead.

- **FR-11.1** `previousTokenType()` must return the type of the most recently
  emitted token, or `-1` if no tokens have been emitted yet. Because the
  `Tokens` slice is initialised as non-nil and empty (FR-2.5), a length
  check is sufficient — no nil guard is needed.
- **FR-11.2** When the previous token is `TokenKeyword`, the next word must be
  classified as `TokenIdentifier` regardless of its value. This prevents
  names that happen to match a register or instruction from being
  misclassified (e.g. `namespace mov` → keyword `namespace` + identifier
  `mov`). Because this rule is checked before profile lookups in
  `classifyWord()` (FR-4.6.8), it takes absolute precedence — no keyword
  argument can be accidentally promoted to an instruction or register.

---

## Architecture

### AR-1: File Layout

The lexer is split across multiple files within `v0/internal/kasm`. Because
each file owns a single concern, modifications to one concern (e.g. adding a
register to the x86_64 profile) do not touch files that own other concerns
(e.g. the scanning loop).

| File                        | Responsibility                                           |
|-----------------------------|----------------------------------------------------------|
| `lexer.go`                  | `Lexer` struct, `LexerNew`, `Start`, scanning methods.   |
| `token.go`                  | `Token` struct definition.                               |
| `token_types.go`            | `TokenType` enum and convenience methods.                |
| `architecture_profile.go`   | `ArchitectureProfile` interface, `defaultKeywords()`.    |
| `profile_x86_64.go`         | `NewX8664Profile()` — concrete x86_64 profile.           |

- **AR-1.1** The core lexer (`lexer.go`) must not import or reference any
  architecture-specific data. It operates exclusively through the
  `ArchitectureProfile` interface. Because the interface is defined in a
  separate file (`architecture_profile.go`), and profiles implement it in
  their own files, there is no import cycle and no compile-time coupling
  between the lexer and any specific architecture.
- **AR-1.2** Each architecture profile lives in its own file (e.g.
  `profile_x86_64.go`, `profile_arm64.go`). Adding a new architecture means
  adding a new file — no existing file is modified. Because Go compiles all
  files in a package together, the new profile is automatically available
  within the package without import changes.
- **AR-1.3** The `ArchitectureProfile` interface and default keyword helper
  live in `architecture_profile.go`, separate from both the lexer and the
  profiles. Because neither the lexer nor the profiles own the interface
  definition, neither has privileged access — all consumers go through the
  same public API.

### AR-2: Separation of Concerns

```
┌─────────────────────────────────────────────────────────┐
│                   v0/internal/kasm                       │
│                                                         │
│  ┌──────────────────┐    ┌────────────────────────────┐ │
│  │   Lexer (core)   │───▶│  ArchitectureProfile (if)  │ │
│  │   lexer.go       │    │  architecture_profile.go   │ │
│  └──────────────────┘    └────────────┬───────────────┘ │
│                                       │ implements      │
│                          ┌────────────┴───────────────┐ │
│                          │  profile_x86_64.go         │ │
│                          │  profile_arm64.go (future) │ │
│                          │  profile_riscv.go (future) │ │
│                          └────────────────────────────┘ │
│                                                         │
│  ┌──────────────────┐    ┌──────────────────┐           │
│  │   token.go       │    │  token_types.go  │           │
│  └──────────────────┘    └──────────────────┘           │
└─────────────────────────────────────────────────────────┘
```

- **AR-2.1** The lexer core depends on the `ArchitectureProfile` interface
  only — never on a concrete profile type. Because `lexer.go` has no import
  of any `profile_*.go` file, this is enforced at the source level.
- **AR-2.2** Concrete profiles may import `v0/architecture` to derive their
  instruction sets from providers. The lexer core must not import
  `v0/architecture`. Because the dependency flows from profile → architecture
  and from lexer → interface, there is no path from lexer → architecture.
- **AR-2.3** Token types and the token struct are architecture-independent.
  They must not contain any architecture-specific fields or constants.
  Because token classification happens at scan time via profile lookup, the
  token data model itself carries no architecture knowledge.

### AR-3: Adding a New Architecture

Adding support for a new architecture requires exactly two steps. Because the
lexer core and the interface are untouched, the change is purely additive —
no existing behaviour can regress.

1. **Create a profile file** (e.g. `profile_arm64.go`) that implements
   `ArchitectureProfile` with the architecture's registers, instructions, and
   keywords.
2. **Wire the profile** in the orchestrator (`assemble_file.go`) so that the
   correct profile is passed to `LexerNew` based on the target architecture.

- **AR-3.1** No changes to `lexer.go`, `token.go`, `token_types.go`, or
  `architecture_profile.go` are required when adding a new architecture.
  Because the lexer depends only on the interface (AR-2.1), and the interface
  is closed for modification, the lexer core is guaranteed stable across
  architecture additions.
- **AR-3.2** The orchestrator is responsible for selecting the correct profile
  based on CLI flags or configuration. The lexer does not perform architecture
  detection. Because profile selection happens before lexer construction
  (FR-2.1), the lexer never needs to reason about which architecture is
  active.

### AR-4: Profile Construction Patterns

- **AR-4.1** A profile may be constructed statically (hardcoded maps) or
  dynamically (from `v0/architecture` providers). Both patterns are valid.
- **AR-4.2** Static profiles are preferred for production use — they avoid
  the overhead of iterating providers at startup and make the full vocabulary
  visible in a single file. Because the maps are literals, the compiler can
  optimise them and reviewers can audit them directly.
- **AR-4.3** A `ProfileFromArchitecture` helper may be provided for testing
  or for architectures that prefer to derive from provider data. This helper
  must lower-case all mnemonics and merge the default keyword set. Because
  it normalises data at construction time, consumers never encounter mixed-
  case keys in profile maps.

---

## Non-Functional Requirements

### NFR-1: Performance

- **NFR-1.1** The lexer must perform a single pass over the input. No
  backtracking or multi-pass scanning is permitted. Because every branch of
  `Start()` consumes at least one character (FR-4), the total work is
  bounded by the input length.
- **NFR-1.2** Profile lookup tables must use `map[string]bool` for O(1)
  average-case classification. The profile methods (`Registers()`,
  `Instructions()`, `Keywords()`) must return pre-built maps, not construct
  them on each call. Because the maps are built at profile construction time
  (FR-1.2.1), per-token overhead is a single map lookup.
- **NFR-1.3** The lexer must avoid allocating intermediate strings where
  possible — `readWord()`, `readNumber()`, and `readComment()` must use
  slice indexing (`Input[start:Position]`) rather than building strings
  character by character. Because Go strings are immutable byte slices,
  sub-slicing does not copy data.
- **NFR-1.4** Profile maps are read-only after construction. No locking or
  synchronisation is required. Because the maps are never written after
  `NewX8664Profile()` returns (FR-1.1.5), multiple lexer instances may
  share the same profile concurrently without data races.

### NFR-2: Correctness

- **NFR-2.1** Line numbering must be 1-based, matching editor conventions.
  Because `Line` starts at `1` (FR-2.4) and increments on each `'\n'`
  (FR-10.2), the first line is always line 1.
- **NFR-2.2** Column numbering must be 1-based, counting bytes from the start
  of the line. Because `Column` resets to `0` on `'\n'` (FR-10.2) and
  increments to `1` on the first character of the next line, the first
  column is always column 1.
- **NFR-2.3** The lexer must handle all byte values without panicking. Unknown
  characters are emitted as single-character `TokenIdentifier` tokens.
  Because the `default` branch of the main switch handles any unrecognised
  character (FR-4.7), no byte value can cause a panic.
- **NFR-2.4** The lexer must produce identical output for a given `(input,
  profile)` pair regardless of when or how many times it is invoked
  (deterministic, side-effect-free). Because the lexer has no mutable
  package-level state and the profile is immutable (FR-1.1.5), the only
  state that changes is local to the `Lexer` instance.

### NFR-3: Testability

- **NFR-3.1** The lexer must be testable with only a source string and a
  profile — no file I/O or external dependencies are required. Because the
  lexer has no I/O and the profile is an interface, all dependencies are
  injectable.
- **NFR-3.2** Tests must live in the `kasm_test` package
  (`v0/internal/kasm/lexer_test.go`) to verify the public API surface only.
- **NFR-3.3** Tests must use the x86_64 profile by default. Architecture-
  specific edge cases (e.g. a word that is a register in ARM but not x86_64)
  must be tested with a custom or mock profile. Because the profile is an
  interface (FR-1.1), tests can supply any implementation — including one
  that returns empty maps.
- **NFR-3.4** A `NewEmptyProfile()` returning an `ArchitectureProfile` with
  empty maps should be provided for tests that need to verify classification
  falls through to `TokenIdentifier` for all words. Because the empty
  profile satisfies all three map methods with valid (empty) maps
  (FR-1.1.4), the lexer operates correctly — it simply classifies every
  word as an identifier.

### NFR-4: Integration

- **NFR-4.1** The lexer receives its input from the pre-processor output (see
  pre-processor requirements).
- **NFR-4.2** The `[]Token` slice produced by `Start()` is passed directly to
  `ParserNew()` for parsing.
- **NFR-4.3** The `Line` and `Column` fields on each token are used by the
  debug context (`debugcontext.DebugContext`) to produce accurate error
  locations when the parser or later stages report issues. Because position
  tracking is centralised in `readChar()` (FR-10), the values are always
  consistent with the source text.
- **NFR-4.4** The orchestrator (`assemble_file.go`) is responsible for creating
  the appropriate `ArchitectureProfile` and passing it to `LexerNew`. The
  lexer does not determine the target architecture. Because profile selection
  is an orchestrator concern, the lexer is decoupled from CLI flags,
  configuration files, and environment detection.

### NFR-5: Extensibility

- **NFR-5.1** Adding a new architecture must not require changes to the lexer
  core, the token types, or the `ArchitectureProfile` interface. Because the
  interface is closed and the lexer depends only on the interface (AR-2.1),
  new architectures are purely additive (AR-3.1).
- **NFR-5.2** Adding a new keyword across all architectures requires only
  updating `defaultKeywords()`. No profile files need modification unless
  they override the default keyword set. Because each profile constructor
  calls `defaultKeywords()` (FR-7.3), the new keyword propagates
  automatically.
- **NFR-5.3** Adding a new token type (e.g. `TokenMemoryOperand`) requires
  adding a constant to `token_types.go`, a convenience method, and a
  recognition rule in `Start()`. It does not require changes to the profile
  interface, because token types are a lexer-level concern — the profile
  supplies vocabulary, not token categories.

---

## Data Model

| Struct                | Purpose                                                          |
|-----------------------|------------------------------------------------------------------|
| `Lexer`               | Holds input, position state, profile ref, accumulating tokens.   |
| `Token`               | Single token: type, literal, line, column.                       |
| `TokenType`           | Integer enum classifying a token.                                |
| `ArchitectureProfile` | Interface supplying register, instruction, keyword maps.         |

## Token Type Constants

| Constant          | Value (iota) | Emitted? | Description                        |
|-------------------|--------------|----------|------------------------------------|
| `TokenWhitespace` | 0            | No       | Whitespace sequences.              |
| `TokenComment`    | 1            | No       | `;`-prefixed comments.             |
| `TokenIdentifier` | 2            | Yes      | Names, labels, punctuation.        |
| `TokenDirective`  | 3            | Yes      | `%`-prefixed directives.           |
| `TokenInstruction`| 4            | Yes      | Profile-recognised mnemonics.      |
| `TokenRegister`   | 5            | Yes      | Profile-recognised register names. |
| `TokenImmediate`  | 6            | Yes      | Numeric literals.                  |
| `TokenString`     | 7            | Yes      | `"…"` string literals.             |
| `TokenKeyword`    | 8            | Yes      | Profile-recognised keywords.       |
