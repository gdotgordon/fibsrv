package store

import (
	"context"
	"sync"
)

// Compile time interface implementation check.
var _ Store = (*MapStore)(nil)

// MapStore is a hash map implementation of the store
type MapStore struct {
	tab map[int]uint64
	mu  sync.Mutex
}

// NewMap returns a new hash map store
func NewMap() Store {
	return &MapStore{tab: make(map[int]uint64)}
}

// Memo gets a memoized fibonacci value
func (ms *MapStore) Memo(ctx context.Context, n int) (uint64, bool, error) {
	ms.mu.Lock()
	v, ok := ms.tab[n]
	ms.mu.Unlock()
	return v, ok, nil
}

// Memoize stores a memoized value
func (ms *MapStore) Memoize(ctx context.Context, n int, val uint64) error {
	ms.mu.Lock()
	ms.tab[n] = val
	ms.mu.Unlock()
	return nil
}

// MemoCount returns the number of memoizations whose value is less than or
// equal to the target.
func (ms *MapStore) MemoCount(ctx context.Context, target uint64) (int, error) {
	cnt := 0
	for _, v := range ms.tab {
		if v < target {
			cnt++
		}
	}
	return cnt, nil
}

// Clear clears the map
func (ms *MapStore) Clear(ctx context.Context) error {
	ms.mu.Lock()
	ms.tab = make(map[int]uint64)
	ms.mu.Unlock()
	return nil
}
