package lineMap

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const (
	LineSnapshotTypeInitial  = "initial"
	LineSnapshotTypeChange   = "change"
	LineSnapshotTypeNoChange = "no-change"
)

type LineChange struct {
	_type    string // "expanding", "contracting", or "unchanged" — set by factory, never by callers.
	origin   int    // 0-based line index in the previous (old) version.
	newIndex int    // 0-based line index in the new version. -1 for contracting (line no longer exists).
	content  string // The actual text of the line.
}

// --- Accessor methods (FR-5.5) ---

// Type returns the change type: "unchanged", "expanding", or "contracting".
func (lc LineChange) Type() string { return lc._type }

// Origin returns the 0-based line index in the previous (old) version.
func (lc LineChange) Origin() int { return lc.origin }

// NewIndex returns the 0-based line index in the new version.
// Returns -1 for contracting changes (the line no longer exists).
func (lc LineChange) NewIndex() int { return lc.newIndex }

// Content returns the actual text of the line.
func (lc LineChange) Content() string { return lc.content }

// --- Factory functions (FR-5.2) ---

// newUnchangedChange creates a LineChange that records an unchanged line.
// Captures where it was (origin), where it is now (newIndex), and its content.
func newUnchangedChange(origin, newIndex int, content string) LineChange {
	return LineChange{
		_type:    "unchanged",
		origin:   origin,
		newIndex: newIndex,
		content:  content,
	}
}

// newExpandingChange creates a LineChange that records an inserted/added line.
// Captures the nearest origin line, the position in the new version, and the
// inserted line content.
func newExpandingChange(origin, newIndex int, content string) LineChange {
	return LineChange{
		_type:    "expanding",
		origin:   origin,
		newIndex: newIndex,
		content:  content,
	}
}

// newContractingChange creates a LineChange that records a removed line.
// Captures the position in the old version and the removed line content.
// newIndex is -1 because the line no longer exists in the new version.
func newContractingChange(origin int, content string) LineChange {
	return LineChange{
		_type:    "contracting",
		origin:   origin,
		newIndex: -1,
		content:  content,
	}
}

// String returns a human-readable representation of the LineChange for debugging.
func (lc LineChange) String() string {
	switch lc._type {
	case "expanding":
		return fmt.Sprintf("LineChange{Type: %s, Origin: %d, NewIndex: %d, Content: %q}",
			lc._type, lc.origin, lc.newIndex, lc.content)
	case "contracting":
		return fmt.Sprintf("LineChange{Type: %s, Origin: %d, Content: %q}",
			lc._type, lc.origin, lc.content)
	default:
		return fmt.Sprintf("LineChange{Type: %s, Origin: %d, NewIndex: %d, Content: %q}",
			lc._type, lc.origin, lc.newIndex, lc.content)
	}
}

type LinesSnapshot struct {
	_type    string
	hash     string
	source   string
	lines    []string
	changes  *map[int]LineChange // Keyed by new-version line index (unchanged + expanding entries).
	removals []LineChange        // Contracting entries (removed lines, no position in the new version).
}

type History struct {
	items []LinesSnapshot
}

// empty - returns true if the history is empty, false otherwise.
func (h *History) empty() bool {
	return len(h.items) == 0
}

// notEmpty - inverse function of `empty()`, returns true if the history is not empty,
// false otherwise.
func (h *History) notEmpty() bool {
	return !h.empty()
}

// latest - returns the latest snapshot in the history. Returns nil if the history is empty.
func (h *History) latest() *LinesSnapshot {
	if h.empty() {
		return nil
	}
	return &h.items[len(h.items)-1]
}

// LineOrigin - traces a line number in the current (latest) snapshot back through
// all change snapshots to find the original line number in the initial snapshot.
// Returns -1 if the line was inserted during pre-processing (expanding).
//
// There is no empty-history guard — LineOrigin is only reachable on a fully
// constructed Instance (FR-2), which guarantees at least one snapshot.
// There is no "contracting" case — contracting entries live in the removals
// slice (FR-5.6), not in the changes map.
func (h *History) LineOrigin(lineNumber int) int {
	current := lineNumber

	// Walk backwards through snapshots (skip the initial one at index 0).
	for i := len(h.items) - 1; i > 0; i-- {
		snapshot := h.items[i]
		if snapshot.changes == nil {
			continue
		}

		change, exists := (*snapshot.changes)[current]
		if !exists {
			// Line was not part of any change, it maps 1:1.
			continue
		}

		if change.Type() == "expanding" {
			// This line was inserted by the preprocessor; it has no origin.
			return -1
		}

		// unchanged — trace through to the original position.
		current = change.Origin()
	}

	return current
}

// snapshotInitial - creates the initial snapshot of the Instance and appends
// it to the history. This method is only called from New() and always creates
// a snapshot of type LineSnapshotTypeInitial. Because it is the only method
// that produces an initial snapshot, a second one cannot be created.
func (h *History) snapshotInitial(instance *Instance) {
	h.items = append(h.items, LinesSnapshot{
		_type:   LineSnapshotTypeInitial,
		hash:    h.snapshotHashGenerate(instance.value),
		source:  instance.value,
		lines:   strings.Split(instance.value, "\n"),
		changes: nil,
	})
}

// snapshotNoChange - records that the source did not change between this
// update and the previous snapshot. Called only from Update().
// No _type parameter — the type is determined by which method is called.
func (h *History) snapshotNoChange(instance *Instance) {
	h.items = append(h.items, LinesSnapshot{
		_type:   LineSnapshotTypeNoChange,
		hash:    h.snapshotHashGenerate(instance.value),
		source:  instance.value,
		lines:   strings.Split(instance.value, "\n"),
		changes: nil,
	})
}

// snapshotChange - records a diff between this update and the previous
// snapshot. Called only from Update(). The changes map is keyed by new-version
// line index (unchanged + expanding). The removals slice contains contracting
// entries. No _type parameter — the type is determined by which method is called.
func (h *History) snapshotChange(instance *Instance, changes map[int]LineChange, removals []LineChange) {
	h.items = append(h.items, LinesSnapshot{
		_type:    LineSnapshotTypeChange,
		hash:     h.snapshotHashGenerate(instance.value),
		source:   instance.value,
		lines:    strings.Split(instance.value, "\n"),
		changes:  &changes,
		removals: removals,
	})
}

// snapshotHashGenerate - generates a hash for the source of a snapshot. This is used
// to quickly compare snapshots and determine if they are identical or not.
func (h *History) snapshotHashGenerate(source string) string {
	return generateSourceHash(source)
}

// snapshotHashCompare - compares the hash of a snapshot with the hash of another snapshot. Returns
// true if the hashes are equal, false otherwise.
func (h *History) snapshotHashCompare(snapshot1, snapshot2 LinesSnapshot) bool {
	return snapshot1.hash == snapshot2.hash
}

// SourceCompare - compares the source of a snapshot with a given value. Returns true if the
// sources are equal, false otherwise.
func (s *LinesSnapshot) SourceCompare(value string) bool {
	return s.hash == generateSourceHash(value)
}

// generateSourceHash - generates a hash for a given source string. This is used to quickly
// compare sources and determine if they are identical or not.
func generateSourceHash(source string) string {
	hash := sha256.Sum256([]byte(source))
	return fmt.Sprintf("%x", hash)
}
