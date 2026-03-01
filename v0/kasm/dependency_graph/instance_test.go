package dependency_graph_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/keurnel/assembler/v0/kasm/dependency_graph"
)

type mockFileInfo struct {
	os.FileInfo
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }

// TestNewInstance - FR-1.1.1 dependency graph receives working directory.
func TestNewInstance(t *testing.T) {

	// Returns an instance when given a valid working directory.
	//
	t.Run("valid working directory", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		if instance == nil {
			t.Errorf("expected instance to be created for valid working directory, but got nil")
		}
	})

	// Panics when given an invalid working directory.
	//
	t.Run("invalid working directory", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for invalid working directory, but did not panic")
			}
			// Restore the original OsStat function after the test.
			dependency_graph.OsStat = os.Stat
		}()

		dependency_graph.New("source code", "/path/to/nonexistent/directory")
	})

	// Panics on permission denied working directory.
	//
	t.Run("permission denied working directory", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrPermission
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for permission denied working directory, but did not panic")
			}
			// Restore the original OsStat function after the test.
			dependency_graph.OsStat = os.Stat
		}()

		dependency_graph.New("source code", "/path/to/permission/denied/directory")
	})

	// Panics when given a path that is not a directory.
	//
	t.Run("working directory is not a directory", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: false}, nil
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for working directory that is not a directory, but did not panic")
			}
			// Restore the original OsStat function after the test.
			dependency_graph.OsStat = os.Stat
		}()

		dependency_graph.New("source code", "/path/to/file/instead/of/directory")
	})
}

// TestAcyclic - FR-1.1.2 dependency graph is acyclic.
func TestAcyclic(t *testing.T) {
	// Returns true for an acyclic graph.
	//
	t.Run("acyclic graph", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "source A")
		nodeB := dependency_graph.DependencyGraphNodeNew("B", "source B")
		nodeC := dependency_graph.DependencyGraphNodeNew("C", "source C")

		instance.AddNode(nodeA)
		instance.AddNode(nodeB)
		instance.AddNode(nodeC)

		edgeA := dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB)
		edgeB := dependency_graph.DependencyGraphEdgeNew("include", nodeB, nodeC)

		nodeA.AddEdge(edgeA)
		nodeB.AddEdge(edgeB)

		if !instance.Acyclic() {
			t.Errorf("expected graph to be acyclic, but got cyclic")
		}
	})

	// Returns true for random acyclic graph.
	//
	t.Run("random acyclic graph", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")

		// Amount random nodes and edges to create.
		//
		numNodes := 10
		numEdges := 15

		nodes := make([]*dependency_graph.DependencyGraphNode, numNodes)
		for i := 0; i < numNodes; i++ {
			nodes[i] = dependency_graph.DependencyGraphNodeNew(
				fmt.Sprintf("Node%d", i),
				fmt.Sprintf("source for node %d", i),
			)
			instance.AddNode(nodes[i])
		}

		// Create random edges while ensuring acyclicity.
		for i := 0; i < numEdges; i++ {
			fromIndex := i % numNodes
			toIndex := (i + 1) % numNodes // Ensure no cycles by connecting to the next node in a circular manner.
			edge := dependency_graph.DependencyGraphEdgeNew(
				fmt.Sprintf("dependency%d", i),
				nodes[fromIndex],
				nodes[toIndex],
			)
			nodes[fromIndex].AddEdge(edge)
		}

	})

	// Returns true for a graph with no edges.
	//
	t.Run("graph with no edges", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "source A")
		nodeB := dependency_graph.DependencyGraphNodeNew("B", "source B")
		nodeC := dependency_graph.DependencyGraphNodeNew("C", "source C")

		instance.AddNode(nodeA)
		instance.AddNode(nodeB)
		instance.AddNode(nodeC)

		if !instance.Acyclic() {
			t.Errorf("expected graph with no edges to be acyclic, but got cyclic")
		}
	})

	// Returns true for an empty graph.
	//
	t.Run("empty graph", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		if !instance.Acyclic() {
			t.Errorf("expected empty graph to be acyclic, but got cyclic")
		}
	})

	// Returns false for a cyclic graph.
	//
	t.Run("cyclic graph", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "source A")
		nodeB := dependency_graph.DependencyGraphNodeNew("B", "source B")
		nodeC := dependency_graph.DependencyGraphNodeNew("C", "source C")

		instance.AddNode(nodeA)
		instance.AddNode(nodeB)
		instance.AddNode(nodeC)

		edgeA := dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB)
		edgeB := dependency_graph.DependencyGraphEdgeNew("include", nodeB, nodeC)
		edgeC := dependency_graph.DependencyGraphEdgeNew("include", nodeC, nodeA)

		nodeA.AddEdge(edgeA)
		nodeB.AddEdge(edgeB)
		nodeC.AddEdge(edgeC)

		if instance.Acyclic() {
			t.Errorf("expected graph to be cyclic, but got acyclic")
		}
	})

	// Returns false for a graph with self-loop.
	//
	t.Run("graph with self-loop", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}

		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "source A")

		instance.AddNode(nodeA)

		edgeA := dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeA)
		nodeA.AddEdge(edgeA)

		if instance.Acyclic() {
			t.Errorf("expected graph with self-loop to be cyclic, but got acyclic")
		}
	})

	// FR-8.3: Returns true for a diamond DAG (A → B, A → C, B → D, C → D).
	//
	t.Run("diamond DAG", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}
		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
		nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
		nodeC := dependency_graph.DependencyGraphNodeNew("C", "")
		nodeD := dependency_graph.DependencyGraphNodeNew("D", "")

		instance.AddNode(nodeA)
		instance.AddNode(nodeB)
		instance.AddNode(nodeC)
		instance.AddNode(nodeD)

		nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))
		nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeC))
		nodeB.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeB, nodeD))
		nodeC.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeC, nodeD))

		if !instance.Acyclic() {
			t.Errorf("expected diamond DAG to be acyclic, but got cyclic")
		}
	})

	// FR-8.3: Returns true for disconnected components (each acyclic).
	//
	t.Run("disconnected components", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}
		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		// Component 1: A → B
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
		nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
		instance.AddNode(nodeA)
		instance.AddNode(nodeB)
		nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))

		// Component 2: C → D (disconnected from A-B)
		nodeC := dependency_graph.DependencyGraphNodeNew("C", "")
		nodeD := dependency_graph.DependencyGraphNodeNew("D", "")
		instance.AddNode(nodeC)
		instance.AddNode(nodeD)
		nodeC.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeC, nodeD))

		if !instance.Acyclic() {
			t.Errorf("expected disconnected acyclic components to be acyclic")
		}
	})

	// FR-8.3: Returns false when one disconnected component has a cycle.
	//
	t.Run("disconnected components with cycle", func(t *testing.T) {
		dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{isDir: true}, nil
		}
		defer func() {
			dependency_graph.OsStat = os.Stat
		}()

		instance := dependency_graph.New("source code", "/path/to/valid/directory")
		// Component 1: A → B (acyclic)
		nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
		nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
		instance.AddNode(nodeA)
		instance.AddNode(nodeB)
		nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))

		// Component 2: C → D → C (cyclic)
		nodeC := dependency_graph.DependencyGraphNodeNew("C", "")
		nodeD := dependency_graph.DependencyGraphNodeNew("D", "")
		instance.AddNode(nodeC)
		instance.AddNode(nodeD)
		nodeC.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeC, nodeD))
		nodeD.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeD, nodeC))

		if instance.Acyclic() {
			t.Errorf("expected cyclic when one disconnected component has a cycle")
		}
	})
}

// ---------------------------------------------------------------------------
// FR-4/FR-5/FR-6: build() — Include Directive Scanning & Recursive Resolution
// ---------------------------------------------------------------------------

// TestBuild_SingleInclude verifies that build() creates a node and populates
// the graph when the source contains a single %include directive.
func TestBuild_SingleInclude(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	dependency_graph.OsReadFile = func(name string) ([]byte, error) {
		return []byte("mov rax, 1"), nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
		dependency_graph.OsReadFile = os.ReadFile
	}()

	source := `%include "helper.kasm"`
	instance := dependency_graph.New(source, "/project")

	nodes := instance.Nodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if _, ok := nodes["/project/helper.kasm"]; !ok {
		t.Error("expected node for '/project/helper.kasm'")
	}
}

// TestBuild_MultipleIncludes verifies that build() creates nodes for multiple
// %include directives.
func TestBuild_MultipleIncludes(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	dependency_graph.OsReadFile = func(name string) ([]byte, error) {
		return []byte("nop"), nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
		dependency_graph.OsReadFile = os.ReadFile
	}()

	source := "%include \"a.kasm\"\n%include \"b.kasm\""
	instance := dependency_graph.New(source, "/project")

	nodes := instance.Nodes()
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

// TestBuild_RecursiveIncludes verifies FR-5: nested %include directives are
// resolved recursively (depth-first).
func TestBuild_RecursiveIncludes(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	dependency_graph.OsReadFile = func(name string) ([]byte, error) {
		switch name {
		case "/project/b.kasm":
			return []byte(`%include "c.kasm"`), nil
		case "/project/c.kasm":
			return []byte("nop"), nil
		default:
			return nil, fmt.Errorf("file not found: %s", name)
		}
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
		dependency_graph.OsReadFile = os.ReadFile
	}()

	source := `%include "b.kasm"`
	instance := dependency_graph.New(source, "/project")

	nodes := instance.Nodes()
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes (b.kasm and c.kasm), got %d", len(nodes))
	}
	if _, ok := nodes["/project/b.kasm"]; !ok {
		t.Error("expected node for '/project/b.kasm'")
	}
	if _, ok := nodes["/project/c.kasm"]; !ok {
		t.Error("expected node for '/project/c.kasm'")
	}
}

// TestBuild_SharedDependency verifies FR-5.3: shared dependencies result in
// a single node.
func TestBuild_SharedDependency(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	dependency_graph.OsReadFile = func(name string) ([]byte, error) {
		switch name {
		case "/project/a.kasm":
			return []byte(`%include "shared.kasm"`), nil
		case "/project/b.kasm":
			return []byte(`%include "shared.kasm"`), nil
		case "/project/shared.kasm":
			return []byte("nop"), nil
		default:
			return nil, fmt.Errorf("file not found: %s", name)
		}
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
		dependency_graph.OsReadFile = os.ReadFile
	}()

	source := "%include \"a.kasm\"\n%include \"b.kasm\""
	instance := dependency_graph.New(source, "/project")

	nodes := instance.Nodes()
	// a.kasm, b.kasm, shared.kasm — shared.kasm is only one node.
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
}

// TestBuild_NonKasmPanics verifies FR-5.4: non-.kasm include targets cause
// a panic.
func TestBuild_NonKasmPanics(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-.kasm include")
		}
		msg := fmt.Sprintf("%v", r)
		if !containsSubstring(msg, "not a .kasm file") {
			t.Errorf("unexpected panic message: %s", msg)
		}
		dependency_graph.OsStat = os.Stat
	}()

	source := `%include "module.asm"`
	dependency_graph.New(source, "/project")
}

// TestBuild_FileReadErrorPanics verifies FR-6.2: unreadable files cause a panic.
func TestBuild_FileReadErrorPanics(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	dependency_graph.OsReadFile = func(name string) ([]byte, error) {
		return nil, fmt.Errorf("permission denied")
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for file read error")
		}
		msg := fmt.Sprintf("%v", r)
		if !containsSubstring(msg, "failed to read file") {
			t.Errorf("unexpected panic message: %s", msg)
		}
		dependency_graph.OsStat = os.Stat
		dependency_graph.OsReadFile = os.ReadFile
	}()

	source := `%include "missing.kasm"`
	dependency_graph.New(source, "/project")
}

// TestBuild_NoIncludes verifies that build() creates no nodes when there are
// no %include directives.
func TestBuild_NoIncludes(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("mov rax, 1", "/project")
	if len(instance.Nodes()) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(instance.Nodes()))
	}
}

// TestBuild_SkipsMalformedDirective verifies FR-4.4: %include without a path
// is silently skipped.
func TestBuild_SkipsMalformedDirective(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	source := "%include\nmov rax, 1"
	instance := dependency_graph.New(source, "/project")
	if len(instance.Nodes()) != 0 {
		t.Errorf("expected 0 nodes for malformed directive, got %d", len(instance.Nodes()))
	}
}

// ---------------------------------------------------------------------------
// FR-11.3: CyclePath
// ---------------------------------------------------------------------------

// TestCyclePath_Acyclic verifies that CyclePath returns nil for acyclic graphs.
func TestCyclePath_Acyclic(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))

	path := instance.CyclePath()
	if path != nil {
		t.Errorf("expected nil for acyclic graph, got %v", path)
	}
}

// TestCyclePath_SimpleCycle verifies that CyclePath returns the correct cycle.
func TestCyclePath_SimpleCycle(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	nodeC := dependency_graph.DependencyGraphNodeNew("C", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	instance.AddNode(nodeC)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))
	nodeB.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeB, nodeC))
	nodeC.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeC, nodeA))

	path := instance.CyclePath()
	if path == nil {
		t.Fatal("expected cycle path, got nil")
	}
	if len(path) < 2 {
		t.Fatalf("expected cycle path with at least 2 elements, got %v", path)
	}
	// The last element must close the cycle (equal to first).
	if path[0] != path[len(path)-1] {
		t.Errorf("expected cycle to close (first=%s, last=%s), got %v", path[0], path[len(path)-1], path)
	}
}

// TestCyclePath_SelfLoop verifies CyclePath with a self-referencing node.
func TestCyclePath_SelfLoop(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	instance.AddNode(nodeA)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeA))

	path := instance.CyclePath()
	if path == nil {
		t.Fatal("expected cycle path for self-loop, got nil")
	}
	if len(path) != 2 || path[0] != "A" || path[1] != "A" {
		t.Errorf("expected [A, A], got %v", path)
	}
}

// TestCyclePath_EmptyGraph verifies CyclePath returns nil for an empty graph.
func TestCyclePath_EmptyGraph(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	if instance.CyclePath() != nil {
		t.Error("expected nil for empty graph")
	}
}

// ---------------------------------------------------------------------------
// FR-11.1: String() — Text Representation
// ---------------------------------------------------------------------------

// TestString_EmptyGraph verifies FR-11.1.3.
func TestString_EmptyGraph(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	result := instance.String()
	if result != "(empty graph)" {
		t.Errorf("expected '(empty graph)', got %q", result)
	}
}

// TestString_SingleNode verifies String() with one isolated node.
func TestString_SingleNode(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	instance.AddNode(nodeA)

	result := instance.String()
	if !containsSubstring(result, "A") {
		t.Errorf("expected 'A' in output, got %q", result)
	}
}

// TestString_WithEdges verifies String() shows tree connectors.
func TestString_WithEdges(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))

	result := instance.String()
	if !containsSubstring(result, "A") {
		t.Errorf("expected 'A' in output, got %q", result)
	}
	if !containsSubstring(result, "B") {
		t.Errorf("expected 'B' in output, got %q", result)
	}
	if !containsSubstring(result, "└──") {
		t.Errorf("expected tree connector in output, got %q", result)
	}
}

// TestString_SharedDependency verifies FR-11.1.2: shared nodes show "(shared)".
func TestString_SharedDependency(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	root := dependency_graph.DependencyGraphNodeNew("root", "")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	shared := dependency_graph.DependencyGraphNodeNew("shared", "")

	instance.AddNode(root)
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	instance.AddNode(shared)

	root.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", root, nodeA))
	root.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", root, nodeB))
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, shared))
	nodeB.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeB, shared))

	result := instance.String()
	if !containsSubstring(result, "(shared)") {
		t.Errorf("expected '(shared)' marker in output, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// FR-11.2: ToDot() — DOT Format Export
// ---------------------------------------------------------------------------

// TestToDot_EmptyGraph verifies FR-11.2.5.
func TestToDot_EmptyGraph(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	result := instance.ToDot()
	expected := "digraph dependencies {\n}"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestToDot_WithNodes verifies FR-11.2.1/FR-11.2.2/FR-11.2.3.
func TestToDot_WithNodes(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))

	result := instance.ToDot()
	if !containsSubstring(result, "digraph dependencies") {
		t.Error("expected 'digraph dependencies' in DOT output")
	}
	if !containsSubstring(result, `"A"`) {
		t.Error("expected quoted node 'A' in DOT output")
	}
	if !containsSubstring(result, `"B"`) {
		t.Error("expected quoted node 'B' in DOT output")
	}
	if !containsSubstring(result, "->") {
		t.Error("expected directed edge in DOT output")
	}
	if !containsSubstring(result, `label="include"`) {
		t.Error("expected edge label 'include' in DOT output")
	}
}

// TestToDot_CyclicHighlighting verifies FR-11.2.4: back-edges are red.
func TestToDot_CyclicHighlighting(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))
	nodeB.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeB, nodeA))

	result := instance.ToDot()
	if !containsSubstring(result, "color=red") {
		t.Error("expected 'color=red' for cycle edge in DOT output")
	}
}

// TestToDot_NoTrailingNewline verifies FR-11.2.6.
func TestToDot_NoTrailingNewline(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	result := instance.ToDot()
	if result[len(result)-1] == '\n' {
		t.Error("DOT output must not end with a trailing newline")
	}
}

// ---------------------------------------------------------------------------
// FR-11.5.2: Idempotent Visualization
// ---------------------------------------------------------------------------

// TestString_Idempotent verifies FR-11.5.2: calling String() multiple times
// produces identical output.
func TestString_Idempotent(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	nodeC := dependency_graph.DependencyGraphNodeNew("C", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	instance.AddNode(nodeC)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeC))

	first := instance.String()
	for i := 0; i < 10; i++ {
		if got := instance.String(); got != first {
			t.Fatalf("String() is not idempotent: call 1 = %q, call %d = %q", first, i+2, got)
		}
	}
}

// TestToDot_Idempotent verifies FR-11.5.2: calling ToDot() multiple times
// produces identical output.
func TestToDot_Idempotent(t *testing.T) {
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}
	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("", "/project")
	nodeA := dependency_graph.DependencyGraphNodeNew("A", "")
	nodeB := dependency_graph.DependencyGraphNodeNew("B", "")
	instance.AddNode(nodeA)
	instance.AddNode(nodeB)
	nodeA.AddEdge(dependency_graph.DependencyGraphEdgeNew("include", nodeA, nodeB))

	first := instance.ToDot()
	for i := 0; i < 10; i++ {
		if got := instance.ToDot(); got != first {
			t.Fatalf("ToDot() is not idempotent: call 1 = %q, call %d = %q", first, i+2, got)
		}
	}
}

// --- helpers ---

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark Acyclic - benchmarks the Acyclic method with a large graph.
func BenchmarkAcyclic(b *testing.B) {
	// ...existing code...
	dependency_graph.OsStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}

	defer func() {
		dependency_graph.OsStat = os.Stat
	}()

	instance := dependency_graph.New("source code", "/path/to/valid/directory")

	// Create a large acyclic graph.
	numNodes := 1000
	for i := 0; i < numNodes; i++ {
		node := dependency_graph.DependencyGraphNodeNew(
			fmt.Sprintf("Node%d", i),
			fmt.Sprintf("source for node %d", i),
		)
		instance.AddNode(node)
		if i > 0 {
			edge := dependency_graph.DependencyGraphEdgeNew(
				fmt.Sprintf("dependency%d", i),
				instance.Nodes()[fmt.Sprintf("Node%d", i-1)],
				node,
			)
			instance.Nodes()[fmt.Sprintf("Node%d", i-1)].AddEdge(edge)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instance.Acyclic()
	}
}
