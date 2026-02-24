package lineMap

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

const (
	LineSnapshotTypeInitial  = "initial"
	LineSnapshotTypeChange   = "change"
	LineSnapshotTypeNoChange = "no-change"

	LineSnapshotTypeExpanding   = "expanding"
	LineSnapshotTypeContracting = "contracting"
)

type LineChange struct {
	_type  string // Expanding, Contracting, Unchanged
	origin int    // Line number in the original source (before the change)

	// Expanding information.
	//
	expandingRangeStart int      // Starting line number in the new source (after the change)
	expandingRangeEnd   int      // Ending line number in the new source (after the change)
	expandingLines      []string // Lines that were added to the new source.

	// Contracting information.
	//
	contractingRangeStart int      // Starting line number in the original source (before the change)
	contractingRangeEnd   int      // Ending line number in the original source (before the change)
	contractingLines      []string // Lines that were removed from the original source.
}

func newLineChange(_type string, origin, rangeStart, rangeEnd int) (*LineChange, error) {
	if rangeStart > rangeEnd || rangeEnd < rangeStart || rangeStart < 0 || rangeEnd < 0 {
		return nil, errors.New("invalid line change range")
	}

	if _type != "expanding" && _type != "contracting" && _type != "unchanged" {
		return nil, errors.New("invalid line change type")
	}

	if _type == "unchanged" && (rangeStart != rangeEnd) {
		return nil, errors.New("unchanged line change must have rangeStart equal to rangeEnd")
	}

	if _type == "expanding" {
		return &LineChange{
			_type:               _type,
			origin:              origin,
			expandingRangeStart: rangeStart,
			expandingRangeEnd:   rangeEnd,
		}, nil
	}

	if _type == "contracting" {
		return &LineChange{
			_type:                 _type,
			origin:                origin,
			contractingRangeStart: rangeStart,
			contractingRangeEnd:   rangeEnd,
		}, nil
	}

	return &LineChange{
		_type:  _type,
		origin: origin,
	}, nil
}

// String - returns a string representation of the LineChange for debugging purposes.
func (lc *LineChange) String() string {
	switch lc._type {
	case LineSnapshotTypeExpanding:
		return fmt.Sprintf("LineChange{Type: %s, Origin: %d, ExpandingRange: [%d-%d], ExpandingLines: %v}",
			lc._type, lc.origin, lc.expandingRangeStart, lc.expandingRangeEnd, lc.expandingLines)
	case LineSnapshotTypeContracting:
		return fmt.Sprintf("LineChange{Type: %s, Origin: %d, ContractingRange: [%d-%d], ContractingLines: %v}",
			lc._type, lc.origin, lc.contractingRangeStart, lc.contractingRangeEnd, lc.contractingLines)
	default:
		return fmt.Sprintf("LineChange{Type: %s, Origin: %d}", lc._type, lc.origin)
	}
}

type LinesSnapshot struct {
	_type   string
	hash    string
	source  string
	lines   []string
	changes *map[int]LineChange
}

type History struct {
	hasInitialSnapshot bool
	items              []LinesSnapshot
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
// Returns -1 if the line cannot be traced (e.g. it was inserted by a preprocessor step).
func (h *History) LineOrigin(lineNumber int) int {
	if h.empty() {
		return -1
	}

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

		switch change._type {
		case LineSnapshotTypeExpanding:
			// This line was inserted by the preprocessor; it has no origin.
			return -1
		case LineSnapshotTypeContracting:
			// This line was removed; it has no origin.
			return -1
		default:
			// unchanged — trace through to the original position
			current = change.origin
		}
	}

	return current
}

// snapshot - creates a snapshot of the current state of `Instance`
// and appends it to the history. When the snapshot type is `LineSnapshotTypeChange`,
// the `changes` parameter should contain the computed diff; for other types it may be nil.
func (h *History) snapshot(instance *Instance, _type string, changes *map[int]LineChange) error {

	// Cannot have more than one initial snapshot in the history.
	//
	if _type == LineSnapshotTypeInitial && h.hasInitialSnapshot {
		return errors.New("initial snapshot already exists in history")
	}

	h.items = append(h.items, LinesSnapshot{
		_type:   _type,
		hash:    h.snapshotHashGenerate(instance.value),
		source:  instance.value,
		lines:   strings.Split(instance.value, "\n"),
		changes: changes,
	})

	if _type == LineSnapshotTypeInitial {
		h.hasInitialSnapshot = true
	}

	return nil
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
