# Dependency Graph

The dependency graph is a directed acyclic graph (DAG) that represents the
`%include` relationships between `.kasm` source files. It is built by the
pre-processor's Phase 1 (Includes) before any file content is inlined, so that
structural problems — circular dependencies, missing files, invalid paths — are
caught before source transformation begins.

The dependency graph lives in the Go package `v0/kasm/dependency_graph`. It is
consumed by the orchestrator (`cmd/cli/cmd/x86_64/assemble_file.go`) which
creates the graph, validates it, and only then proceeds to inline included
content via `PreProcessingHandleIncludes`.

Related documents:

- [Pre-Processor Requirements](requirements.md) — FR-1 (Includes), FR-1.5
  (Recursive Includes), FR-1.6 (Circular Include Detection).

---

## Data Model

### Instance

The top-level graph object. Created via `New(source, cwd)`.

```
Instance {
    metaData  *InstanceMetaData            // Reserved for future metadata (creation time, source file, etc.).
    cwd       string                       // Working directory for resolving relative include paths.
    source    string                       // Original top-level source code.
    nodes     map[string]*DependencyGraphNode  // All nodes in the graph, keyed by file path / name.
}
```

### InstanceMetaData

Reserved for future use. Currently empty.

```
InstanceMetaData {}
```

### DependencyGraphNode

Represents a single file in the dependency graph.

```
DependencyGraphNode {
    name    string                    // File path or identifier for the node.
    source  string                    // Source code content of the file.
    edges   []*DependencyGraphEdge    // Outgoing edges (files this node includes).
}
```

### DependencyGraphEdge

Represents a directed dependency from one file to another.

```
DependencyGraphEdge {
    dependencyType  string                // Type of the dependency (e.g. "include").
    from            *DependencyGraphNode  // Source node (the file that contains the directive).
    to              *DependencyGraphNode  // Target node (the file being included).
}
```

---

## FR-1: Graph Construction

`New(source, cwd) → *Instance`

Constructs a new dependency graph from the given top-level source string and
working directory. The constructor validates the working directory (FR-2),
then builds the graph by scanning for `%include` directives (FR-4).

- **FR-1.1** `New` is the sole public constructor for `Instance`. It returns a
  fully constructed graph ready for acyclicity checking and traversal.
- **FR-1.2** The constructor must panic if the working directory is invalid
  (see FR-2). This is a fatal error — the graph cannot be built without a
  valid base path.
- **FR-1.3** After validation, the constructor calls `build()` to scan the
  source for `%include` directives and populate the graph with nodes and edges.
- **FR-1.4** The root source (the content passed to `New`) is not itself added
  as a named node. The graph represents only the included files and their
  relationships. The orchestrator is responsible for seeding the root file
  path into any seen-set used for cross-invocation cycle detection (see
  pre-processor requirements FR-1.6.6).

---

## FR-2: Working Directory Validation

Before the graph is built, the provided `cwd` must be validated. The graph
cannot resolve relative `%include` paths without a valid working directory.

- **FR-2.1** If the working directory does not exist (`os.IsNotExist`), the
  constructor must panic with a message containing the path and the phrase
  `"working directory does not exist"`.
- **FR-2.2** If the working directory cannot be accessed due to permissions
  (`os.IsPermission`), the constructor must panic with a message containing
  the path and the phrase `"permission denied"`.
- **FR-2.3** If the path exists but is not a directory, the constructor must
  panic with a message containing the path and the phrase `"is not a
  directory"`.
- **FR-2.4** Any other unexpected `os.Stat` error must cause a panic wrapping
  the original error.
- **FR-2.5** The `os.Stat` call must be performed through the package-level
  variable `OsStat` (defaulting to `os.Stat`) so that tests can inject a mock
  without touching the real filesystem.

---

## FR-3: Node Management

Nodes represent files in the dependency graph. Each node is uniquely identified
by its name (file path).

- **FR-3.1** `AddNode(node)` adds a node to the graph's internal map, keyed by
  the node's name.
- **FR-3.2** If a node with the same name already exists, `AddNode` must be a
  no-op — the existing node is kept, and the new one is discarded. This
  ensures shared dependencies (files included by multiple parents) result in
  a single node (see FR-5.3).
- **FR-3.3** `Nodes()` returns the full `map[string]*DependencyGraphNode` for
  read access by the orchestrator and tests. The caller must not mutate the
  returned map.
- **FR-3.4** `DependencyGraphNodeNew(name, source)` creates a new node with an
  empty edge slice. It does not add the node to any graph — the caller must
  call `AddNode` explicitly.

---

## FR-4: Include Directive Scanning (`build`)

The `build()` method scans the source for `%include` directives and populates
the graph with nodes and edges.

- **FR-4.1** The method splits the source into lines and processes each line
  independently.
- **FR-4.2** A line is recognised as an `%include` directive if, after trimming
  leading/trailing whitespace, it starts with `%include` followed by
  whitespace and a quoted file path: `%include "path/to/file.kasm"`.
- **FR-4.3** Lines that do not start with `%include` (after trimming) must be
  silently skipped.
- **FR-4.4** If a `%include` line has fewer than two whitespace-delimited parts
  (i.e. `%include` without a path), it must be silently skipped.
- **FR-4.5** The file path is extracted from the second field. Surrounding
  double quotes must be stripped before use.
- **FR-4.6** For each valid `%include` directive, the method must:
  1. Resolve the file path relative to the graph's `cwd`.
  2. Read the file content (see FR-6).
  3. Create a `DependencyGraphNode` for the included file (if one does not
     already exist).
  4. Create a `DependencyGraphEdge` of type `"include"` from the current
     context node to the included file node.
  5. Recursively scan the included file's content for its own `%include`
     directives (see FR-5).

---

## FR-5: Recursive Resolution

The dependency graph must be built recursively so that nested `%include`
directives (includes within included files) are fully resolved.

- **FR-5.1** When a file is included and its content is read, the graph must
  scan that content for additional `%include` directives and resolve those as
  well, building a complete graph of all transitive dependencies.
- **FR-5.2** The graph must be built in a depth-first manner: nested
  dependencies are resolved before their parents. This ensures that the graph
  structure reflects the true inclusion order.
- **FR-5.3** Shared dependencies — files included by multiple parent files —
  must result in a single `DependencyGraphNode`. Multiple edges may point to
  the same node, but the node (and its content) exists only once. This is
  enforced by the `AddNode` deduplication (FR-3.2).
- **FR-5.4** Only `.kasm` files may appear as include targets. If a
  non-`.kasm` path is encountered during recursive scanning, the graph must
  panic with a message containing the offending file path.

---

## FR-6: File Resolution & Reading

The graph must resolve relative paths and read file content to build nodes for
included files.

- **FR-6.1** Include paths are resolved relative to the graph's `cwd` using
  `filepath.Join(cwd, relativePath)`.
- **FR-6.2** File content is read via `os.ReadFile`. If the file cannot be read
  (not found, permission denied, etc.), the graph must panic with a message
  containing the file path and the underlying error. This satisfies the
  requirement that unresolvable dependencies are fatal errors.
- **FR-6.3** The read content is stored in the `DependencyGraphNode.source`
  field and is available for recursive scanning (FR-5.1).

---

## FR-7: Edge Management

Edges represent directed `%include` relationships between files.

- **FR-7.1** `DependencyGraphEdgeNew(dependencyType, from, to)` creates a new
  edge. The `dependencyType` field identifies the kind of dependency (e.g.
  `"include"`).
- **FR-7.2** `AddEdge(edge)` on a `DependencyGraphNode` appends the edge to
  the node's outgoing edge list. There is no deduplication of edges — if a
  file includes the same target twice, two edges are created (this is caught
  as a validation error by the pre-processor's duplicate detection, FR-1.2.2
  in the pre-processor requirements).
- **FR-7.3** Edges are directional: `from` is the file containing the
  `%include` directive, `to` is the file being included.

---

## FR-8: Acyclicity Checking

`Acyclic() → bool`

Determines whether the dependency graph is a DAG (no cycles). This is the
primary validation the orchestrator performs before proceeding with include
inlining.

- **FR-8.1** `Acyclic()` returns `true` if the graph contains no cycles,
  `false` otherwise.
- **FR-8.2** Cycle detection uses depth-first search (DFS) with a visited set
  and a recursion stack. A node is on the recursion stack while its subtree is
  being explored; if a back-edge to a node on the recursion stack is found, a
  cycle exists.
- **FR-8.3** The method must correctly handle:
  - **Empty graph** (no nodes) — acyclic.
  - **Graph with no edges** (isolated nodes) — acyclic.
  - **Linear chain** (A → B → C) — acyclic.
  - **Diamond DAG** (A → B, A → C, B → D, C → D) — acyclic.
  - **Simple cycle** (A → B → C → A) — cyclic.
  - **Self-loop** (A → A) — cyclic.
  - **Disconnected components** — each component is checked independently.
- **FR-8.4** The visited and recursion-stack maps must be pre-allocated with
  `len(nodes)` capacity to avoid map growth during traversal.
- **FR-8.5** The method must return as soon as a cycle is found (early exit)
  rather than traversing the entire graph.

### FR-8.6: DFS Helper (`cyclic`)

`cyclic(nodeName, visited, recStack) → bool`

Internal recursive helper that performs the DFS traversal for a single
connected component.

- **FR-8.6.1** Marks the current node as visited and adds it to the recursion
  stack.
- **FR-8.6.2** For each outgoing edge, checks whether the target is already on
  the recursion stack (cycle found) or has not been visited (recurse).
- **FR-8.6.3** On return (no cycle found in subtree), removes the current node
  from the recursion stack.
- **FR-8.6.4** Returns `true` if a cycle is detected in the subtree rooted at
  the given node, `false` otherwise.

---

## FR-9: Error Handling

All errors during graph construction and validation are reported via `panic`.
The orchestrator is responsible for recovering from panics and translating them
into user-facing diagnostics via `debugcontext`.

- **FR-9.1** Working directory validation errors must panic with a descriptive
  message containing the path (FR-2).
- **FR-9.2** File read errors during recursive resolution must panic with a
  message containing the file path and the underlying I/O error (FR-6.2).
- **FR-9.3** Non-`.kasm` include paths must panic with a message containing
  the offending path (FR-5.4).
- **FR-9.4** Cycle detection does **not** panic — it returns a boolean. The
  orchestrator decides how to report the error (typically via
  `debugcontext.Error`). See pre-processor requirements FR-1.6.

---

## FR-10: Testability

- **FR-10.1** The package-level variable `OsStat` (defaulting to `os.Stat`)
  must be used for all filesystem stat calls so that tests can inject mock
  implementations without touching the real filesystem.
- **FR-10.2** `DependencyGraphNodeNew` and `DependencyGraphEdgeNew` are public
  constructors so that tests can build arbitrary graph topologies for
  acyclicity checking without requiring real files.
- **FR-10.3** `AddNode` and `AddEdge` are public so that tests can construct
  graphs programmatically.

---

## FR-11: Visualization

The dependency graph must be renderable into human-readable representations so
that developers can inspect, debug, and document the include structure of a
project. Visualization is a read-only operation — it must never mutate the
graph.

### FR-11.1: Text Representation (`String`)

`String() → string`

Produces a plain-text, tree-style representation of the dependency graph
suitable for terminal output and log files.

- **FR-11.1.1** The method must return a multi-line string that shows every node
  and its outgoing edges in an indented tree format. The root level lists all
  top-level nodes (nodes that are not the target of any edge). Each child is
  indented with two spaces per depth level.
  ```
  main.kasm
    ├── io.kasm
    │   └── constants.kasm
    └── math.kasm
        └── constants.kasm (shared)
  ```
- **FR-11.1.2** Shared dependencies (nodes reached via multiple parents) must be
  annotated with `(shared)` on second and subsequent appearances to indicate
  that the node has already been expanded elsewhere in the tree. The subtree
  must not be repeated — only the node name and the `(shared)` marker are
  printed.
- **FR-11.1.3** If the graph is empty (no nodes), the method must return the
  string `"(empty graph)"`.
- **FR-11.1.4** Each node must be displayed using its `name` field (the file
  path). No path normalisation or truncation is applied — the name is printed
  as-is.
- **FR-11.1.5** Edge type labels (e.g. `"include"`) are not displayed in the
  text representation. The tree structure already implies inclusion.

### FR-11.2: DOT Format Export (`ToDot`)

`ToDot() → string`

Produces a [Graphviz DOT](https://graphviz.org/doc/info/lang.html) representation
of the dependency graph, suitable for rendering with `dot`, `neato`, or
compatible tools.

- **FR-11.2.1** The output must be a valid DOT `digraph` with the name
  `"dependencies"`.
- **FR-11.2.2** Each `DependencyGraphNode` must be emitted as a DOT node. The
  node identifier is the node's `name` field, quoted to handle paths containing
  special characters.
- **FR-11.2.3** Each `DependencyGraphEdge` must be emitted as a directed edge
  (`->`) from the `from` node to the `to` node. The edge label must be the
  `dependencyType` value (e.g. `"include"`).
- **FR-11.2.4** If `Acyclic()` returns `false`, edges that form part of a cycle
  must be annotated with `[color=red]` so that cycles are visually distinct
  when rendered. Detection of which specific edges participate in a cycle is
  performed by running a second DFS pass that records back-edges.
- **FR-11.2.5** If the graph is empty, the method must return a valid but empty
  digraph:
  ```dot
  digraph dependencies {
  }
  ```
- **FR-11.2.6** DOT output must use `\n` (LF) line endings and must not include
  a trailing newline after the closing `}`.

### FR-11.3: Cycle Path Reporting

When a cycle is detected, the visualization layer must be able to report the
specific nodes involved in the cycle so that error messages and visual output
can pinpoint the problem.

- **FR-11.3.1** A new method `CyclePath() → []string` must return the ordered
  list of node names that form the first cycle found during DFS traversal, or
  `nil` if the graph is acyclic. The last element connects back to the first,
  closing the cycle.
  Example: for A → B → C → A, the return value is `["A", "B", "C", "A"]`.
- **FR-11.3.2** `CyclePath` must share the same DFS logic as `Acyclic` / `cyclic`
  (FR-8) to guarantee consistent results. It must not perform a separate,
  potentially divergent traversal.
- **FR-11.3.3** The orchestrator may use `CyclePath` to enrich the error message
  recorded via `debugcontext.Error` when a circular inclusion is detected,
  listing the full chain of files involved.

### FR-11.4: Integration with Orchestrator & Verbose Mode

The orchestrator already supports a `--verbose` (`-v`) flag that prints debug
context entries. Visualization output must integrate with this mechanism.

- **FR-11.4.1** When verbose mode is enabled, the orchestrator must log the
  text representation of the dependency graph (`String()`) via
  `debugCtx.Trace` after the graph is constructed and validated. This allows
  developers to see the full include tree in the console output.
- **FR-11.4.2** The DOT representation is not emitted to the console by default.
  It is available programmatically via `ToDot()` for tooling, IDE integrations,
  or future CLI sub-commands (e.g. `keurnel-asm dependency-graph --format dot`).
- **FR-11.4.3** Visualization methods must not perform I/O. They return strings;
  the caller decides where to write them (stdout, file, debug context).

### FR-11.5: Architectural Constraints

- **FR-11.5.1** `String()` and `ToDot()` are public methods on `*Instance`.
  `CyclePath()` is also public on `*Instance`.
- **FR-11.5.2** Visualization methods must be pure — they must not modify the
  graph's nodes, edges, or metadata. Calling `String()` or `ToDot()` multiple
  times must produce identical output for an unchanged graph.
- **FR-11.5.3** Visualization methods must not allocate new nodes or edges.
  Temporary state (visited sets, string builders) must be local to the method
  call.
- **FR-11.5.4** `String()` and `ToDot()` must use `strings.Builder` for output
  construction to avoid repeated string concatenation.

---

## NFR-1: Performance

- **NFR-1.1** `Acyclic()` must complete in O(V + E) time where V is the number
  of nodes and E is the number of edges. The DFS-based algorithm satisfies
  this.
- **NFR-1.2** Visited and recursion-stack maps must be pre-allocated to avoid
  rehashing during traversal (FR-8.4).
- **NFR-1.3** The graph should handle dependency trees of at least 1 000 nodes
  without noticeable latency. The existing benchmark (`BenchmarkAcyclic`)
  validates this.

---

## Integration with Orchestrator

The orchestrator (`cmd/cli/cmd/x86_64/assemble_file.go`) uses the dependency
graph as follows:

1. Obtains the working directory (`os.Getwd`).
2. Creates the graph: `dependency_graph.New(source, cwd)`.
3. Checks acyclicity: `graph.Acyclic()`.
4. If cyclic, records an error via `debugcontext.Error` and aborts include
   processing.
5. If acyclic, proceeds to call `PreProcessingHandleIncludes(source)` to
   inline file content.

Cross-invocation circular inclusion detection (the seen-set maintained by the
orchestrator across recursive invocations of `PreProcessingHandleIncludes`) is
**not** the responsibility of the dependency graph. It is specified in the
pre-processor requirements (FR-1.6.2 – FR-1.6.7).

