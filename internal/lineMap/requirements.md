# lineMap — Requirements

## Overview

The `lineMap` package tracks how source lines transform during pre-processing steps
(includes, macros, conditionals). It enables tracing any line in the final processed
source back to its original location in the initial source file.

---

## Functional Requirements

### FR-1: Source Loading

A `Source` represents a validated, loaded source file. If a `Source` value exists,
it is guaranteed to hold a valid path and its file content. There is no unloaded
or partially-initialised state.

- **FR-1.1** A `Source` is created exclusively through `LoadSource(path)`, which
  validates the path, reads the file, and returns a ready-to-use `Source` — or an error.
  There is no separate `Load()` step.
- **FR-1.2** `LoadSource` must validate that the path ends with the `.kasm` extension
  (case-sensitive). Paths with any other extension must be rejected with a descriptive error.
- **FR-1.3** `LoadSource` must call `os.Stat` and return an error for **any** failure
  (file-not-found, permission denied, broken symlink, etc.).
- **FR-1.4** `LoadSource` must verify that the path refers to a regular file, not a
  directory. Directories must be rejected with a descriptive error.
- **FR-1.5** `LoadSource` must read the file content via `os.ReadFile` and return an
  error if reading fails.
- **FR-1.6** On success, the returned `Source` must hold the file path and its content.
  Both are immutable after construction.
- **FR-1.7** `Source` must expose `Path() string` and `Content() string` accessor methods
  for read-only access.
- **FR-1.8** Because all validation happens at construction, callers can trust that any
  `Source` value they receive is valid. No nil-checks or error-checks are needed after
  construction.

### FR-2: Instance Lifecycle

An `Instance` represents an initialised, indexed line map. If an `Instance` value
exists, it is guaranteed to hold a valid source, an initial snapshot, and be ready
for `Update()` calls. There is no uninitialised or partially-constructed state.

- **FR-2.1** An `Instance` is created exclusively through `New(value, source)`, which
  takes the initial source string and a `Source` obtained from `LoadSource()`. It
  performs the initial indexing (first snapshot) and returns a ready-to-use `*Instance`.
  `New()` is infallible — it cannot fail because `Source` is guaranteed valid (FR-1) and
  initial snapshot creation is unconditional.
- **FR-2.2** Because `Source` is guaranteed valid after construction (see FR-1), `New()`
  does not need to perform source validation.
- **FR-2.3** `New()` must create the initial snapshot (type `LineSnapshotTypeInitial`)
  as part of construction. There is no separate `InitialIndex()` step.
- **FR-2.4** After `New()` returns successfully, `Update()` can be called immediately
  without any precondition checks. Callers can trust that any `*Instance` they receive
  is fully initialised.
- **FR-2.5** There is no `state` field. The existence of the `Instance` itself is the
  proof that it is valid and indexed.

### FR-3: Initial Indexing

Initial indexing is an internal implementation detail of `New()`. It is not exposed
as a public method, and it is structurally impossible to create a second initial
snapshot.

- **FR-3.1** The initial snapshot must split `Instance.value` into lines and store them
  in the history.
- **FR-3.2** The initial snapshot must have type `LineSnapshotTypeInitial`.
- **FR-3.3** `History` exposes separate internal snapshot methods:
  - `snapshotInitial(instance)` — creates the initial snapshot. Called only from `New()`.
  - `snapshotNoChange(instance)` — records a no-change update. Called only from `Update()`.
  - `snapshotChange(instance, changes)` — records a change update. Called only from `Update()`.
- **FR-3.4** Because `snapshotInitial` is only called from `New()`, and neither
  `snapshotNoChange` nor `snapshotChange` produce an initial snapshot, a second initial
  snapshot cannot be created. No snapshot method accepts a `_type` string parameter —
  the snapshot type is determined by which method is called.

### FR-4: Updating (Snapshotting)

`Update()` transitions an `Instance` to a new source value and records the
transformation in the history. Because `New()` guarantees at least one snapshot
exists (FR-2), `Update()` can rely on `latest()` always returning a valid
snapshot. There are no precondition checks or error paths.

- **FR-4.1** `Update(newValue)` is infallible — it cannot fail because the initial
  snapshot is guaranteed to exist (FR-2) and change detection operates on data that
  is always present.
- **FR-4.2** If the new value is identical to the latest snapshot (compared by hash),
  a no-change snapshot is recorded via `snapshotNoChange()`.
- **FR-4.3** If the new value differs from the latest snapshot, a change snapshot is
  recorded via `snapshotChange()` with the computed change map.
- **FR-4.4** After an update, `Instance.value` must reflect the new value.
- **FR-4.5** `History` exposes two separate internal methods for update snapshots
  (replacing the single `snapshotUpdate`):
  - `snapshotNoChange(instance)` — records that the source did not change.
  - `snapshotChange(instance, changes)` — records a diff. The `changes` parameter is
    always non-nil.
  Neither method accepts a `_type` string parameter — the snapshot type is determined
  by which method is called, not by a runtime value.

### FR-5: Change Detection (Diff)

Change detection produces a detailed, per-line diff between two source versions.
Every line in both the old and new version is accounted for. Each classification
is represented by a dedicated factory function that only accepts the fields
relevant to that variant. There is no generic constructor that accepts a type
string — the variant is determined by which factory is called.

#### FR-5.1: Algorithm

- **FR-5.1.1** Changes between two source versions must be computed using a Longest
  Common Subsequence (LCS) algorithm at the line level.

#### FR-5.2: Change Variants and Factories

Each line is classified as one of three variants, each with its own factory:

- **FR-5.2.1** `newUnchangedChange(origin, newIndex, content)` — the line exists in
  both versions. Records where it was (origin), where it is now (newIndex), and its
  content. This enables consumers to see positional shifts even for unchanged lines.
- **FR-5.2.2** `newExpandingChange(origin, newIndex, content)` — the line is new
  (inserted/added) in the new version. Records the nearest origin line it maps to, its
  position in the new version (newIndex), and the inserted line content.
- **FR-5.2.3** `newContractingChange(origin, content)` — the line was removed from the
  old version. Records its position in the old version (origin) and the removed line
  content.

#### FR-5.3: Infallible Construction

- **FR-5.3.1** All three factory functions are infallible — they return a `LineChange`
  value (not a pointer, no error).
- **FR-5.3.2** Invalid states are prevented by the function signatures: each factory
  only accepts the parameters that are meaningful for its variant.
- **FR-5.3.3** There is no generic `newLineChange(_type string, ...)` constructor. The
  change type is set internally by each factory, not passed as a runtime parameter.

#### FR-5.4: LineChange Detail

Every `LineChange` must carry enough information for a consumer to produce a
detailed diff report without needing to look up the snapshot source:

- **FR-5.4.1** `_type` — always set by the factory, never by callers. One of
  `"unchanged"`, `"expanding"`, `"contracting"`.
- **FR-5.4.2** `origin` — the 0-based line index in the previous (old) version. Present
  on all three variants.
- **FR-5.4.3** `newIndex` — the 0-based line index in the new version. Present on
  `unchanged` and `expanding` changes. Set to `-1` for `contracting` changes (the line
  no longer exists in the new version).
- **FR-5.4.4** `content` — the actual text of the line. For `unchanged` and `expanding`
  this is the line in the new version. For `contracting` this is the line that was
  removed.
- **FR-5.4.5** `sourceFile` — the file path that a line originated from. Set on
  `expanding` changes when the line was inserted from an included file (e.g. via
  `%include`). Empty string for lines that originate from the main source file or for
  `unchanged` and `contracting` changes. This enables consumers to see which included
  file contributed each inserted line.

#### FR-5.5: Accessor Methods

`LineChange` must expose read-only accessor methods so consumers don't access
internal fields directly:

- **FR-5.5.1** `Type() string` — returns the change type.
- **FR-5.5.2** `Origin() int` — returns the origin line index.
- **FR-5.5.3** `NewIndex() int` — returns the new line index (`-1` for contracting).
- **FR-5.5.4** `Content() string` — returns the line content.
- **FR-5.5.5** `String() string` — returns a human-readable representation for debugging.
- **FR-5.5.6** `SourceFile() string` — returns the file path that the line originated
  from. Empty string if the line is from the main source file or is not an expanding
  change.

#### FR-5.6: Change Map Keying

- **FR-5.6.1** The change map returned by `changes()` must be keyed by the **new version
  line index** for `unchanged` and `expanding` entries.
- **FR-5.6.2** Contracting entries (removed lines) must be stored in a separate
  `removals` slice on the snapshot rather than mixed into the same map, because they
  have no position in the new version.

### FR-6: Line Origin Tracing

`LineOrigin(lineNumber)` traces a line in the latest processed source back to its
original position in the initial source. Because an `Instance` is guaranteed to have
an initial snapshot (FR-2), and contracting entries are stored in `removals` rather
than in the changes map (FR-5.6), the tracing logic only needs to handle two cases
in the map: `unchanged` and `expanding`.

- **FR-6.1** `LineOrigin(lineNumber)` must walk backwards through the change snapshots
  (skipping the initial snapshot at index 0) and trace `lineNumber` back to the
  original line index.
- **FR-6.2** For `unchanged` entries in the changes map, the origin is the corresponding
  line index in the previous version. Tracing continues with that origin.
- **FR-6.3** For `expanding` entries in the changes map, the line was inserted during
  pre-processing and has no origin. `LineOrigin()` must return `-1`.
- **FR-6.4** If a line number is not present in a snapshot's changes map, it maps 1:1
  (unchanged without positional shift). Tracing continues with the same line number.
- **FR-6.5** There is no `h.empty()` guard. `LineOrigin` is only reachable on a fully
  constructed `Instance` (FR-2), which guarantees at least one snapshot exists.
- **FR-6.6** There is no `"contracting"` case in the changes map lookup. Contracting
  entries live in the `removals` slice (FR-5.6.2) and cannot appear as map values.
  The changes map only contains `unchanged` and `expanding` entries.

### FR-7: Line History

`LineHistory(lineNumber)` returns the evolution of a specific line across all
snapshots, from oldest to newest. Because contracting entries live in the
`removals` slice (FR-5.6.2), the changes map only contains `unchanged` and
`expanding` entries — those are the only two cases to handle.

- **FR-7.1** `LineHistory(lineNumber)` must return a slice of `LineChange` entries
  describing how a specific line evolved across all change snapshots, in
  **chronological order** (oldest change first, most recent last).
- **FR-7.2** The method must walk backwards through snapshots (skipping the initial
  snapshot at index 0, which has no changes) to build the origin chain, then reverse
  the result to produce chronological order.
- **FR-7.3** For `unchanged` entries in the changes map, the line existed in the
  previous version at `Origin()`. Tracing continues with that origin index.
- **FR-7.4** For `expanding` entries in the changes map, the line was inserted during
  pre-processing and did not exist before this snapshot. The expanding entry is
  recorded and tracing **stops** — there is no earlier origin to follow.
- **FR-7.5** If a line number is not present in a snapshot's changes map, it maps 1:1
  (unchanged, no positional shift). An `unchanged` entry is synthesised with content
  resolved from the snapshot's lines, and tracing continues with the same line number.
- **FR-7.6** There is no `"contracting"` case in the changes map lookup. Contracting
  entries live in the `removals` slice (FR-5.6.2) and cannot appear as map values.
- **FR-7.7** Each `LineChange` in the returned slice carries full detail (type, origin,
  newIndex, content) as specified by FR-5.4, so consumers can render a complete
  per-line history without additional lookups.

### FR-8: Snapshot Hashing

Snapshot hashing enables fast equality checks without comparing full source strings.
The hash function is a pure, stateless operation — it does not depend on `History`
or any other struct state.

- **FR-8.1** Each snapshot must store a SHA-256 hash of its source content, computed at
  snapshot creation time.
- **FR-8.2** `SourceCompare(value)` on `LinesSnapshot` must compare the snapshot's stored
  hash against the hash of the provided value, enabling fast equality checks in `Update()`.
- **FR-8.3** `generateSourceHash(source)` is a package-level function, not a method on
  `History`. It takes a string and returns its SHA-256 hex digest. It is called directly
  by the snapshot factory methods and `SourceCompare`.
- **FR-8.4** There is no `snapshotHashGenerate` wrapper method on `History` — the
  package-level `generateSourceHash` is called directly.
- **FR-8.5** There is no `snapshotHashCompare` method. It was never called and is removed.
  Consumers compare hashes via `SourceCompare` or direct string comparison.

### FR-9: History Management

`History` is an ordered collection of snapshots. Because an `Instance` is guaranteed
to have at least one snapshot (FR-2), several defensive helpers are unnecessary.

- **FR-9.1** The `History` must store an ordered slice of `LinesSnapshot` entries.
- **FR-9.2** Only one `LineSnapshotTypeInitial` snapshot may exist. This is enforced
  structurally by the API (see FR-3.3 / FR-3.4), not by runtime checks.
- **FR-9.3** `latest()` must return a pointer to the most recent snapshot. There is no
  nil return and no `empty()` guard — `latest()` is only called on a fully constructed
  `Instance` (FR-2), which guarantees at least one snapshot exists.
- **FR-9.4** There is no `empty()` or `notEmpty()` method. These were only used by
  `latest()` and `LineOrigin()` as defensive guards for a state that is impossible by
  construction. The `len(h.items)` check is used directly where needed (e.g. `latest()`
  is guaranteed non-empty, `LineOrigin` skips index 0).

### FR-10: Accessor Methods

Accessor methods on `Instance` provide read-only access to the current state.
Because `Instance` is guaranteed to have at least one snapshot (FR-2), accessors
that depend on snapshot state do not need nil guards.

- **FR-10.1** `Value()` must return the current source string of the instance.
- **FR-10.2** `Lines()` must return the lines from the latest snapshot. There is no nil
  guard — the latest snapshot is guaranteed to exist (FR-2).
- **FR-10.3** `SnapshotCount()` must return the total number of snapshots in the history.
- **FR-10.4** `LatestSnapshot()` must return a pointer to the most recent snapshot. There
  is no nil return — the latest snapshot is guaranteed to exist (FR-2).

### FR-11: Facade

The facade provides a simplified, high-level API for the most common lineMap
workflow: load a source file, track it through pre-processing transformations,
and trace lines back to their origin. It composes `LoadSource`, `New`, `Update`,
`LineOrigin`, and `LineHistory` into a minimal surface that eliminates boilerplate
for callers.

#### FR-11.1: Construction

- **FR-11.1.1** `Track(path)` is the single entry point. It calls `LoadSource(path)` to
  validate and read the file, then calls `New(content, source)` to create the `Instance`
  with its initial snapshot. It returns a `*Tracker` — or an error if `LoadSource` fails.
- **FR-11.1.2** `Track` is the only way to create a `Tracker`. If a `Tracker` exists, it
  is guaranteed to hold a valid, fully initialised `Instance`.
- **FR-11.1.3** The file content used for the initial snapshot is the content already
  read by `LoadSource` — the caller does not need to read the file separately.

#### FR-11.2: Snapshotting

- **FR-11.2.1** `Snapshot(source)` records a new version of the source after a
  pre-processing step. It delegates to `Instance.Update(source)`.
- **FR-11.2.2** `Snapshot` is infallible — it delegates to `Update` which is infallible
  (FR-4.1).
- **FR-11.2.3** `SnapshotWithInclusions(source, inclusions)` records a new version of
  the source after handling `%include` directives. After calling `Instance.Update(source)`,
  it walks the expanding entries in the latest snapshot's changes map and annotates each
  one with the `sourceFile` of the included file it belongs to.
  - The `inclusions` parameter is a list of `Inclusion` structs, each carrying
    `FilePath` and `LineNumber`.
  - The annotation is derived from the `; FILE: <path>` / `; END FILE: <path>` comment
    markers that `PreProcessingHandleIncludes` wraps around each included file's content.
    Lines between a `; FILE:` and its corresponding `; END FILE:` marker are annotated
    with that file path.
  - Lines outside any `; FILE:` / `; END FILE:` block (i.e. from the main source) are
    not annotated (their `sourceFile` remains empty).

#### FR-11.3: Tracing

- **FR-11.3.1** `Origin(lineNumber)` traces a line in the latest processed source back
  to its original line number. It delegates to `Instance.LineOrigin(lineNumber)`.
  Returns `-1` if the line was inserted during pre-processing.
- **FR-11.3.2** `History(lineNumber)` returns the chronological evolution of a line
  across all snapshots. It delegates to `Instance.LineHistory(lineNumber)`.

#### FR-11.4: Read Access

- **FR-11.4.1** `Source()` returns the current processed source string. It delegates to
  `Instance.Value()`.
- **FR-11.4.2** `Lines()` returns the lines of the current processed source. It delegates
  to `Instance.Lines()`.
- **FR-11.4.3** `FilePath()` returns the original file path that was passed to `Track()`.
  It delegates to `Source.Path()`.

#### FR-11.5: Design Constraints

- **FR-11.5.1** The `Tracker` struct is defined in a separate file (`tracker.go`) to keep
  the facade isolated from the core implementation.
- **FR-11.5.2** The `Tracker` does not duplicate logic — every method delegates to the
  underlying `Instance` or `Source`. It is a pure composition layer.
- **FR-11.5.3** The `Tracker` does not expose the underlying `Instance` or `Source`
  directly. Callers interact only through the facade methods.

---

## Non-Functional Requirements

### NFR-1: Performance

- **NFR-1.1** The LCS computation has O(m × n) time and space complexity. For large files,
  this must remain acceptable within the assembler's total processing time.
- **NFR-1.2** Hash-based equality checks must be used to avoid unnecessary diff computations
  when sources have not changed.

### NFR-2: Correctness

- **NFR-2.1** Line numbering must be 0-based (index into the `lines` slice).
- **NFR-2.2** Snapshot integrity must be maintained — once created, a snapshot's content and
  hash must not be modified.

### NFR-3: Testability

- **NFR-3.1** File I/O functions (`os.Stat`, `os.ReadFile`) must be injectable via package-level
  variables (`osStat`, `osReadFile`) to allow unit testing without real files.

### NFR-4: Integration

- **NFR-4.1** The `lineMap` package must integrate with the assembler's pre-processing pipeline
  as follows:
  1. Load the source file via `LoadSource(path)`.
  2. Create an `Instance` via `New(value, source)` — this establishes the baseline snapshot.
  3. After each pre-processing step (includes → macros → conditionals), call `Update()` with
     the transformed source.
  4. Use `LineOrigin()` / `LineHistory()` during error reporting or debugging to map processed
     line numbers back to original source locations.

---

## Data Model

| Struct            | Purpose                                                        |
|-------------------|----------------------------------------------------------------|
| `Source`          | Represents the source file (path + content).                   |
| `Instance`        | Main entry point; holds current value, source, and history.    |
| `History`         | Ordered collection of `LinesSnapshot` entries.                 |
| `LinesSnapshot`   | Immutable snapshot: type, hash, source, lines, changes.        |
| `LineChange`      | Describes a single line-level change (type, origin, ranges).   |

## Snapshot Types

| Constant                   | Meaning                                     |
|----------------------------|---------------------------------------------|
| `LineSnapshotTypeInitial`  | First snapshot, created by `New()`.         |
| `LineSnapshotTypeChange`   | Source changed; diff is attached.            |
| `LineSnapshotTypeNoChange` | Source identical to previous snapshot.       |

## Change Types

| Type           | Meaning                                          |
|----------------|--------------------------------------------------|
| `unchanged`    | Line is the same in both versions.               |
| `expanding`    | Line was added/inserted in the new version.      |
| `contracting`  | Line was removed from the old version.           |

