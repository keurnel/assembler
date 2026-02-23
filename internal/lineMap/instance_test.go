package lineMap

import (
	"os"
	"testing"
)

type fakeFileInfo struct {
	os.FileInfo
	isDir bool
}

func (f *fakeFileInfo) IsDir() bool {
	return f.isDir
}

func TestNew(t *testing.T) {
	// ==============================================================
	//
	// Returns error when `source.Load()` returns an error.
	//
	// ==============================================================
	t.Run("source.Load() returns an error", func(t *testing.T) {
		osStat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}

		_, err := New("value", Source{path: "fakePath"})
		if err == nil {
			t.Error("Expected error when `source.Load()` returns an error, got nil")
		}

		if err.Error() != "file does not exist" {
			t.Errorf("Expected error message 'file does not exist', got '%s'", err.Error())
		}

		osStat = os.Stat // Reset osStat to its original implementation for other tests.
	})

	// ==============================================================
	//
	// Returns error when `source.Load()` returns an error because the path is a directory.
	//
	// ==============================================================
	t.Run("source.Load() returns an error because the path is a directory", func(t *testing.T) {
		osStat = func(name string) (os.FileInfo, error) {
			return &fakeFileInfo{isDir: true}, nil
		}

		_, err := New("value", Source{path: "fakePath"})
		if err == nil {
			t.Error("Expected error when `source.Load()` returns an error because the path is a directory, got nil")
		}

		if err.Error() != "lineMap error: source path is a directory were a file is expected" {
			t.Errorf("Expected error message 'lineMap error: source path is a directory were a file is expected', got '%s'", err.Error())
		}

		osStat = os.Stat // Reset osStat to its original implementation for other tests.
	})

	// ==============================================================
	//
	// Returns a new instance of `Instance` when `source.Load()` does not return an error.
	//
	// ==============================================================
	t.Run("source.Load() does not return an error", func(t *testing.T) {
		osStat = func(name string) (os.FileInfo, error) {
			if name != "fakePath" {
				t.Errorf("Expected `osStat` to be called with 'fakePath', got '%s'", name)
				return nil, os.ErrNotExist
			}
			return &fakeFileInfo{isDir: false}, nil
		}

		osReadFile = func(name string) ([]byte, error) {
			return []byte("fake file content"), nil
		}

		instance, err := New("value", Source{path: "fakePath"})
		if err != nil {
			t.Errorf("Expected no error when `source.Load()` does not return an error, got '%s'", err.Error())
			return
		}

		if instance == nil {
			t.Error("Expected a new instance of `Instance`, got nil")
			return
		}

		if instance.value != "value" {
			t.Errorf("Expected instance value to be 'value', got '%s'", instance.value)
		}

		osStat = os.Stat
		osReadFile = os.ReadFile
	})
}

func TestInitialIndex(t *testing.T) {
	// ==============================================================
	//
	// Returns false when `Instance.history` is not empty.
	//
	// ==============================================================
	t.Run("Instance.history is not empty", func(t *testing.T) {
		instance := &Instance{
			history: History{items: []LinesSnapshot{
				{lines: []string{"line1", "line2"}},
			}},
		}

		result := instance.InitialIndex()
		if result != nil {
			t.Errorf("Expected InitialIndex to return an error when Instance.history is not empty, got nil")
		}
	})
	// ==============================================================
	//
	// Indexes the lines in `Instance.value` when `Instance.history` is empty and returns true.
	//
	// ==============================================================
	t.Run("Instance.history is empty", func(t *testing.T) {
		instance := &Instance{
			value:   "line1\nline2\nline3",
			history: History{items: []LinesSnapshot{}},
		}

		result := instance.InitialIndex()
		if result != nil {
			t.Errorf("Expected InitialIndex to return nil when Instance.history is empty, got '%s'", result.Error())
		}

		if len(instance.history.items) != 1 {
			t.Errorf("Expected Instance.history to have 1 item after InitialIndex, got %d", len(instance.history.items))
		}

		if instance.history.items[0]._type != LineSnapshotTypeInitial {
			t.Errorf("Expected the snapshot type to be '%s', got '%s'", LineSnapshotTypeInitial, instance.history.items[0]._type)
		}
	})
	// ==============================================================
	//
	// Initial index returns false when called multiple times, only the first call should perform the indexing and return true.
	//
	// ==============================================================
	t.Run("InitialIndex returns false when called multiple times", func(t *testing.T) {
		instance := &Instance{
			value:   "line1\nline2\nline3",
			history: History{items: []LinesSnapshot{}},
		}

		firstCallResult := instance.InitialIndex()
		if firstCallResult != nil {
			t.Errorf("Expected first call to InitialIndex to return nil, got '%s'", firstCallResult.Error())
		}

		secondCallResult := instance.InitialIndex()
		if secondCallResult == nil {
			t.Error("Expected second call to InitialIndex to return an error, got nil")
		}
	})
	// ==============================================================
	//
	// InitialIndex should create a correct snapshot of the initial state of the `Instance` and store it in the history.
	//
	// ==============================================================
	t.Run("InitialIndex creates a correct snapshot of the initial state of the Instance", func(t *testing.T) {
		instance := &Instance{
			value:   "line1\nline2\nline3",
			history: History{items: []LinesSnapshot{}},
		}

		succeed := instance.InitialIndex()
		if succeed != nil {
			t.Errorf("Expected InitialIndex to succeed when Instance.history is empty, got false")
		}

		if len(instance.history.items) != 1 {
			t.Errorf("Expected Instance.history to have 1 item after InitialIndex, got %d", len(instance.history.items))
		}

		snapshot := instance.history.items[0]
		if snapshot._type != LineSnapshotTypeInitial {
			t.Errorf("Expected snapshot type to be '%s', got '%s'", LineSnapshotTypeInitial, snapshot._type)
		}

		if snapshot.source != instance.value {
			t.Errorf("Expected snapshot source to be '%s', got '%s'", instance.value, snapshot.source)
		}

		if len(snapshot.lines) != 3 {
			t.Errorf("Expected snapshot lines to have 3 lines, got %d", len(snapshot.lines))
		}

		// Check if each line on the snapshot matches the expected lines.
		//
		expectedLines := []string{"line1", "line2", "line3"}
		for i, line := range snapshot.lines {
			if line != expectedLines[i] {
				t.Errorf("Expected line %d to be '%s', got '%s'", i+1, expectedLines[i], line)
			}
		}
	})
}

func TestInstance_Update(t *testing.T) {
	// ==============================================================
	//
	// Returns error when no initial snapshot exists in the history.
	//
	// ==============================================================
	t.Run("Returns error when no initial snapshot exists in the history", func(t *testing.T) {
		instance := &Instance{
			value:   "line1\nline2\nline3",
			history: History{items: []LinesSnapshot{}},
		}

		err := instance.Update("new value")
		if err == nil {
			t.Error("Expected Update to return an error when no initial snapshot exists in the history, got nil")
		}
	})
	// ==============================================================
	//
	// Expanding a line
	//
	// ==============================================================
	t.Run("Expanding a line", func(t *testing.T) {
		instance := &Instance{
			value:   "line1\nline2\nline3",
			history: History{items: []LinesSnapshot{}},
		}

		instance.InitialIndex()

		err := instance.Update("line1\nline2\nline3\nline4")
		if err != nil {
			t.Errorf("Expected Update to succeed when updating with a new value, got '%s'", err.Error())
		}

		// Print result in history
		if len(instance.history.items) != 2 {
			t.Errorf("Expected Instance.history to have 2 items after Update, got %d", len(instance.history.items))
		}

		snapshot := instance.history.items[1]

		if snapshot._type != LineSnapshotTypeChange {
			t.Errorf("Expected snapshot type to be '%s', got '%s'", LineSnapshotTypeChange, snapshot._type)
		}

		if snapshot.source != "line1\nline2\nline3\nline4" {
			t.Errorf("Expected snapshot source to be 'line1\\nline2\\nline3\\nline4', got '%s'", snapshot.source)
		}

		if len(snapshot.lines) != 4 {
			t.Errorf("Expected snapshot lines to have 4 lines, got %d", len(snapshot.lines))
		}
	})
}
