package debugcontext

import "sync"

// DebugContext is a passive, append-only data structure that accumulates
// diagnostic entries as the assembler pipeline progresses. It is thread-safe
// for concurrent writes.
//
// Create a DebugContext exclusively through NewDebugContext(). It is passed
// through the pipeline by reference — every stage records entries into the
// same context.
//
// The context does not perform I/O or formatting. A separate renderer
// consumes the entries to produce output.
//
// See FR-1 through FR-9 in .requirements/assembler/debug-information.md.
type DebugContext struct {
	filePath string   // Primary source file path (FR-1.2).
	phase    string   // Current pipeline phase (FR-2).
	entries  []*Entry // Recorded entries in insertion order (FR-8.2).
	mu       sync.Mutex
}

// NewDebugContext is the sole constructor. It returns a *DebugContext
// initialised with the primary source file path, an empty entry list,
// and the phase set to "" (no phase).
func NewDebugContext(filePath string) *DebugContext {
	return &DebugContext{
		filePath: filePath,
		phase:    "",
		entries:  make([]*Entry, 0),
	}
}

// --- Phases (FR-2) ---

// SetPhase sets the current pipeline phase. Subsequent entries are tagged
// with this phase until it is changed again.
func (c *DebugContext) SetPhase(name string) {
	c.mu.Lock()
	c.phase = name
	c.mu.Unlock()
}

// Phase returns the current pipeline phase name.
func (c *DebugContext) Phase() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.phase
}

// --- Location helpers (FR-4) ---

// Loc creates a Location using the primary file path from the context.
func (c *DebugContext) Loc(line, column int) Location {
	return Loc(c.filePath, line, column)
}

// LocIn creates a Location with an explicit file path (used for lines
// originating from included files).
func (c *DebugContext) LocIn(filePath string, line, column int) Location {
	return Loc(filePath, line, column)
}

// --- Recording methods (FR-3.3) ---

// record is the internal method that creates an entry and appends it to
// the context. It is thread-safe.
func (c *DebugContext) record(severity string, location Location, message string) *Entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &Entry{
		severity: severity,
		phase:    c.phase,
		message:  message,
		location: location,
	}
	c.entries = append(c.entries, entry)
	return entry
}

// Error records an entry with severity "error" and returns the *Entry
// for optional chaining (WithSnippet, WithHint).
func (c *DebugContext) Error(location Location, message string) *Entry {
	return c.record(SeverityError, location, message)
}

// Warning records an entry with severity "warning" and returns the *Entry
// for optional chaining.
func (c *DebugContext) Warning(location Location, message string) *Entry {
	return c.record(SeverityWarning, location, message)
}

// Info records an entry with severity "info" and returns the *Entry
// for optional chaining.
func (c *DebugContext) Info(location Location, message string) *Entry {
	return c.record(SeverityInfo, location, message)
}

// Trace records an entry with severity "trace" and returns the *Entry
// for optional chaining.
func (c *DebugContext) Trace(location Location, message string) *Entry {
	return c.record(SeverityTrace, location, message)
}

// --- Querying entries (FR-5) ---

// Entries returns all recorded entries in insertion order.
func (c *DebugContext) Entries() []*Entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]*Entry, len(c.entries))
	copy(result, c.entries)
	return result
}

// Errors returns only entries with severity "error".
func (c *DebugContext) Errors() []*Entry {
	return c.filter(SeverityError)
}

// Warnings returns only entries with severity "warning".
func (c *DebugContext) Warnings() []*Entry {
	return c.filter(SeverityWarning)
}

// HasErrors returns true if at least one "error" entry exists. This is the
// primary check used to decide whether the pipeline should abort.
func (c *DebugContext) HasErrors() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, e := range c.entries {
		if e.severity == SeverityError {
			return true
		}
	}
	return false
}

// Count returns the total number of entries.
func (c *DebugContext) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// FilePath returns the primary source file path.
func (c *DebugContext) FilePath() string {
	return c.filePath
}

// filter returns all entries matching the given severity.
func (c *DebugContext) filter(severity string) []*Entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result []*Entry
	for _, e := range c.entries {
		if e.severity == severity {
			result = append(result, e)
		}
	}
	return result
}
