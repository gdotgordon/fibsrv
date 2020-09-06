package store

import "context"

// FibPair is a fibonacci offset number and it value
type FibPair struct {
	Num   int    `db:"num"`
	Value uint64 `db:"value"`
}

// Store is the wrapper around a store such as a database
type Store interface {
	Memo(context.Context, int) (uint64, bool, error)
	Memoize(context.Context, int, uint64) error
	FindLessEqual(context.Context, uint64) (*FibPair, error)
	MemoCount(context.Context, uint64) (int, error)
	Clear(context.Context) error
}
