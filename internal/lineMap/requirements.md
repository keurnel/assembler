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
- **FR-3.3** `History` exposes two separate internal snapshot methods:
  - `snapshotInitial(instance)` — creates the initial snapshot. Called only from `New()`.
  - `snapshotUpdate(instance, type, changes)` — creates subsequent snapshots. Called only
    from `Update()`. Accepts only `LineSnapshotTypeChange` or `LineSnapshotTypeNoChange`.
- **FR-3.4** Because `snapshotInitial` is only called from `New()`, and `snapshotUpdate`
  does not accept `LineSnapshotTypeInitial`, a second initial snapshot cannot be created.
  There is no `hasInitialSnapshot` flag and no runtime guard — the invariant is enforced
  by the API surface, not by conditional checks.

### FR-4: Updating (Snapshotting)

- **FR-4.1** `Update(newValue)` can be called on any `*Instance` without precondition checks,
  because the initial snapshot is guaranteed to exist (see FR-2).
- **FR-4.2** If the new value is identical to the latest snapshot (compared by hash), a
  `LineSnapshotTypeNoChange` snapshot must be created with `nil` changes.
- **FR-4.3** If the new value differs from the latest snapshot, a `LineSnapshotTypeChange`
  snapshot must be created with a computed change map.
- **FR-4.4** After a successful update, `Instance.value` must reflect the new value.

### FR-5: Change Detection (Diff)

- **FR-5.1** Changes between two source versions must be computed using a Longest Common
  Subsequence (LCS) algorithm at the line level.
- **FR-5.2** Each line in the new version must be classified as one of:
  - **unchanged** — the line exists at the same logical position in both versions.
  - **expanding** — the line is new (inserted/added) in the new version.
  - **contracting** — the line was removed from the old version.
- **FR-5.3** Each `LineChange` must record:
  - The change type (`unchanged`, `expanding`, `contracting`).
  - The origin line number (index in the previous version).
  - For expanding changes: the range and content of inserted lines.
  - For contracting changes: the range and content of removed lines.

### FR-6: Line Origin Tracing

- **FR-6.1** `LineOrigin(lineNumber)` must trace a line number in the latest (processed)
  source back through all change snapshots to the original line number in the initial
  snapshot.
- **FR-6.2** For unchanged lines, the origin must be the corresponding line index in the
  previous snapshot.
- **FR-6.3** For lines that were inserted during pre-processing (expanding), `LineOrigin()`
  must return `-1` to indicate the line has no origin in the initial source.
- **FR-6.4** For lines that were removed (contracting), `LineOrigin()` must return `-1`.
- **FR-6.5** If the history is empty, `LineOrigin()` must return `-1`.

### FR-7: Line History

- **FR-7.1** `LineHistory(lineNumber)` must return a slice of `LineChange` entries describing
  how a specific line evolved across all snapshots (e.g. unchanged → expanded → unchanged).
- **FR-7.2** The history must walk backwards through snapshots, following the origin chain.

### FR-8: Snapshot Hashing

- **FR-8.1** Each snapshot must store a SHA-256 hash of its source content.
- **FR-8.2** Hash comparison must be used for fast equality checks between snapshots
  (`SourceCompare`, `snapshotHashCompare`).

### FR-9: History Management

- **FR-9.1** The `History` must store an ordered list of `LinesSnapshot` entries.
- **FR-9.2** Only one `LineSnapshotTypeInitial` snapshot may exist. This is enforced
  structurally by the API (see FR-3.3 / FR-3.4), not by runtime checks.
- **FR-9.3** `latest()` must return the most recent snapshot, or `nil` if the history is empty.
- **FR-9.4** `empty()` / `notEmpty()` must accurately report whether the history contains
  any snapshots.

### FR-10: Accessor Methods

- **FR-10.1** `Value()` must return the current source string of the instance.
- **FR-10.2** `Lines()` must return the lines from the latest snapshot, or `nil` if no
  snapshots exist.
- **FR-10.3** `SnapshotCount()` must return the total number of snapshots in the history.
- **FR-10.4** `LatestSnapshot()` must return a pointer to the most recent snapshot.

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

