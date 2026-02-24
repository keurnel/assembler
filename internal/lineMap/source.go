package lineMap

import (
	"errors"
	"os"
	"strings"
)

var (
	osStat     = os.Stat
	osReadFile = os.ReadFile
)

// Source represents a validated, loaded source file. If a Source value
// exists, it is guaranteed to hold a valid path and its file content.
// There is no unloaded or partially-initialised state.
//
// Create a Source exclusively through LoadSource().
type Source struct {
	// path - absolute path to the source file.
	path string
	// content - the content of the source file.
	content string
}

// LoadSource validates the path, reads the file, and returns a ready-to-use
// Source — or an error. This is the only way to construct a Source.
func LoadSource(path string) (Source, error) {
	// Validate file extension.
	//
	if !strings.HasSuffix(path, ".kasm") {
		return Source{}, errors.New("lineMap error: source file must have a .kasm extension")
	}

	// Check if file exists and is accessible.
	//
	file, err := osStat(path)
	if err != nil {
		return Source{}, err
	}

	// Ensure path is not a directory.
	//
	if file.IsDir() {
		return Source{}, errors.New("lineMap error: source path is a directory where a file is expected")
	}

	// Read the file content.
	//
	contentBytes, err := osReadFile(path)
	if err != nil {
		return Source{}, err
	}

	return Source{
		path:    path,
		content: string(contentBytes),
	}, nil
}

// Path returns the file path of the source.
func (s Source) Path() string {
	return s.path
}

// Content returns the loaded content of the source file.
func (s Source) Content() string {
	return s.content
}
