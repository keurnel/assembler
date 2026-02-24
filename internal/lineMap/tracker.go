package lineMap

// Tracker provides a simplified, high-level API for the most common lineMap
// workflow: load a source file, track it through pre-processing transformations,
// and trace lines back to their origin.
//
// Create a Tracker exclusively through Track(). If a Tracker exists, it is
// guaranteed to hold a valid, fully initialised Instance.
type Tracker struct {
	instance *Instance
	source   Source
}

// Track is the single entry point for the facade. It validates and reads the
// file at path, creates an Instance with the initial snapshot, and returns a
// ready-to-use *Tracker — or an error if the file cannot be loaded.
//
// The caller does not need to read the file separately — the content read by
// LoadSource is used for the initial snapshot.
func Track(path string) (*Tracker, error) {
	src, err := LoadSource(path)
	if err != nil {
		return nil, err
	}

	instance := New(src.Content(), src)

	return &Tracker{
		instance: instance,
		source:   src,
	}, nil
}

// --- Snapshotting (FR-11.2) ---

// Snapshot records a new version of the source after a pre-processing step.
// It is infallible — it delegates to Instance.Update which is infallible (FR-4.1).
func (t *Tracker) Snapshot(source string) {
	t.instance.Update(source)
}

// --- Tracing (FR-11.3) ---

// Origin traces a line in the latest processed source back to its original
// line number in the initial source. Returns -1 if the line was inserted
// during pre-processing.
func (t *Tracker) Origin(lineNumber int) int {
	return t.instance.LineOrigin(lineNumber)
}

// History returns the chronological evolution of a line across all snapshots.
// Each entry carries full detail (type, origin, newIndex, content).
func (t *Tracker) History(lineNumber int) []LineChange {
	return t.instance.LineHistory(lineNumber)
}

// --- Read Access (FR-11.4) ---

// Source returns the current processed source string.
func (t *Tracker) Source() string {
	return t.instance.Value()
}

// Lines returns the lines of the current processed source.
func (t *Tracker) Lines() []string {
	return t.instance.Lines()
}

// FilePath returns the original file path that was passed to Track().
func (t *Tracker) FilePath() string {
	return t.source.Path()
}
