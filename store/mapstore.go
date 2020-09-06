package store

import (
	"context"
	"sync"
)

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

// FindLessEqual finds the highest n and value memoized value less
// than or equal to the target
func (ms *MapStore) FindLessEqual(ctx context.Context, target uint64) (*FibPair, error) {
	max := uint64(0)
	n := 0
	for k, v := range ms.tab {
		if v > max && v <= target {
			if v == target {
				return &FibPair{k, v}, nil
			}
			max = v
			n = k
		}
	}
	return &FibPair{n, max}, nil
}

// MemoCount returns the number of memoizations whose value is less than or
// equal to the target.
func (ms *MapStore) MemoCount(context.Context, uint64) (int, error) {
	return len(ms.tab), nil
}

// Clear clears the map
func (ms *MapStore) Clear(ctx context.Context) error {
	ms.mu.Lock()
	ms.tab = make(map[int]uint64)
	ms.mu.Unlock()
	return nil
}
