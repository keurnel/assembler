package lineMap

import (
	"strings"
)

// Instance - represents a singular instance of a line map. If an Instance
// exists, it is guaranteed to hold a valid source, an initial snapshot, and
// be ready for Update() calls.
type Instance struct {
	value string

	// Child structs.
	//
	source  Source
	history History
}

// New - creates a new instance of a line map. It performs the initial
// indexing (first snapshot) and returns a ready-to-use *Instance.
// The provided Source must have been obtained from LoadSource().
func New(value string, source Source) *Instance {
	instance := &Instance{
		value:   value,
		source:  source,
		history: History{},
	}

	// Perform initial indexing as part of construction so that the
	// returned Instance is always fully initialised.
	instance.history.snapshotInitial(instance)

	return instance
}

// Update - updates the value of `Instance.value` and creates a snapshot of the new state in `Instance.history`.
// Update is infallible — the initial snapshot is guaranteed to exist (FR-2), so latest() always returns
// a valid snapshot and change detection always has data to work with.
func (i *Instance) Update(newValue string) {

	// Get latest snapshot from the instance history.
	//
	latestSnapshot := i.history.latest()

	// Are there changes between the new value and the latest snapshot in the history? If not, we place
	// a snapshot in the history that indicates that there are no changes at this point in time.
	//
	if latestSnapshot.SourceCompare(newValue) {
		i.history.snapshotNoChange(i)
		return
	}

	// Collect changes between the new value and the last snapshot in the history.
	//
	changes, removals := i.changes(newValue)

	i.value = strings.Clone(newValue)

	i.history.snapshotChange(i, changes, removals)
}

// changes - computes line-level changes between the current latest snapshot and a new value.
// It uses a longest common subsequence (LCS) approach to identify which lines are unchanged,
// which were expanded (added/replaced), and which were contracted (removed).
//
// Returns:
//   - changes: map keyed by new-version line index (unchanged + expanding entries).
//   - removals: slice of contracting entries (removed lines, no position in the new version).
//
// This method is only called from Update(), which is only callable on a fully constructed
// Instance — so the history is guaranteed non-empty and latest() is guaranteed non-nil.
func (i *Instance) changes(newValue string) (map[int]LineChange, []LineChange) {

	lastSnapshot := i.history.latest()
	currentLines := lastSnapshot.lines
	newLines := strings.Split(newValue, "\n")

	changes := make(map[int]LineChange)
	var removals []LineChange

	// Compute LCS table for the two line slices.
	//
	lcs := computeLCS(currentLines, newLines)

	// Walk through the LCS to build the change map.
	// `ci` tracks position in currentLines, `ni` tracks position in newLines.
	//
	ci, ni := 0, 0
	for _, commonLine := range lcs {
		// Lines removed from current (contracting): lines in current before this common line.
		for ci < len(currentLines) && currentLines[ci] != commonLine {
			removals = append(removals, newContractingChange(ci, currentLines[ci]))
			ci++
		}

		// Lines added in new (expanding): lines in new before this common line.
		for ni < len(newLines) && newLines[ni] != commonLine {
			originLine := ci
			if originLine >= len(currentLines) {
				originLine = len(currentLines) - 1
			}
			if originLine < 0 {
				originLine = 0
			}
			changes[ni] = newExpandingChange(originLine, ni, newLines[ni])
			ni++
		}

		// This line is unchanged — record the mapping so we can trace origin.
		changes[ni] = newUnchangedChange(ci, ni, commonLine)

		ci++
		ni++
	}

	// Remaining lines in current that were removed (contracting).
	for ci < len(currentLines) {
		removals = append(removals, newContractingChange(ci, currentLines[ci]))
		ci++
	}

	// Remaining lines in new that were added (expanding).
	for ni < len(newLines) {
		originLine := len(currentLines) - 1
		if originLine < 0 {
			originLine = 0
		}
		changes[ni] = newExpandingChange(originLine, ni, newLines[ni])
		ni++
	}

	return changes, removals
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

// Lines - returns the lines from the latest snapshot. There is no nil guard —
// the latest snapshot is guaranteed to exist (FR-2).
func (i *Instance) Lines() []string {
	return i.history.latest().lines
}

// LatestSnapshot - returns the most recent snapshot in the history. There is no
// nil return — the latest snapshot is guaranteed to exist (FR-2).
func (i *Instance) LatestSnapshot() *LinesSnapshot {
	return i.history.latest()
}

// LineHistory - returns how a line evolved across all change snapshots, in chronological
// order (oldest change first, most recent last). Each entry carries full detail (type,
// origin, newIndex, content).
//
// The changes map only contains "unchanged" and "expanding" entries (FR-5.6).
// For expanding entries, tracing stops — the line did not exist before that snapshot.
func (i *Instance) LineHistory(lineNumber int) []LineChange {
	var history []LineChange
	currentLine := lineNumber

	// Walk backwards through change snapshots (skip the initial at index 0).
	for j := len(i.history.items) - 1; j > 0; j-- {
		snapshot := i.history.items[j]
		if snapshot.changes == nil {
			continue
		}

		change, exists := (*snapshot.changes)[currentLine]
		if !exists {
			// Line was not part of any change, it maps 1:1.
			// Resolve content from the snapshot lines.
			content := snapshot.lines[currentLine]
			history = append(history, newUnchangedChange(currentLine, currentLine, content))
			continue
		}

		history = append(history, change)

		if change.Type() == "expanding" {
			// This line was inserted at this snapshot — no earlier origin to trace.
			break
		}

		// unchanged — continue tracing with the origin index.
		currentLine = change.Origin()
	}

	// Reverse to produce chronological order (oldest first).
	for left, right := 0, len(history)-1; left < right; left, right = left+1, right-1 {
		history[left], history[right] = history[right], history[left]
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
