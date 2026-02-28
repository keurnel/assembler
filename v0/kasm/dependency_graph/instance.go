package dependency_graph

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

var (
	OsStat = os.Stat
)

type Instance struct {
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
		cwd:    cwd,
		source: source,
		nodes:  make(map[string]*DependencyGraphNode),
	}

	// -- FR-1.1.1 - dependency graph receives working directory
	//
	err := instance.validWorkingDirectory()
	if err != nil {
		panic(err)
	}

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

// Acyclic - checks if the graph is acyclic.
func (i *Instance) Acyclic() bool {
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		isCyclic atomic.Bool
	)

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeName := range i.nodes {
		mu.Lock()
		alreadyVisited := visited[nodeName]
		mu.Unlock()

		if alreadyVisited || isCyclic.Load() {
			continue
		}

		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			mu.Lock()
			if visited[name] {
				mu.Unlock()
				return
			}
			if i.Cyclic(name, visited, recStack) {
				isCyclic.Store(true)
			}
			mu.Unlock()
		}(nodeName)
	}

	wg.Wait()
	return !isCyclic.Load()
}

// Cyclic - helper function for Acyclic to perform DFS and detect cycles in the graph.
func (i *Instance) Cyclic(nodeName string, visited, recStack map[string]bool) bool {
	visited[nodeName] = true
	recStack[nodeName] = true

	node := i.nodes[nodeName]
	for _, edge := range node.edges {
		if !visited[edge.to.name] {
			if i.Cyclic(edge.to.name, visited, recStack) {
				return true
			}
		} else if recStack[edge.to.name] {
			return true
		}
	}

	recStack[nodeName] = false
	return false
}
