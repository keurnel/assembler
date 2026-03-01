package dependency_graph

import (
	"fmt"
	"os"
	"strings"
)

var (
	OsStat = os.Stat
)

type InstanceMetaData struct {
}

type Instance struct {
	// metaData - metadata about the graph (e.g., creation time, source file)
	metaData *InstanceMetaData

	// cwd - the current working directory for resolving relative paths
	cwd string
	// source - original source code
	source string
	// nodes - map of nodes in the graph
	nodes map[string]*DependencyGraphNode
}

// New - creates a new instance of the dependency graph.
func New(source, cwd string) *Instance {
	instance := Instance{
		metaData: &InstanceMetaData{},
		cwd:      cwd,
		source:   source,
		nodes:    make(map[string]*DependencyGraphNode),
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

// build - builds the dependency graph from the source code.
func (i *Instance) build() {

	type IncludeDirective struct {
		relativePath string
	}

	// Split lines
	//
	lines := strings.Split(i.source, "\n")

	// Are the first lines include directives?
	//
	for _, line := range lines {

		line = strings.TrimSpace(line)

		// Skip non-include lines.
		//
		if !strings.HasPrefix(line, "%include") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			// Invalid include directive, skip.
			continue
		}

		directive := parts[0]
		if directive != "%include" {
			continue
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
