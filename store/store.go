package store

// Store is the wrapper around a store such as a database
type Store interface {
	Memo(int) (uint64, bool, error)
	Memoize(int, uint64) error
	FindLessEqual(uint64) (int, uint64, error)
	Clear() error
}
