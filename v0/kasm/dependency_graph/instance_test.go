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
}

// Benchmark Acyclic - benchmarks the Acyclic method with a large graph.
func BenchmarkAcyclic(b *testing.B) {
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
