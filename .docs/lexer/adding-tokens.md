# Adding a New Token to the Lexer

This guide explains how to introduce a new token type to the `.kasm` lexer.
It covers every file that must be touched, the order in which changes should
be made, and the conventions that keep the lexer architecture-agnostic and
single-pass.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture Refresher](#architecture-refresher)
3. [Decision: Profile-Based vs Syntax-Based Token](#decision-profile-based-vs-syntax-based-token)
4. [Adding a Syntax-Based Token](#adding-a-syntax-based-token)
5. [Adding a Profile-Based Token](#adding-a-profile-based-token)
6. [Adding a Vocabulary Entry to an Existing Token Type](#adding-a-vocabulary-entry-to-an-existing-token-type)
7. [Downstream Impact](#downstream-impact)
8. [Testing](#testing)
9. [Checklist](#checklist)

---

## Overview

The lexer (`v0/kasm/lexer.go`) performs a single-pass, left-to-right scan of
pre-processed `.kasm` source and emits a `[]Token` slice. Each token carries
a `TokenType`, a `Literal` string, and `Line`/`Column` coordinates.

Token classification falls into two categories:

| Category       | Determined by                 | Examples                                   |
|----------------|-------------------------------|--------------------------------------------|
| Syntax-based   | Character pattern at scan time| `TokenDirective` (`%`), `TokenImmediate` (digit), `TokenString` (`"`) |
| Profile-based  | Lookup against `ArchitectureProfile` maps | `TokenRegister`, `TokenInstruction`, `TokenKeyword` |

Adding a new token type always touches **at least three files**. Depending
on the category, additional files may be involved.

---

## Architecture Refresher

```
v0/kasm/
├── lexer.go                             # Lexer struct, LexerNew, Start(), scanning methods
├── token.go                             # Token struct (Type, Literal, Line, Column)
├── token_types.go                       # TokenType enum (iota), convenience methods
├── lexer_test.go                        # Tests (kasm_test package)
├── profile/
│   ├── architecture_profile.go          # ArchitectureProfile interface, defaultKeywords()
│   ├── profile_x86_64.go               # NewX8664Profile(), staticProfile, register/instruction maps
│   └── (future profile files)

cmd/cli/cmd/x86_64/
├── assemble_file.go                     # Orchestrator — creates profile, wires lexer
```

Data flow:

```
ArchitectureProfile ──────┐
                          ▼
pre-processed source → LexerNew(input, profile) → .Start() → []Token → parser
```

Key design constraints (from requirements):

- The lexer core (`lexer.go`) must **never** import architecture-specific
  packages. All vocabulary flows through `ArchitectureProfile` (AR-2.1).
- The lexer performs a **single pass** — no backtracking (NFR-1.1).
- Every character is consumed by exactly one branch of the main `switch` in
  `Start()`. No character is silently dropped or double-counted (FR-4).
- Token types are a **lexer-level concern** — the profile supplies vocabulary,
  not token categories (NFR-5.3).
- The profile is **immutable** after construction (FR-1.1.5).

---

## Decision: Profile-Based vs Syntax-Based Token

Before writing code, decide which category the new token belongs to.

### Use a syntax-based token when:

- The token is recognised by a **leading character or character pattern**
  that is unique and unambiguous (e.g. `%` for directives, `"` for strings,
  digit for immediates).
- The token is **architecture-independent** — it exists in every `.kasm`
  dialect regardless of target CPU.
- The recognition logic belongs in the scanner loop, not in a vocabulary map.

### Use a profile-based token when:

- The token is a **word** (letters, digits, underscores, dots) whose
  classification depends on whether it appears in a vocabulary set.
- The vocabulary may **differ between architectures** (e.g. ARM registers
  are different from x86_64 registers, but both are `TokenRegister`).
- Adding a new word to the vocabulary should not require changes to
  `lexer.go`.

### Neither — just add vocabulary to an existing type:

Most of the time you don't need a new `TokenType` at all. If you need to
add a new register, instruction, or keyword, see
[Adding a Vocabulary Entry to an Existing Token Type](#adding-a-vocabulary-entry-to-an-existing-token-type).

---

## Adding a Syntax-Based Token

This section walks through adding a token type that is recognised by a
character pattern. We use a hypothetical `TokenCharLiteral` (for `'A'`
syntax) as a running example.

### Step 1 — Add the constant to `token_types.go`

Append a new constant to the `iota` block. **Always append at the end** —
inserting in the middle shifts all subsequent values and breaks any code that
depends on the integer value.

```go
const (
    TokenWhitespace TokenType = iota
    TokenComment
    TokenIdentifier
    TokenDirective
    TokenInstruction
    TokenRegister
    TokenImmediate
    TokenString
    TokenKeyword
    TokenCharLiteral // ← new
)
```

### Step 2 — Add a convenience method to `token_types.go`

Follow the existing pattern — one method per token type that returns `true`
only for its own constant:

```go
// CharLiteral returns true only for TokenCharLiteral.
func (tT TokenType) CharLiteral() bool {
    return tT == TokenCharLiteral
}
```

If the new token should be ignored by the parser (like whitespace and
comments), update `Ignored()`:

```go
func (tT TokenType) Ignored() bool {
    switch tT {
    case TokenWhitespace, TokenComment, TokenMySkippedType:
        return true
    default:
        return false
    }
}
```

### Step 3 — Add a reader method to `lexer.go`

Write a private method that consumes the token's characters from the input.
Follow the naming convention `read<Thing>()`:

```go
// readCharLiteral reads a character literal enclosed in single quotes.
func (l *Lexer) readCharLiteral() string {
    l.readChar() // skip opening '
    start := l.Position
    for l.Ch != '\'' && l.Ch != 0 {
        l.readChar()
    }
    str := l.Input[start:l.Position]
    if l.Ch == '\'' {
        l.readChar() // skip closing '
    } else if l.debugCtx != nil {
        l.debugCtx.Warning(
            l.debugCtx.Loc(l.Line, l.Column),
            "unterminated character literal",
        )
    }
    return str
}
```

Rules for reader methods:

- Use `l.readChar()` for all advancement — this keeps `Line`/`Column`
  accurate (FR-10).
- Use slice indexing (`l.Input[start:l.Position]`) for the literal — this
  avoids allocations (NFR-1.3).
- Handle unterminated/malformed input gracefully — consume until EOF, never
  panic (FR-3.3).
- If a debug context is attached, record warnings for malformed input.

### Step 4 — Add a branch to `Start()` in `lexer.go`

Insert a new `case` in the main `switch` inside `Start()`. The position
of the case matters — it must come **before** the `default` branch (which
catches all unrecognised single characters):

```go
func (l *Lexer) Start() []Token {
    // ...existing setup...

    for l.Ch != 0 {
        line := l.Line
        col := l.Column

        switch {
        // ...existing cases (whitespace, comment, directive, string, digit, word)...

        // Character literal — '…'.
        case l.Ch == '\'':
            ch := l.readCharLiteral()
            l.addToken(TokenCharLiteral, ch, line, col)

        // Any other single character — emit as identifier.
        default:
            l.addToken(TokenIdentifier, string(l.Ch), line, col)
            l.readChar()
        }
    }

    // ...existing trace logging...
    return l.Tokens
}
```

Important: capture `line` and `col` **before** calling the reader method so
the token's position reflects its first character (FR-3.5).

### Step 5 — Write tests

See [Testing](#testing) below.

### Step 6 — Update downstream consumers

See [Downstream Impact](#downstream-impact) below.

---

## Adding a Profile-Based Token

Profile-based tokens extend the word-classification mechanism. This is a
rarer change — most vocabulary additions don't need a new type. Use this
path when an entirely new word category is introduced (e.g. a hypothetical
`TokenMacroName` for macro identifiers).

### Step 1 — Add the constant and convenience method

Same as Steps 1–2 in the syntax-based flow above. Add the constant to the
`iota` block in `token_types.go` and write the convenience method.

### Step 2 — Extend the `ArchitectureProfile` interface

In `v0/kasm/profile/architecture_profile.go`, add a new method to the
interface:

```go
type ArchitectureProfile interface {
    Registers() map[string]bool
    Instructions() map[string]bool
    Keywords() map[string]bool
    MacroNames() map[string]bool  // ← new
}
```

> **Warning:** This is a breaking change — every existing profile
> implementation must be updated. This is why profile-based token types are
> rare. Prefer using an existing category or a syntax-based token when
> possible.

### Step 3 — Update all profile implementations

Update every concrete profile and the empty profile:

**`profile/architecture_profile.go`** — update `emptyProfile`:

```go
type emptyProfile struct {
    registers    map[string]bool
    instructions map[string]bool
    keywords     map[string]bool
    macroNames   map[string]bool  // ← new
}

func NewEmptyProfile() ArchitectureProfile {
    return &emptyProfile{
        // ...existing fields...
        macroNames: make(map[string]bool),
    }
}

func (p *emptyProfile) MacroNames() map[string]bool { return p.macroNames }
```

**`profile/profile_x86_64.go`** — update `staticProfile` and the constructor:

```go
type staticProfile struct {
    registers    map[string]bool
    instructions map[string]bool
    keywords     map[string]bool
    macroNames   map[string]bool  // ← new
}

func (p *staticProfile) MacroNames() map[string]bool { return p.macroNames }

func NewX8664Profile() ArchitectureProfile {
    return &staticProfile{
        registers:    x8664Registers(),
        instructions: x8664Instructions(),
        keywords:     defaultKeywords(),
        macroNames:   make(map[string]bool), // empty for now, or populate
    }
}
```

Also update `FromArchitecture()` if it exists.

### Step 4 — Update `classifyWord()` in `lexer.go`

Add a lookup for the new profile method. The order of lookups in
`classifyWord()` defines priority — a word that appears in multiple maps
is classified by whichever lookup comes first:

```go
func classifyWord(word string, lexer *Lexer) TokenType {
    if lexer.previousTokenType() == TokenKeyword {
        return TokenIdentifier
    }

    lower := strings.ToLower(word)

    if lexer.profile.Registers()[lower] {
        return TokenRegister
    }
    if lexer.profile.Instructions()[lower] {
        return TokenInstruction
    }
    if lexer.profile.MacroNames()[lower] {  // ← new lookup
        return TokenMacroName
    }
    if lexer.profile.Keywords()[lower] {
        return TokenKeyword
    }

    return TokenIdentifier
}
```

Decide where in the priority chain the new lookup should fall. The current
order is: registers → instructions → keywords → identifier. Insert the new
lookup at the appropriate position.

### Step 5 — Write tests and update downstream consumers

See [Testing](#testing) and [Downstream Impact](#downstream-impact).

---

## Adding a Vocabulary Entry to an Existing Token Type

This is the most common change. It requires **no changes to `lexer.go`**,
**no changes to `token_types.go`**, and **no changes to the
`ArchitectureProfile` interface**.

### Adding a new register

Edit `v0/kasm/profile/profile_x86_64.go` — add the entry to the
`x8664Registers()` map:

```go
func x8664Registers() map[string]bool {
    return map[string]bool{
        // ...existing registers...
        "ymm0": true, // ← new AVX register
    }
}
```

### Adding a new instruction mnemonic

Edit `v0/kasm/profile/profile_x86_64.go` — add the entry to the
`x8664Instructions()` map:

```go
func x8664Instructions() map[string]bool {
    return map[string]bool{
        // ...existing instructions...
        "vaddps": true, // ← new AVX instruction
    }
}
```

If the instruction also has architecture metadata (opcode, variants), add it
to the appropriate file in `v0/architecture/x86/_64/` as well. The profile
and the architecture metadata are independent — both must be updated
separately.

### Adding a new keyword (all architectures)

Edit `v0/kasm/profile/architecture_profile.go` — add the entry to
`defaultKeywords()`:

```go
func defaultKeywords() map[string]bool {
    return map[string]bool{
        "namespace": true,
        "section":   true, // ← new keyword
    }
}
```

Because every profile constructor calls `defaultKeywords()` (FR-7.3), the
new keyword is automatically available in all profiles.

### Adding a keyword for one architecture only

Edit that architecture's profile constructor to extend the keywords map
after calling `defaultKeywords()`:

```go
func NewX8664Profile() ArchitectureProfile {
    kw := defaultKeywords()
    kw["segment"] = true  // x86_64-specific keyword
    return &staticProfile{
        registers:    x8664Registers(),
        instructions: x8664Instructions(),
        keywords:     kw,
    }
}
```

### Write a test

For any vocabulary addition, write a test that tokenises the new word and
asserts its type:

```go
func TestLexer_NewAVXRegister(t *testing.T) {
    tokens := kasm.LexerNew("ymm0", x86Profile).Start()
    requireTokenCount(t, tokens, 1)
    requireToken(t, tokens[0], kasm.TokenRegister, "ymm0")
}
```

---

## Downstream Impact

Adding a new token type affects consumers of the `[]Token` slice. Consider
each downstream stage:

### Parser (`v0/kasm/parsing.go`)

The parser switches on `TokenType` to build AST nodes. A new token type may
need:

- A new case in the parser's main dispatch loop.
- A new AST node type in `ast.go` if the token introduces a new language
  construct.
- Or the token may fit into an existing parsing rule (e.g. a new operand
  kind handled alongside existing operands).

If the parser does not handle the new token type, it will fall through to the
parser's default/error handling — tokens won't be silently lost, but they
will produce parse errors. Ensure the parser is updated before merging.

### Semantic analyser (`v0/kasm/semantic.go`)

If the new token type leads to a new AST node type, the semantic analyser
needs corresponding collection/validation logic. See
`.docs/semantic-analyzer/adding-features.md` for details.

### Debug context

The `Token.Line` and `Token.Column` fields are already set by the lexer.
No debug context changes are needed for new token types — the existing
infrastructure handles all token types uniformly.

### Tests across stages

- **Lexer tests** (`lexer_test.go`): test the token is emitted correctly.
- **Parser tests** (`parsing_test.go`): test the token is parsed into the
  correct AST node.
- **Semantic tests** (`semantic_test.go`): test any new validation rules.

---

## Testing

Tests live in `v0/kasm/lexer_test.go` in the `kasm_test` package (black-box
testing).

### Test helpers

| Helper              | Purpose                                        |
|---------------------|------------------------------------------------|
| `requireTokenCount` | Asserts the number of tokens returned.         |
| `requireToken`      | Asserts a token's `Type` and `Literal`.        |
| `x86Profile`        | Package-level `profile.NewX8664Profile()` used by default. |

### Writing a test for a syntax-based token

Test the new character pattern in isolation:

```go
func TestLexer_CharLiteral(t *testing.T) {
    tokens := kasm.LexerNew("'A'", x86Profile).Start()
    requireTokenCount(t, tokens, 1)
    requireToken(t, tokens[0], kasm.TokenCharLiteral, "A")
}
```

Cover edge cases:

```go
func TestLexer_CharLiteralEmpty(t *testing.T) {
    tokens := kasm.LexerNew("''", x86Profile).Start()
    requireTokenCount(t, tokens, 1)
    requireToken(t, tokens[0], kasm.TokenCharLiteral, "")
}

func TestLexer_CharLiteralUnterminated(t *testing.T) {
    tokens := kasm.LexerNew("'A", x86Profile).Start()
    requireTokenCount(t, tokens, 1)
    requireToken(t, tokens[0], kasm.TokenCharLiteral, "A")
}
```

### Writing a test for a profile-based token

Use the default `x86Profile` if the word is in the x86_64 vocabulary, or
create a custom profile for testing:

```go
func TestLexer_MacroName(t *testing.T) {
    // Use a custom profile that includes the macro name
    tokens := kasm.LexerNew("MY_MACRO", customProfile).Start()
    requireTokenCount(t, tokens, 1)
    requireToken(t, tokens[0], kasm.TokenMacroName, "MY_MACRO")
}
```

### Writing a test for a vocabulary addition

```go
func TestLexer_NewInstruction_VADDPS(t *testing.T) {
    tokens := kasm.LexerNew("vaddps", x86Profile).Start()
    requireTokenCount(t, tokens, 1)
    requireToken(t, tokens[0], kasm.TokenInstruction, "vaddps")
}
```

### Test coverage requirements

For every change, provide at least:

- **Happy path**: valid input → correct token type and literal.
- **Case insensitivity** (for profile-based tokens): upper-case, lower-case,
  and mixed-case variants produce the same token type.
- **Edge case**: empty content, unterminated constructs, EOF mid-token.
- **Context interaction**: the new token after a keyword (FR-11.2), the new
  token among other tokens in a realistic instruction line.

### Running tests

```bash
cd v0/kasm && go test -v -run TestLexer
```

---

## Checklist

Use this checklist when adding any token to the lexer:

### New token type (syntax-based)

- [ ] **Constant**: `TokenMyType` appended to the `iota` block in
      `token_types.go` (appended at the end, not inserted).
- [ ] **Convenience method**: `MyType() bool` method added to `TokenType`
      in `token_types.go`.
- [ ] **`Ignored()` updated** (only if the new token should be skipped by
      the parser).
- [ ] **Reader method**: `read<Thing>()` added to `lexer.go` using
      `readChar()` for advancement and slice indexing for the literal.
- [ ] **Branch in `Start()`**: new `case` added to the main `switch` in
      `Start()`, before the `default` branch. Captures `line`/`col` before
      calling the reader.
- [ ] **Graceful degradation**: reader handles unterminated/malformed input
      by consuming to EOF without panicking.
- [ ] **Debug context**: warnings recorded for malformed input when
      `debugCtx` is attached.
- [ ] **Tests**: happy path, edge cases, and context interaction tests in
      `lexer_test.go`.
- [ ] **Parser updated**: new token type handled in the parser's dispatch.
- [ ] **Requirements**: no changes to `requirements.md` without approval,
      but note the new token in any design documents.

### New token type (profile-based)

- [ ] All items from the syntax-based checklist, plus:
- [ ] **Interface extended**: new method added to `ArchitectureProfile` in
      `profile/architecture_profile.go`.
- [ ] **All profiles updated**: `emptyProfile`, `staticProfile`,
      `NewX8664Profile()`, and `FromArchitecture()` all implement the new
      method.
- [ ] **`classifyWord()` updated**: new lookup added at the correct priority
      position in `lexer.go`.
- [ ] **No `lexer.go` architecture imports**: `lexer.go` still imports only
      the `profile` sub-package, never a concrete architecture package.

### Vocabulary addition (no new token type)

- [ ] **Profile map updated**: entry added to the correct map in the correct
      profile file (`profile_x86_64.go`, `architecture_profile.go`, etc.).
- [ ] **Architecture metadata** (if instruction): instruction added to the
      appropriate file in `v0/architecture/x86/_64/`.
- [ ] **Test**: tokenisation test asserting correct type for the new word.
- [ ] **No changes to `lexer.go`**: vocabulary additions are data-only.
- [ ] **No changes to `token_types.go`**: existing types cover the new word.

