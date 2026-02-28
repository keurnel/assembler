package dependency_graph

type DependencyGraphEdge struct {
	// type - the type of dependency (e.g., "include", "import", etc.)
	dependencyType string
	// from - the source node of the edge
	from *DependencyGraphNode
	// to - the destination node of the edge
	to *DependencyGraphNode
}

// DependencyGraphEdgeNew - creates a new dependency graph edge with the given type, source node, and destination node.
func DependencyGraphEdgeNew(dependencyType string, from, to *DependencyGraphNode) *DependencyGraphEdge {
	return &DependencyGraphEdge{
		dependencyType: dependencyType,
		from:           from,
		to:             to,
	}
}

type DependencyGraphNode struct {
	// name - the file name or module name associated with this node
	name string
	// source - the source code associated with this node
	source string
	// edges - list of edges from this node to other nodes in the graph
	edges []*DependencyGraphEdge
}

// DependencyGraphNodeNew - creates a new dependency graph node with the given name and source code.
func DependencyGraphNodeNew(name, source string) *DependencyGraphNode {
	return &DependencyGraphNode{
		name:   name,
		source: source,
		edges:  []*DependencyGraphEdge{},
	}
}

// AddEdge - adds an edge from this node to another node in the graph.
func (n *DependencyGraphNode) AddEdge(edge *DependencyGraphEdge) {
	n.edges = append(n.edges, edge)
}

func (n *DependencyGraphNode) isCyclic(visited, recStack map[string]bool) bool {
	if !visited[n.name] {
		// Mark the current node as visited and part of the recursion stack
		visited[n.name] = true
		recStack[n.name] = true

		// Recur for all the vertices adjacent to this vertex
		for _, edge := range n.edges {
			if !visited[edge.to.name] && edge.to.isCyclic(visited, recStack) {
				return true
			} else if recStack[edge.to.name] {
				return true
			}
		}
	}

	// Remove the vertex from recursion stack
	recStack[n.name] = false
	return false
}
