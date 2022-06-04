// Package sm implements a finite state machine (FSM).
//
// In the context of Raft, an FSM is state that is created
// by applying a series of commands from the logs. The FSM
// in this package is essentially the task database.
package sm

// A KVIterator iterates over a collection of key/value pairs.
type KVIterator interface {
	// ForEach applies the visit function to each pair.
	ForEach(visit func(k, v []byte) error) error

	// Close frees any allocated resources.
	Close() error
}

// A KVStore is a collection of key/value pairs.
type KVStore interface {
	// Write adds a key/value pair to the collection.
	Write(k, v []byte) error

	// DeleteAll removes all pairs from the collection.
	DeleteAll() error

	// Iterator supports iteration over the collection.
	Iterator() KVIterator
}
