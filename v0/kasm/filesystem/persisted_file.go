package filesystem

import "os"

var (
	OsReadFile = os.ReadFile
)

type PersistedFile struct {
	// Path - the file path on persistent storage (e.g., disk).
	Path string
	// Content - the file content as it is on persistent storage.
	Content *string
}

// PersistedFileNew - creates a new PersistedFile with the given path and content.
func PersistedFileNew(path string) PersistedFile {
	return PersistedFile{
		Path:    path,
		Content: nil,
	}
}

// Load - loads the content of the file from persistent storage (e.g., disk) into the content field of the struct.
func (p *PersistedFile) Load() error {
	content, err := OsReadFile(p.Path)
	if err != nil {
		return err
	}

	contentStr := string(content)
	p.Content = &contentStr

	return nil
}

// Loaded - returns true if the content of the file has been loaded into memory, false otherwise.
func (p *PersistedFile) Loaded() bool {
	return p.Content != nil
}
