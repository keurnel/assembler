package dependency_graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	OsStat     = os.Stat
	OsReadFile = os.ReadFile
)

type InstanceMetaData struct {
}

type Instance struct {
	// metaData - metadata about the graph (e.g., creation time, source file)
	metaData *InstanceMetaData

	// cwd - the current working directory for resolving relative paths
	cwd string
	// rootFilePath - absolute path of the root source file
	rootFilePath string
	// source - original source code
	source string
	// nodes - map of nodes in the graph
	nodes map[string]*DependencyGraphNode
}

// New - creates a new instance of the dependency graph.
// rootFilePath is the absolute path of the top-level source file; it is added
// as a node so that cycles involving the root are reported starting from it.
// Pass an empty string to omit the root node (e.g. in tests that build graphs
// programmatically).
func New(source, cwd, rootFilePath string) *Instance {
	instance := Instance{
		metaData:     &InstanceMetaData{},
		cwd:          cwd,
		rootFilePath: rootFilePath,
		source:       source,
		nodes:        make(map[string]*DependencyGraphNode),
	}

	// -- FR-1.1.1 - dependency graph receives working directory
	//
	err := instance.validWorkingDirectory()
	if err != nil {
		panic(err)
	}

	instance.build()

	return &instance
}

// Nodes - returns the nodes in the dependency graph.
func (i *Instance) Nodes() map[string]*DependencyGraphNode {
	return i.nodes
}

// validWorkingDirectory - validates the provided working directory path.
func (i *Instance) validWorkingDirectory() error {
	stat, err := OsStat(i.cwd)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("working directory does not exist: %s", i.cwd)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied for working directory: %s", i.cwd)
		}

		// Other unexpected errors.
		return fmt.Errorf("invalid working directory: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("working directory is not a directory: %s", i.cwd)
	}

	return nil
}

// AddNode - adds a node to the dependency graph.
func (i *Instance) AddNode(node *DependencyGraphNode) {
	if _, exists := i.nodes[node.name]; exists {
		// Node already exists, skip adding.
		return
	}
	i.nodes[node.name] = node
}

// build - builds the dependency graph from the source code by scanning for
// %include directives and recursively resolving nested dependencies (FR-4, FR-5).
// When rootFilePath is set, a node is created for the root file so that cycles
// involving it are reported starting from the root.
func (i *Instance) build() {
	var rootNode *DependencyGraphNode
	if i.rootFilePath != "" {
		rootNode = DependencyGraphNodeNew(i.rootFilePath, i.source)
		i.AddNode(rootNode)
	}
	i.scanSource(i.source, rootNode)
}

// scanSource recursively scans the given source for %include directives,
// creates nodes and edges, and recurses into included files. The parentNode
// is nil for the top-level source (FR-1.4: the root source is not added as
// a named node; only included files become nodes).
func (i *Instance) scanSource(source string, parentNode *DependencyGraphNode) {
	lines := strings.Split(source, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// FR-4.3: Skip non-include lines.
		if !strings.HasPrefix(line, "%include") {
			continue
		}

		parts := strings.Fields(line)

		// FR-4.4: Skip malformed directives without a path.
		if len(parts) < 2 {
			continue
		}

		if parts[0] != "%include" {
			continue
		}

		// FR-4.5: Strip surrounding double quotes from the file path.
		rawPath := parts[1]
		if len(rawPath) >= 2 && rawPath[0] == '"' && rawPath[len(rawPath)-1] == '"' {
			rawPath = rawPath[1 : len(rawPath)-1]
		}

		// FR-5.4: Only .kasm files may appear as include targets.
		if !strings.HasSuffix(rawPath, ".kasm") {
			panic(fmt.Sprintf("dependency graph error: included file '%s' is not a .kasm file", rawPath))
		}

		// FR-6.1: Resolve the file path relative to the graph's cwd.
		resolvedPath := rawPath
		if !filepath.IsAbs(rawPath) {
			resolvedPath = filepath.Join(i.cwd, rawPath)
		}

		// FR-3.2: Check if a node for this path already exists (shared dependency).
		existingNode, alreadyExists := i.nodes[resolvedPath]

		if !alreadyExists {
			// FR-6.2: Read file content.
			contentBytes, err := OsReadFile(resolvedPath)
			if err != nil {
				panic(fmt.Sprintf("dependency graph error: failed to read file '%s': %v", resolvedPath, err))
			}

			// FR-6.3 / FR-3.4: Create a new node with the file content.
			existingNode = DependencyGraphNodeNew(resolvedPath, string(contentBytes))
			i.AddNode(existingNode)
		}

		// FR-4.6.4 / FR-7: Create a directed edge from parent to included file.
		if parentNode != nil {
			edge := DependencyGraphEdgeNew("include", parentNode, existingNode)
			parentNode.AddEdge(edge)
		}

		// FR-5.1/FR-5.2: Recursively scan the included file's content for
		// its own %include directives (depth-first). Skip if the node was
		// already processed (shared dependency — FR-5.3).
		if !alreadyExists {
			i.scanSource(existingNode.source, existingNode)
		}
	}
}

// Acyclic - checks if the graph is acyclic.
func (i *Instance) Acyclic() bool {
	visited := make(map[string]bool, len(i.nodes))
	recStack := make(map[string]bool, len(i.nodes))

	for nodeName := range i.nodes {
		if !visited[nodeName] {
			if i.cyclic(nodeName, visited, recStack) {
				return false
			}
		}
	}

	return true
}

// cyclic - performs DFS to detect cycles in the graph.
func (i *Instance) cyclic(nodeName string, visited, recStack map[string]bool) bool {
	visited[nodeName] = true
	recStack[nodeName] = true

	for _, edge := range i.nodes[nodeName].edges {
		target := edge.to.name
		if recStack[target] {
			return true
		}
		if !visited[target] {
			if i.cyclic(target, visited, recStack) {
				return true
			}
		}
	}

	recStack[nodeName] = false
	return false
}

// CyclePath returns the ordered list of node names that form the first cycle
// found during DFS traversal, or nil if the graph is acyclic (FR-11.3.1).
// The last element connects back to the first, closing the cycle.
// Example: for A → B → C → A, the return value is ["A", "B", "C", "A"].
func (i *Instance) CyclePath() []string {
	visited := make(map[string]bool, len(i.nodes))
	recStack := make(map[string]bool, len(i.nodes))

	for nodeName := range i.nodes {
		if !visited[nodeName] {
			if path := i.cyclicWithPath(nodeName, visited, recStack, nil); path != nil {
				return path
			}
		}
	}

	return nil
}

// cyclicWithPath performs DFS to detect cycles, returning the cycle path when
// found. Shares the same DFS logic as cyclic (FR-11.3.2) but tracks the
// traversal path to report which nodes form the cycle.
func (i *Instance) cyclicWithPath(nodeName string, visited, recStack map[string]bool, path []string) []string {
	visited[nodeName] = true
	recStack[nodeName] = true
	path = append(path, nodeName)

	for _, edge := range i.nodes[nodeName].edges {
		target := edge.to.name
		if recStack[target] {
			// Found a cycle. Extract the cycle from the path.
			cycleStart := -1
			for idx, name := range path {
				if name == target {
					cycleStart = idx
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path[cycleStart:])+1)
				copy(cycle, path[cycleStart:])
				cycle[len(cycle)-1] = target // close the cycle
				return cycle
			}
			return []string{target}
		}
		if !visited[target] {
			if result := i.cyclicWithPath(target, visited, recStack, path); result != nil {
				return result
			}
		}
	}

	recStack[nodeName] = false
	return nil
}

// String returns a plain-text, tree-style representation of the dependency
// graph suitable for terminal output and log files (FR-11.1).
func (i *Instance) String() string {
	// FR-11.1.3: Empty graph.
	if len(i.nodes) == 0 {
		return "(empty graph)"
	}

	// Find root nodes: nodes that are not the target of any edge.
	targets := make(map[string]bool, len(i.nodes))
	for _, node := range i.nodes {
		for _, edge := range node.edges {
			targets[edge.to.name] = true
		}
	}

	roots := make([]string, 0)
	for name := range i.nodes {
		if !targets[name] {
			roots = append(roots, name)
		}
	}

	// If all nodes are targets (e.g. a pure cycle with no root), list all.
	if len(roots) == 0 {
		for name := range i.nodes {
			roots = append(roots, name)
		}
	}

	// FR-11.5.2: Sort roots for deterministic output across calls.
	sort.Strings(roots)

	var sb strings.Builder
	expanded := make(map[string]bool, len(i.nodes))

	for idx, rootName := range roots {
		if idx > 0 {
			sb.WriteByte('\n')
		}
		i.writeTree(&sb, rootName, "", expanded)
	}

	return sb.String()
}

// writeTree recursively writes the tree representation for a single node.
func (i *Instance) writeTree(sb *strings.Builder, nodeName, prefix string, expanded map[string]bool) {
	node, exists := i.nodes[nodeName]
	if !exists {
		return
	}

	// FR-11.1.2: Mark shared dependencies.
	if expanded[nodeName] {
		sb.WriteString(nodeName)
		sb.WriteString(" (shared)\n")
		return
	}

	sb.WriteString(nodeName)
	sb.WriteByte('\n')
	expanded[nodeName] = true

	for idx, edge := range node.edges {
		childIsLast := idx == len(node.edges)-1

		// Write the connector prefix.
		sb.WriteString(prefix)
		if childIsLast {
			sb.WriteString("└── ")
		} else {
			sb.WriteString("├── ")
		}

		// Compute the prefix for the child's subtree.
		childPrefix := prefix
		if childIsLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		i.writeTree(sb, edge.to.name, childPrefix, expanded)
	}
}

// ToDot produces a Graphviz DOT representation of the dependency graph
// (FR-11.2). Cycle edges are highlighted in red when the graph is cyclic
// (FR-11.2.4).
func (i *Instance) ToDot() string {
	var sb strings.Builder

	sb.WriteString("digraph dependencies {\n")

	// FR-11.2.5: Empty graph produces a valid but empty digraph.
	if len(i.nodes) == 0 {
		sb.WriteByte('}')
		return sb.String()
	}

	// FR-11.2.4: Detect back-edges for cycle highlighting.
	backEdges := make(map[string]bool)
	if !i.Acyclic() {
		visited := make(map[string]bool, len(i.nodes))
		recStack := make(map[string]bool, len(i.nodes))
		for nodeName := range i.nodes {
			if !visited[nodeName] {
				i.collectBackEdges(nodeName, visited, recStack, backEdges)
			}
		}
	}

	// FR-11.2.2: Emit nodes. Sorted for deterministic output (FR-11.5.2).
	sortedNames := make([]string, 0, len(i.nodes))
	for name := range i.nodes {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		sb.WriteString(fmt.Sprintf("  %q;\n", name))
	}

	// FR-11.2.3: Emit edges. Iterated in sorted node order (FR-11.5.2).
	for _, name := range sortedNames {
		node := i.nodes[name]
		for _, edge := range node.edges {
			edgeKey := edge.from.name + " -> " + edge.to.name
			if backEdges[edgeKey] {
				sb.WriteString(fmt.Sprintf("  %q -> %q [label=%q, color=red];\n",
					edge.from.name, edge.to.name, edge.dependencyType))
			} else {
				sb.WriteString(fmt.Sprintf("  %q -> %q [label=%q];\n",
					edge.from.name, edge.to.name, edge.dependencyType))
			}
		}
	}

	sb.WriteByte('}')
	return sb.String()
}

// collectBackEdges performs DFS and records edges that form back-edges
// (edges to nodes currently on the recursion stack), used for cycle
// highlighting in DOT output (FR-11.2.4).
func (i *Instance) collectBackEdges(nodeName string, visited, recStack map[string]bool, backEdges map[string]bool) {
	visited[nodeName] = true
	recStack[nodeName] = true

	for _, edge := range i.nodes[nodeName].edges {
		target := edge.to.name
		if recStack[target] {
			edgeKey := edge.from.name + " -> " + edge.to.name
			backEdges[edgeKey] = true
		} else if !visited[target] {
			i.collectBackEdges(target, visited, recStack, backEdges)
		}
	}

	recStack[nodeName] = false
}
