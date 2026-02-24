package lineMap

import (
	"errors"
	"os"
)

var (
	osStat     = os.Stat
	osReadFile = os.ReadFile
)

type Source struct {
	// path - absolute path to the source file.
	path string
	// content - the content of the source file.
	content string
}

func (s *Source) Load() error {
	// Check if file exists before loading.
	//
	file, err := osStat(s.path)
	if os.IsNotExist(err) {
		return err
	}

	// Ensure `file` is not a directory.
	//
	if file.IsDir() {
		return errors.New("lineMap error: source path is a directory were a file is expected")
	}

	// Load the content of the source file.
	contentBytes, err := osReadFile(s.path)
	if err != nil {
		return err
	}
	s.content = string(contentBytes)

	return nil
}
