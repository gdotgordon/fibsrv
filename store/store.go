// Package store defines the storage interface Store and contains any
// implementations of that interface.  These are the storage-related
// methods to support the "business logic" in the service package.
// The storage methods do things such as fetching or storing a
// memoized value, as well as counting the number of memos compared
// to a target value.
//
// The key implmentation is the PostgresStore.  There is also a
// hash map-based store used as a sort of mock for unit tests.
// The hash map store weas also used to bootstrap the service
// code logic before attemtping the Postgress store.
package store

import "context"

// FibPair is a fibonacci offset number and it value
type FibPair struct {
	Num   int    `db:"num"`
	Value uint64 `db:"value"`
}

// Store is the wrapper around a store such as a database
type Store interface {

	// Memo tries to fetch a memoized value.  The bool
	// return is false if the entry does not exist.
	Memo(context.Context, int) (uint64, bool, error)

	// Memoize stores a fibonacci number/value pair.
	Memoize(context.Context, int, uint64) error

	// MemoCount finds the number of memoized values whose value
	// is less than or equal to a target value.
	MemoCount(context.Context, uint64) (int, error)

	// Clear clears all rows from the store.
	Clear(context.Context) error
}
