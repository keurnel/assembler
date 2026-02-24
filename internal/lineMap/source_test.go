package lineMap

import (
	"errors"
	"os"
	"testing"
)

// ============================================================
// Helpers
// ============================================================

// stubFileInfo is a minimal os.FileInfo stub used for testing.
type stubFileInfo struct {
	os.FileInfo
	isDir bool
}

func (s *stubFileInfo) IsDir() bool { return s.isDir }

// withStubs replaces osStat and osReadFile with the provided fakes and
// restores the originals when the test finishes.
func withStubs(t *testing.T, statFn func(string) (os.FileInfo, error), readFn func(string) ([]byte, error)) {
	t.Helper()
	origStat := osStat
	origRead := osReadFile
	osStat = statFn
	osReadFile = readFn
	t.Cleanup(func() {
		osStat = origStat
		osReadFile = origRead
	})
}

// ============================================================
// TestLoadSource
// ============================================================

func TestLoadSource(t *testing.T) {
	// ==============================================================
	//
	// FR-1.2: Rejects files without .kasm extension.
	//
	// ==============================================================
	t.Run("rejects file without .kasm extension", func(t *testing.T) {
		_, err := LoadSource("/tmp/test.asm")
		if err == nil {
			t.Fatal("Expected error for non-.kasm extension, got nil")
		}

		expected := "lineMap error: source file must have a .kasm extension"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})

	// ==============================================================
	//
	// FR-1.2: Rejects files with no extension at all.
	//
	// ==============================================================
	t.Run("rejects file with no extension", func(t *testing.T) {
		_, err := LoadSource("Makefile")
		if err == nil {
			t.Fatal("Expected error for file with no extension, got nil")
		}

		expected := "lineMap error: source file must have a .kasm extension"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})

	// ==============================================================
	//
	// FR-1.2: Rejects .KASM uppercase extension (case-sensitive).
	//
	// ==============================================================
	t.Run("rejects .KASM uppercase extension", func(t *testing.T) {
		_, err := LoadSource("/tmp/test.KASM")
		if err == nil {
			t.Fatal("Expected error for .KASM extension (case-sensitive), got nil")
		}
	})

	// ==============================================================
	//
	// FR-1.2: Rejects path that contains .kasm but doesn't end with it.
	//
	// ==============================================================
	t.Run("rejects path with .kasm in middle", func(t *testing.T) {
		_, err := LoadSource("/tmp/test.kasm.bak")
		if err == nil {
			t.Fatal("Expected error for path ending in .kasm.bak, got nil")
		}
	})

	// ==============================================================
	//
	// FR-1.3: Returns error when file does not exist.
	//
	// ==============================================================
	t.Run("returns error when file does not exist", func(t *testing.T) {
		withStubs(t,
			func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			nil,
		)

		_, err := LoadSource("/tmp/missing.kasm")
		if err == nil {
			t.Fatal("Expected error for missing file, got nil")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected os.ErrNotExist, got '%s'", err.Error())
		}
	})

	// ==============================================================
	//
	// FR-1.3: Returns error for permission denied (non-IsNotExist stat error).
	//
	// ==============================================================
	t.Run("returns error for permission denied", func(t *testing.T) {
		permErr := errors.New("permission denied")
		withStubs(t,
			func(name string) (os.FileInfo, error) { return nil, permErr },
			nil,
		)

		_, err := LoadSource("/tmp/secret.kasm")
		if err == nil {
			t.Fatal("Expected error for permission denied, got nil")
		}
		if err != permErr {
			t.Errorf("Expected permission denied error, got '%s'", err.Error())
		}
	})

	// ==============================================================
	//
	// FR-1.4: Returns error when path is a directory.
	//
	// ==============================================================
	t.Run("returns error when path is a directory", func(t *testing.T) {
		withStubs(t,
			func(name string) (os.FileInfo, error) { return &stubFileInfo{isDir: true}, nil },
			nil,
		)

		_, err := LoadSource("/tmp/somedir.kasm")
		if err == nil {
			t.Fatal("Expected error when path is a directory, got nil")
		}

		expected := "lineMap error: source path is a directory where a file is expected"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})

	// ==============================================================
	//
	// FR-1.5: Returns error when os.ReadFile fails.
	//
	// ==============================================================
	t.Run("returns error when ReadFile fails", func(t *testing.T) {
		readErr := errors.New("disk I/O error")
		withStubs(t,
			func(name string) (os.FileInfo, error) { return &stubFileInfo{isDir: false}, nil },
			func(name string) ([]byte, error) { return nil, readErr },
		)

		_, err := LoadSource("/tmp/broken.kasm")
		if err == nil {
			t.Fatal("Expected error when ReadFile fails, got nil")
		}
		if err != readErr {
			t.Errorf("Expected disk I/O error, got '%s'", err.Error())
		}
	})

	// ==============================================================
	//
	// FR-1.6: Successfully loads file content into Source.
	//
	// ==============================================================
	t.Run("loads file content successfully", func(t *testing.T) {
		fileContent := "section .text\n    mov rax, 1\n    syscall"
		withStubs(t,
			func(name string) (os.FileInfo, error) { return &stubFileInfo{isDir: false}, nil },
			func(name string) ([]byte, error) { return []byte(fileContent), nil },
		)

		src, err := LoadSource("/tmp/main.kasm")
		if err != nil {
			t.Fatalf("Expected no error, got '%s'", err.Error())
		}

		if src.content != fileContent {
			t.Errorf("Expected content '%s', got '%s'", fileContent, src.content)
		}
	})

	// ==============================================================
	//
	// FR-1.6: Loads empty file content.
	//
	// ==============================================================
	t.Run("loads empty file", func(t *testing.T) {
		withStubs(t,
			func(name string) (os.FileInfo, error) { return &stubFileInfo{isDir: false}, nil },
			func(name string) ([]byte, error) { return []byte(""), nil },
		)

		src, err := LoadSource("/tmp/empty.kasm")
		if err != nil {
			t.Fatalf("Expected no error, got '%s'", err.Error())
		}
		if src.content != "" {
			t.Errorf("Expected empty content, got '%s'", src.content)
		}
	})

	// ==============================================================
	//
	// FR-1.1 / FR-1.6: Passes the correct path to osStat and osReadFile,
	// and stores it on the returned Source.
	//
	// ==============================================================
	t.Run("passes correct path to osStat and osReadFile", func(t *testing.T) {
		expectedPath := "/absolute/path/to/file.kasm"
		var statPath, readPath string

		withStubs(t,
			func(name string) (os.FileInfo, error) {
				statPath = name
				return &stubFileInfo{isDir: false}, nil
			},
			func(name string) ([]byte, error) {
				readPath = name
				return []byte("content"), nil
			},
		)

		src, err := LoadSource(expectedPath)
		if err != nil {
			t.Fatalf("LoadSource failed: %s", err.Error())
		}

		if statPath != expectedPath {
			t.Errorf("Expected osStat path '%s', got '%s'", expectedPath, statPath)
		}
		if readPath != expectedPath {
			t.Errorf("Expected osReadFile path '%s', got '%s'", expectedPath, readPath)
		}
		if src.path != expectedPath {
			t.Errorf("Expected Source.path '%s', got '%s'", expectedPath, src.path)
		}
	})

	// ==============================================================
	//
	// FR-1.8: Returned Source on error has zero value (empty path and content).
	//
	// ==============================================================
	t.Run("returns zero-value Source on error", func(t *testing.T) {
		src, err := LoadSource("/tmp/test.txt")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if src.path != "" {
			t.Errorf("Expected empty path on error, got '%s'", src.path)
		}
		if src.content != "" {
			t.Errorf("Expected empty content on error, got '%s'", src.content)
		}
	})
}

// ============================================================
// TestSource_Path
// ============================================================

func TestSource_Path(t *testing.T) {
	// ==============================================================
	//
	// FR-1.7: Returns the file path that the Source was created with.
	//
	// ==============================================================
	t.Run("returns the path", func(t *testing.T) {
		withStubs(t,
			func(name string) (os.FileInfo, error) { return &stubFileInfo{isDir: false}, nil },
			func(name string) ([]byte, error) { return []byte(""), nil },
		)

		src, err := LoadSource("/home/user/main.kasm")
		if err != nil {
			t.Fatalf("LoadSource failed: %s", err.Error())
		}

		if src.Path() != "/home/user/main.kasm" {
			t.Errorf("Expected Path() to return '/home/user/main.kasm', got '%s'", src.Path())
		}
	})
}

// ============================================================
// TestSource_Content
// ============================================================

func TestSource_Content(t *testing.T) {
	// ==============================================================
	//
	// FR-1.7: Returns the loaded file content.
	//
	// ==============================================================
	t.Run("returns the loaded content", func(t *testing.T) {
		withStubs(t,
			func(name string) (os.FileInfo, error) { return &stubFileInfo{isDir: false}, nil },
			func(name string) ([]byte, error) { return []byte("mov rax, 1"), nil },
		)

		src, err := LoadSource("/tmp/test.kasm")
		if err != nil {
			t.Fatalf("LoadSource failed: %s", err.Error())
		}

		if src.Content() != "mov rax, 1" {
			t.Errorf("Expected Content() to be 'mov rax, 1', got '%s'", src.Content())
		}
	})
}
