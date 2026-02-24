package lineMap

import (
	"errors"
	"strings"
)

const (
	InstanceStateInitial int8 = iota
	InstanceStateIndexed
)

// Instance - represents a singular instance of a line map.
type Instance struct {
	// Instance related data.
	//
	state int8
	value string

	// Child structs.
	//
	source  Source
	history History
}

// New - creates a new instance of a line map.
func New(value string, source Source) (*Instance, error) {

	err := source.Load()
	if err != nil {
		return nil, err
	}

	instance := Instance{
		state:   InstanceStateInitial,
		value:   value,
		source:  source,
		history: History{},
	}

	return &instance, nil
}

// InitialIndex - perform initial indexing of the lines in the `Instance.value` and
// stores the line map in the `Instance.history`. This method only executes once when
// the `Instance.history` is empty.
func (i *Instance) InitialIndex() error {
	// Does the history already have an initial snapshot? If so,
	// we return an error.
	//
	if i.history.hasInitialSnapshot {
		return errors.New("line map: initial snapshot already exists in history")
	}

	// Trigger snapshot of the initial `Instance` state.
	//
	err := i.history.snapshot(i, LineSnapshotTypeInitial, nil)
	if err != nil {
		return err
	}

	return nil
}

// Update - updates the value of `Instance.value` and creates a snapshot of the new state in `Instance.history`.
func (i *Instance) Update(newValue string) error {

	// Before we can make an update, we need to ensure that the `Instance.history` has an
	// initial snapshot. If not, we return an error.
	//
	if !i.history.hasInitialSnapshot {
		return errors.New("line map: initial snapshot does not exist in history")
	}

	// Get latest snapshot from the instance history.
	//
	latestSnapshot := i.history.latest()

	// Are there changes between the new value and the latest snapshot in the history? If not, we place
	// a snapshot in the history that indicates that there are no changes at this point in time.
	//
	if latestSnapshot.SourceCompare(newValue) {
		err := i.history.snapshot(i, LineSnapshotTypeNoChange, nil)
		if err != nil {
			return err
		}

		return nil
	}

	// Collect changes between the new value and the last snapshot in the history.
	//
	changes, err := i.changes(newValue)
	if err != nil {
		return err
	}

	i.value = strings.Clone(newValue)

	err = i.history.snapshot(i, LineSnapshotTypeChange, &changes)
	if err != nil {
		return err
	}

	return nil
}

// changes - computes line-level changes between the current latest snapshot and a new value.
// It uses a longest common subsequence (LCS) approach to identify which lines are unchanged,
// which were expanded (added/replaced), and which were contracted (removed).
func (i *Instance) changes(newValue string) (map[int]LineChange, error) {

	if i.history.empty() {
		return nil, errors.New("line map: history is empty, cannot compute changes")
	}

	lastSnapshot := i.history.latest()
	if lastSnapshot == nil {
		return nil, errors.New("line map: no snapshot available")
	}
	currentLines := lastSnapshot.lines
	newLines := strings.Split(newValue, "\n")

	changes := make(map[int]LineChange)

	// Compute LCS table for the two line slices.
	//
	lcs := computeLCS(currentLines, newLines)

	// Walk through the LCS to build the change map.
	// `ci` tracks position in currentLines, `ni` tracks position in newLines.
	//
	ci, ni := 0, 0
	for _, commonLine := range lcs {
		// Find the next occurrence of commonLine in both slices.
		//
		// Lines removed from current (contracting): lines in current before this common line.
		for ci < len(currentLines) && currentLines[ci] != commonLine {
			change, _ := newLineChange("contracting", ci, ci, ci)
			change.contractingLines = []string{currentLines[ci]}
			changes[ci] = *change
			ci++
		}

		// Lines added in new (expanding): lines in new before this common line.
		for ni < len(newLines) && newLines[ni] != commonLine {
			// These new lines map back to the current position in the original.
			originLine := ci
			if originLine >= len(currentLines) {
				originLine = len(currentLines) - 1
			}
			if originLine < 0 {
				originLine = 0
			}
			change, _ := newLineChange("expanding", originLine, ni, ni)
			change.expandingLines = []string{newLines[ni]}
			changes[ni] = *change
			ni++
		}

		// This line is unchanged — record the mapping so we can trace origin.
		change, _ := newLineChange("unchanged", ci, ni, ni)
		changes[ni] = *change

		ci++
		ni++
	}

	// Remaining lines in current that were removed (contracting).
	for ci < len(currentLines) {
		change, _ := newLineChange("contracting", ci, ci, ci)
		change.contractingLines = []string{currentLines[ci]}
		changes[ci] = *change
		ci++
	}

	// Remaining lines in new that were added (expanding).
	for ni < len(newLines) {
		originLine := len(currentLines) - 1
		if originLine < 0 {
			originLine = 0
		}
		change, _ := newLineChange("expanding", originLine, ni, ni)
		change.expandingLines = []string{newLines[ni]}
		changes[ni] = *change
		ni++
	}

	return changes, nil
}

// computeLCS - computes the longest common subsequence (LCS) of two string slices
// using dynamic programming.
//
// The LCS is the longest sequence of lines that appear in the same order in both
// slices, but not necessarily contiguously. This is used to determine which lines
// are unchanged between two versions of source code.
//
// Time complexity:  O(m × n) where m = len(a), n = len(b)
// Space complexity: O(m × n) for the DP table
//
// Example:
//
//	a = ["foo", "bar", "baz"]
//	b = ["foo", "qux", "baz"]
//	→ LCS = ["foo", "baz"]
func computeLCS(a, b []string) []string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return nil
	}

	// Build a (m+1) × (n+1) DP table where dp[i][j] represents the length
	// of the LCS for the first i elements of `a` and the first j elements of `b`.
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// Fill the DP table bottom-up.
	// Row 0 and column 0 remain 0 (base case: empty subsequence).
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				// Lines match — extend the LCS by 1.
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				// Best LCS comes from skipping the current line in `a`.
				dp[i][j] = dp[i-1][j]
			} else {
				// Best LCS comes from skipping the current line in `b`.
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack through the DP table from dp[m][n] to reconstruct
	// the actual LCS lines in order.
	lcsLen := dp[m][n]
	result := make([]string, lcsLen)
	i, j := m, n
	idx := lcsLen - 1 // Fill result slice from the end.
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			// Both lines match — this line is part of the LCS.
			result[idx] = a[i-1]
			idx--
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			// Move up in the table (skip line in `a`).
			i--
		} else {
			// Move left in the table (skip line in `b`).
			j--
		}
	}

	return result
}

// LineOrigin - traces a line number in the current (latest) processed source
// back to its original line number in the initial source.
// Returns -1 if the line cannot be traced (e.g. it was inserted during preprocessing).
func (i *Instance) LineOrigin(lineNumber int) int {
	return i.history.LineOrigin(lineNumber)
}

// SnapshotCount - returns the number of snapshots stored in the history.
func (i *Instance) SnapshotCount() int {
	return len(i.history.items)
}

// Value - returns the current value of the instance.
func (i *Instance) Value() string {
	return i.value
}

// Lines - returns the lines from the latest snapshot.
func (i *Instance) Lines() []string {
	latest := i.history.latest()
	if latest == nil {
		return nil
	}
	return latest.lines
}

// LatestSnapshot - returns the latest snapshot in the history.
// Returns nil if history is empty.
func (i *Instance) LatestSnapshot() *LinesSnapshot {
	return i.history.latest()
}

// LineHistory - returns how a line is evolved over the snapshots in the history. It returns a slice of
// LineChange structs that represent the state of the line in each snapshot (e.g. unchanged, expanded, contracted).
func (i *Instance) LineHistory(lineNumber int) []LineChange {
	var history []LineChange
	currentLine := lineNumber

	// Walk backwards through snapshots to trace the line history.
	for j := len(i.history.items) - 1; j >= 0; j-- {
		snapshot := i.history.items[j]
		if snapshot.changes == nil {
			continue
		}

		change, exists := (*snapshot.changes)[currentLine]
		if !exists {
			// Line was not part of any change, it maps 1:1.
			history = append(history, LineChange{
				_type:  "unchanged",
				origin: currentLine,
			})
			continue
		}

		history = append(history, change)

		if change._type == "contracting" {
			currentLine = change.origin
		} else if change._type == "expanding" {
			currentLine = change.origin
		}
	}

	return history
}

func (i *Instance) PrintHistory() {
	for idx, snapshot := range i.history.items {
		println("Snapshot", idx, "Type:", snapshot._type, "Hash:", snapshot.hash)
		println("Source:", snapshot.source)
		println("Lines:")
		for lineNum, line := range snapshot.lines {
			println(lineNum, ":", line)
		}

		println()
	}
}
