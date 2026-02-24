package lineMap

import (
	"errors"
	"strings"
	"sync"
)

const (
	InstanceStateInitial int8 = iota
	InstanceState
)

// Instance - represents a singular instance of a line map.
type Instance struct {
	// Instance related data.
	//
	state      int8
	value      string
	valueMutex *sync.Mutex

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
		state:      InstanceStateInitial,
		value:      value,
		valueMutex: &sync.Mutex{},
		source:     source,
		history:    History{},
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
	err := i.history.snapshot(i, LineSnapshotTypeInitial)
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
	latestSnapshot := i.history.items[len(i.history.items)-1]

	// Are there changes between the new value and the latest snapshot in the history? If not, we place
	// a snapshot in the history that indicates that there are no changes at this point in time.
	//
	if latestSnapshot.SourceCompare(newValue) {
		err := i.history.snapshot(i, LineSnapshotTypeNoChange)
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

	println(len(changes))

	i.value = strings.Clone(newValue)

	i.history.snapshot(i, LineSnapshotTypeChange)

	return nil
}

// changes - returns changes between a new value and the last snapshot in the history.
func (i *Instance) changes(newValue string) (map[int]LineChange, error) {

	if i.history.empty() {
		return nil, errors.New("line map: history is empty, cannot compute changes")
	}

	lastSnapshot := i.history.items[len(i.history.items)-1]

	changes := make(map[int]LineChange)
	currentValue := lastSnapshot.source

	// Compute changes between the current value and the new value.
	//
	currentLines := strings.Split(currentValue, "\n")
	newLines := strings.Split(newValue, "\n")

	println(len(currentLines), len(newLines))

	return changes, nil
}
