package filesystem_test

import (
	"os"
	"testing"

	"github.com/keurnel/assembler/v0/kasm/filesystem"
)

func TestFileInMemoryNew(t *testing.T) {
	scenarios := []struct {
		name                 string
		osReadFile           func(string) ([]byte, error)
		expectedFileInMemory *filesystem.FileInMemory
		expectedError        error
		expectedValue        string
	}{
		{
			name: "Not existing file",
			osReadFile: func(string) ([]byte, error) {
				return nil, &os.PathError{Op: "open", Path: "nonexistent.asm", Err: os.ErrNotExist}
			},
			expectedFileInMemory: nil,
			expectedError:        &os.PathError{Op: "open", Path: "nonexistent.asm", Err: os.ErrNotExist},
			expectedValue:        "",
		},
		{
			name: "Existing empty file",
			osReadFile: func(string) ([]byte, error) {
				return []byte(""), nil
			},
			expectedFileInMemory: &filesystem.FileInMemory{
				Source:  filesystem.PersistedFile{Path: "dummy.asm", Content: nil},
				Value:   "",
				LineMap: []filesystem.LineMap{},
			},
			expectedError: nil,
			expectedValue: "",
		},
		{
			name: "Existing file with content",
			osReadFile: func(string) ([]byte, error) {
				return []byte("section .text\n"), nil
			},
			expectedFileInMemory: &filesystem.FileInMemory{
				Source:  filesystem.PersistedFile{Path: "dummy.asm", Content: nil},
				Value:   "",
				LineMap: []filesystem.LineMap{},
			},
			expectedError: nil,
			expectedValue: "section .text\n",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Override the OsReadFile function with the scenario's mock implementation
			filesystem.OsReadFile = scenario.osReadFile

			// Create a PersistedFile with a dummy path
			persistedFile := filesystem.PersistedFileNew("dummy.asm")

			// Call FileInMemoryNew with the persisted file
			fileInMemory, err := filesystem.FileInMemoryNew(persistedFile)

			// Check if the returned error matches the expected error
			if (err != nil && scenario.expectedError == nil) || (err == nil && scenario.expectedError != nil) {
				t.Fatalf("Expected error: %v, got: %v", scenario.expectedError, err)
			}
			if err != nil && scenario.expectedError != nil && err.Error() != scenario.expectedError.Error() {
				t.Fatalf("Expected error message: %v, got: %v", scenario.expectedError.Error(), err.Error())
			}

			// Check if the returned FileInMemory matches the expected FileInMemory
			if scenario.expectedFileInMemory != nil {
				if fileInMemory == nil {
					t.Fatalf("Expected FileInMemory: %v, got: nil", scenario.expectedFileInMemory)
				}
				if fileInMemory.Source.Path != scenario.expectedFileInMemory.Source.Path {
					t.Errorf("Expected Source.Path: %v, got: %v", scenario.expectedFileInMemory.Source.Path, fileInMemory.Source.Path)
				}
				if fileInMemory.Value != scenario.expectedValue {
					t.Errorf("Expected Value: %v, got: %v", scenario.expectedValue, fileInMemory.Value)
				}
				if len(fileInMemory.LineMap) != len(scenario.expectedFileInMemory.LineMap) {
					t.Errorf("Expected LineMap length: %d, got: %d", len(scenario.expectedFileInMemory.LineMap), len(fileInMemory.LineMap))
				}
			} else if fileInMemory != nil {
				t.Fatalf("Expected FileInMemory: nil, got: %v", fileInMemory)
			}
		})
	}
}

func TestFileInMemory_RecordInclusion(t *testing.T) {
	filesystem.OsReadFile = func(string) ([]byte, error) {
		return []byte("section .text\n"), nil
	}

	str := "section .text\n"
	strPtr := &str

	mainFile := filesystem.FileInMemory{
		Source:  filesystem.PersistedFile{Path: "main.asm", Content: strPtr},
		Value:   "section .data\n",
		LineMap: []filesystem.LineMap{},
	}

	includedFile := filesystem.PersistedFile{
		Path: "included.asm",
		Content: func() *string {
			content := "section .text\n"
			return &content
		}(),
	}

	mainFile.RecordInclusion(&includedFile, 2)

	// Get origin of line 2
	origin := mainFile.Origin(2)
	if len(origin) != 1 {
		t.Fatalf("Expected origin length: 1, got: %d", len(origin))
	}

	for i, lineMap := range origin {
		println("index: ", i, "filename:", lineMap.Filename())
	}

}
